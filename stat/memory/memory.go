package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/stat"
	"golang.org/x/time/rate"
)

type MemoryUser struct {
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

func (u *MemoryUser) Close() error {
	u.ResetTraffic()
	u.cancel()
	return nil
}

func (u *MemoryUser) AddIP(ip string) bool {
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

func (u *MemoryUser) DelIP(ip string) bool {
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

func (u *MemoryUser) SetIPLimit(n int) {
	u.maxIPNum = n
}

func (u *MemoryUser) GetIPLimit() int {
	return u.maxIPNum
}

func (u *MemoryUser) AddTraffic(sent, recv int) {
	if u.sendLimiter != nil && sent != 0 {
		u.sendLimiter.WaitN(u.ctx, sent)
	} else if u.recvLimiter != nil && recv != 0 {
		u.recvLimiter.WaitN(u.ctx, recv)
	}
	atomic.AddUint64(&u.sent, uint64(sent))
	atomic.AddUint64(&u.recv, uint64(recv))
}

func (u *MemoryUser) SetSpeedLimit(send, recv int) {
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

func (u *MemoryUser) GetSpeedLimit() (send, recv int) {
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

func (u *MemoryUser) Hash() string {
	return u.hash
}

func (u *MemoryUser) GetTraffic() (uint64, uint64) {
	return atomic.LoadUint64(&u.sent), atomic.LoadUint64(&u.recv)
}

func (u *MemoryUser) ResetTraffic() {
	atomic.StoreUint64(&u.sent, 0)
	atomic.StoreUint64(&u.recv, 0)
	atomic.StoreUint64(&u.lastSent, 0)
	atomic.StoreUint64(&u.lastRecv, 0)
}

func (u *MemoryUser) GetAndResetTraffic() (uint64, uint64) {
	sent := atomic.SwapUint64(&u.sent, 0)
	recv := atomic.SwapUint64(&u.recv, 0)
	atomic.StoreUint64(&u.lastSent, 0)
	atomic.StoreUint64(&u.lastRecv, 0)
	return sent, recv
}

func (u *MemoryUser) speedUpdater() {
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

func (m *MemoryUser) GetSpeed() (uint64, uint64) {
	m.speedLock.Lock()
	defer m.speedLock.Unlock()
	return m.sendSpeed, m.recvSpeed
}

type MemoryAuthenticator struct {
	stat.Authenticator
	sync.RWMutex

	users map[string]*MemoryUser
	ctx   context.Context
}

func (a *MemoryAuthenticator) AuthUser(hash string) (bool, stat.User) {
	a.RLock()
	defer a.RUnlock()
	if user, found := a.users[hash]; found {
		return true, user
	}
	return false, nil
}

func (a *MemoryAuthenticator) AddUser(hash string) error {
	a.Lock()
	defer a.Unlock()
	if _, found := a.users[hash]; found {
		return common.NewError("Hash " + hash + " is already exist")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	meter := &MemoryUser{
		hash:    hash,
		ctx:     ctx,
		cancel:  cancel,
		ipTable: make(map[string]struct{}),
	}
	go meter.speedUpdater()
	a.users[hash] = meter
	return nil
}

func (a *MemoryAuthenticator) DelUser(hash string) error {
	a.Lock()
	defer a.Unlock()
	meter, found := a.users[hash]
	if !found {
		return common.NewError("Hash " + hash + "is not exist")
	}
	meter.Close()
	delete(a.users, hash)
	return nil
}

func (a *MemoryAuthenticator) ListUsers() []stat.User {
	a.RLock()
	defer a.RUnlock()
	result := make([]stat.User, 0, len(a.users))
	for _, u := range a.users {
		result = append(result, u)
	}
	return result
}

func NewMemoryAuth(ctx context.Context, config *conf.GlobalConfig) (stat.Authenticator, error) {
	u := &MemoryAuthenticator{
		ctx:   ctx,
		users: make(map[string]*MemoryUser),
	}
	for hash := range config.Hash {
		u.AddUser(hash)
	}
	return u, nil
}

func init() {
	stat.RegisterAuthCreator("memory", NewMemoryAuth)
}
