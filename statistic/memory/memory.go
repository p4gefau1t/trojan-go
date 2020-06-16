package memory

import (
	"context"
	"github.com/p4gefau1t/trojan-go/log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/p4gefau1t/trojan-go/config"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/statistic"
	"golang.org/x/time/rate"
)

const Name = "MEMORY"

type User struct {
	sent        uint64
	recv        uint64
	lastSent    uint64
	lastRecv    uint64
	speedLock   sync.Mutex
	sendSpeed   uint64
	recvSpeed   uint64
	hash        string
	ipTableLock sync.Mutex
	ipTable     map[string]struct{}
	maxIPNum    int
	sendLimiter *rate.Limiter
	recvLimiter *rate.Limiter
	ctx         context.Context
	cancel      context.CancelFunc
}

func (u *User) Close() error {
	u.ResetTraffic()
	u.cancel()
	return nil
}

func (u *User) AddIP(ip string) bool {
	if u.maxIPNum <= 0 {
		return true
	}
	u.ipTableLock.Lock()
	defer u.ipTableLock.Unlock()
	_, found := u.ipTable[ip]
	if found {
		return true
	}
	if len(u.ipTable)+1 > u.maxIPNum {
		return false
	}
	u.ipTable[ip] = struct{}{}
	return true
}

func (u *User) DelIP(ip string) bool {
	if u.maxIPNum <= 0 {
		return true
	}
	u.ipTableLock.Lock()
	defer u.ipTableLock.Unlock()
	_, found := u.ipTable[ip]
	if !found {
		return false
	}
	delete(u.ipTable, ip)
	return true
}

func (u *User) GetIP() int {
	u.ipTableLock.Lock()
	defer u.ipTableLock.Unlock()
	return len(u.ipTable)
}

func (u *User) SetIPLimit(n int) {
	u.maxIPNum = n
}

func (u *User) GetIPLimit() int {
	return u.maxIPNum
}

func (u *User) AddTraffic(sent, recv int) {
	if u.sendLimiter != nil && sent != 0 {
		u.sendLimiter.WaitN(u.ctx, sent)
	} else if u.recvLimiter != nil && recv != 0 {
		u.recvLimiter.WaitN(u.ctx, recv)
	}
	atomic.AddUint64(&u.sent, uint64(sent))
	atomic.AddUint64(&u.recv, uint64(recv))
}

func (u *User) SetSpeedLimit(send, recv int) {
	if send <= 0 {
		u.sendLimiter = nil
	} else {
		u.sendLimiter = rate.NewLimiter(rate.Limit(send), send*2)
	}
	if recv <= 0 {
		u.recvLimiter = nil
	} else {
		u.recvLimiter = rate.NewLimiter(rate.Limit(recv), recv*2)
	}
}

func (u *User) SetTrafficTotal(send, recv uint64) {
	atomic.StoreUint64(&u.sent, send)
	atomic.StoreUint64(&u.recv, recv)
}

func (u *User) GetSpeedLimit() (send, recv int) {
	sendLimit := 0
	recvLimit := 0
	if u.sendLimiter != nil {
		sendLimit = int(u.sendLimiter.Limit())
	}
	if u.recvLimiter != nil {
		recvLimit = int(u.recvLimiter.Limit())
	}
	return sendLimit, recvLimit
}

func (u *User) Hash() string {
	return u.hash
}

func (u *User) GetTraffic() (uint64, uint64) {
	return atomic.LoadUint64(&u.sent), atomic.LoadUint64(&u.recv)
}

func (u *User) ResetTraffic() {
	atomic.StoreUint64(&u.sent, 0)
	atomic.StoreUint64(&u.recv, 0)
	atomic.StoreUint64(&u.lastSent, 0)
	atomic.StoreUint64(&u.lastRecv, 0)
}

func (u *User) GetAndResetTraffic() (uint64, uint64) {
	sent := atomic.SwapUint64(&u.sent, 0)
	recv := atomic.SwapUint64(&u.recv, 0)
	atomic.StoreUint64(&u.lastSent, 0)
	atomic.StoreUint64(&u.lastRecv, 0)
	return sent, recv
}

func (u *User) speedUpdater() {
	for {
		select {
		case <-u.ctx.Done():
			return
		case <-time.After(time.Second):
			u.speedLock.Lock()
			sent, recv := u.GetTraffic()
			u.sendSpeed = sent - u.lastSent
			u.recvSpeed = recv - u.lastRecv
			u.lastSent = sent
			u.lastRecv = recv
			u.speedLock.Unlock()
		}
	}
}

func (u *User) GetSpeed() (uint64, uint64) {
	u.speedLock.Lock()
	defer u.speedLock.Unlock()
	return u.sendSpeed, u.recvSpeed
}

type Authenticator struct {
	sync.RWMutex

	users map[string]*User
	ctx   context.Context
}

func (a *Authenticator) AuthUser(hash string) (bool, statistic.User) {
	a.RLock()
	defer a.RUnlock()
	if user, found := a.users[hash]; found {
		return true, user
	}
	return false, nil
}

func (a *Authenticator) AddUser(hash string) error {
	a.Lock()
	defer a.Unlock()
	if _, found := a.users[hash]; found {
		return common.NewError("Hash " + hash + " is already exist")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	meter := &User{
		hash:    hash,
		ctx:     ctx,
		cancel:  cancel,
		ipTable: make(map[string]struct{}),
	}
	go meter.speedUpdater()
	a.users[hash] = meter
	return nil
}

func (a *Authenticator) DelUser(hash string) error {
	a.Lock()
	defer a.Unlock()
	meter, found := a.users[hash]
	if !found {
		return common.NewError("hash " + hash + " not found")
	}
	meter.Close()
	delete(a.users, hash)
	return nil
}

func (a *Authenticator) ListUsers() []statistic.User {
	a.RLock()
	defer a.RUnlock()
	result := make([]statistic.User, 0, len(a.users))
	for _, u := range a.users {
		result = append(result, u)
	}
	return result
}

func (a *Authenticator) Close() error {
	return nil
}

func NewAuthenticator(ctx context.Context) (statistic.Authenticator, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	u := &Authenticator{
		ctx:   ctx,
		users: make(map[string]*User),
	}
	for _, password := range cfg.Passwords {
		hash := common.SHA224String(password)
		u.AddUser(hash)
	}
	log.Debug("memory authenticator created")
	return u, nil
}

func init() {
	statistic.RegisterAuthenticatorCreator(Name, NewAuthenticator)
}
