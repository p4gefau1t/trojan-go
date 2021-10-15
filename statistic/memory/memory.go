package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/statistic/sqlite"
)

const Name = "MEMORY"

type User struct {
	// WARNING: do not change the order of these fields.
	// 64-bit fields that use `sync/atomic` package functions
	// must be 64-bit aligned on 32-bit systems.
	// Reference: https://github.com/golang/go/issues/599
	// Solution: https://github.com/golang/go/issues/11891#issuecomment-433623786
	Sent        uint64
	Recv        uint64
	lastSent    uint64
	lastRecv    uint64
	sendSpeed   uint64
	recvSpeed   uint64
	Hash        string
	ipTable     sync.Map
	ipNum       int32
	MaxIPNum    int
	limiterLock sync.RWMutex
	SendLimiter *rate.Limiter
	RecvLimiter *rate.Limiter
	ctx         context.Context
	cancel      context.CancelFunc
}

func (u *User) Close() error {
	u.ResetTraffic()
	u.cancel()
	return nil
}

func (u *User) AddIP(ip string) bool {
	if u.MaxIPNum <= 0 {
		return true
	}
	_, found := u.ipTable.Load(ip)
	if found {
		return true
	}
	if int(u.ipNum)+1 > u.MaxIPNum {
		return false
	}
	u.ipTable.Store(ip, true)
	atomic.AddInt32(&u.ipNum, 1)
	return true
}

func (u *User) DelIP(ip string) bool {
	if u.MaxIPNum <= 0 {
		return true
	}
	_, found := u.ipTable.Load(ip)
	if !found {
		return false
	}
	u.ipTable.Delete(ip)
	atomic.AddInt32(&u.ipNum, -1)
	return true
}

func (u *User) GetIP() int {
	return int(u.ipNum)
}

func (u *User) setIPLimit(n int) {
	u.MaxIPNum = n
}

func (u *User) GetIPLimit() int {
	return u.MaxIPNum
}

func (u *User) AddSentTraffic(sent int) {
	u.limiterLock.RLock()
	if u.SendLimiter != nil && sent >= 0 {
		u.SendLimiter.WaitN(u.ctx, sent)
	}
	u.limiterLock.RUnlock()
	atomic.AddUint64(&u.Sent, uint64(sent))
}

func (u *User) AddRecvTraffic(recv int) {
	u.limiterLock.RLock()
	if u.RecvLimiter != nil && recv >= 0 {
		u.RecvLimiter.WaitN(u.ctx, recv)
	}
	u.limiterLock.RUnlock()
	atomic.AddUint64(&u.Recv, uint64(recv))
}

func (u *User) SetSpeedLimit(send, recv int) {
	u.limiterLock.Lock()
	defer u.limiterLock.Unlock()

	if send <= 0 {
		u.SendLimiter = nil
	} else {
		u.SendLimiter = rate.NewLimiter(rate.Limit(send), send*2)
	}
	if recv <= 0 {
		u.RecvLimiter = nil
	} else {
		u.RecvLimiter = rate.NewLimiter(rate.Limit(recv), recv*2)
	}
}

func (u *User) GetSpeedLimit() (send, recv int) {
	u.limiterLock.RLock()
	defer u.limiterLock.RUnlock()

	if u.SendLimiter != nil {
		send = int(u.SendLimiter.Limit())
	}
	if u.RecvLimiter != nil {
		recv = int(u.RecvLimiter.Limit())
	}
	return
}

func (u *User) GetHash() string {
	return u.Hash
}

func (u *User) setTraffic(send, recv uint64) {
	atomic.StoreUint64(&u.Sent, send)
	atomic.StoreUint64(&u.Recv, recv)
}

func (u *User) GetTraffic() (uint64, uint64) {
	return atomic.LoadUint64(&u.Sent), atomic.LoadUint64(&u.Recv)
}

func (u *User) ResetTraffic() (uint64, uint64) {
	sent := atomic.SwapUint64(&u.Sent, 0)
	recv := atomic.SwapUint64(&u.Recv, 0)
	atomic.StoreUint64(&u.lastSent, 0)
	atomic.StoreUint64(&u.lastRecv, 0)
	return sent, recv
}

func (u *User) speedUpdater() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-u.ctx.Done():
			return
		case <-ticker.C:
			sent, recv := u.GetTraffic()
			atomic.StoreUint64(&u.sendSpeed, sent-u.lastSent)
			atomic.StoreUint64(&u.recvSpeed, recv-u.lastRecv)
			atomic.StoreUint64(&u.lastSent, sent)
			atomic.StoreUint64(&u.lastRecv, recv)
		}
	}
}

func (u *User) trafficUpdater(pst statistic.Persistencer) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-u.ctx.Done():
			return
		case <-ticker.C:
			if pst != nil {
				sent, recv := u.GetTraffic()
				log.Debugf("Update %s traffic", u.Hash)
				err := pst.UpdateUserTraffic(u.Hash, sent, recv)
				if err != nil {
					log.Debugf("Update user %s traffic failed: %s", u.Hash, err)
				}
			}
		}
	}
}

func (u *User) GetSpeed() (uint64, uint64) {
	return atomic.LoadUint64(&u.sendSpeed), atomic.LoadUint64(&u.recvSpeed)
}

type Authenticator struct {
	users sync.Map
	pst   statistic.Persistencer
	ctx   context.Context
}

func (a *Authenticator) AuthUser(hash string) (bool, statistic.User) {
	if user, found := a.users.Load(hash); found {
		return true, user.(*User)
	}
	return false, nil
}

func (a *Authenticator) AddUser(hash string) error {
	if _, found := a.users.Load(hash); found {
		return common.NewError("hash " + hash + " is already exist")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	meter := &User{
		Hash:   hash,
		ctx:    ctx,
		cancel: cancel,
	}
	go meter.speedUpdater()
	a.users.Store(hash, meter)
	if a.pst != nil {
		go meter.trafficUpdater(a.pst)
		err := a.pst.SaveUser(meter)
		if err != nil {
			log.Errorf("Save user %s failed: %s", hash, err)
		}
	}
	return nil
}

func (a *Authenticator) DelUser(hash string) error {
	meter, found := a.users.Load(hash)
	if !found {
		return common.NewError("hash " + hash + " not found")
	}
	meter.(*User).Close()
	a.users.Delete(hash)
	if a.pst != nil {
		a.pst.DeleteUser(hash)
	}
	return nil
}

func (a *Authenticator) ListUsers() []statistic.User {
	result := make([]statistic.User, 0)
	a.users.Range(func(k, v interface{}) bool {
		result = append(result, v.(*User))
		return true
	})
	return result
}

func (a *Authenticator) Close() error {
	return nil
}

func (a *Authenticator) SetUserTraffic(hash string, sent, recv uint64) error {
	u, exist := a.users.Load(hash)
	if !exist {
		return common.NewErrorf("user %v not found", hash)
	}
	user := u.(*User)
	user.setTraffic(sent, recv)
	if a.pst != nil {
		err := a.pst.SaveUser(user)
		if err != nil {
			log.Errorf("Save user %s failed: %s", hash, err)
		}
	}
	return nil
}

func (a *Authenticator) SetUserSpeedLimit(hash string, send, recv int) error {
	u, exist := a.users.Load(hash)
	if !exist {
		return common.NewErrorf("user %v not found", hash)
	}
	user := u.(*User)
	user.SetSpeedLimit(send, recv)
	if a.pst != nil {
		err := a.pst.SaveUser(user)
		if err != nil {
			log.Errorf("Save user %s failed: %s", hash, err)
		}
	}
	return nil
}

func (a *Authenticator) SetUserIPLimit(hash string, limit int) error {
	u, exist := a.users.Load(hash)
	if !exist {
		return common.NewErrorf("user %v not found", hash)
	}
	user := u.(*User)
	user.setIPLimit(limit)
	if a.pst != nil {
		err := a.pst.SaveUser(user)
		if err != nil {
			log.Errorf("Save user %s failed: %s", hash, err)
		}
	}
	return nil
}

func NewAuthenticator(ctx context.Context) (statistic.Authenticator, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	a := &Authenticator{
		ctx: ctx,
	}
	var err error
	if cfg.Sqlite != "" {
		a.pst, err = sqlite.NewSqlitePersistencer(cfg.Sqlite)
		if err != nil {
			return nil, err
		}
	}
	if a.pst != nil {
		err := a.pst.ListUser(func(hash string, u statistic.Metadata) bool {
			if _, found := a.users.Load(hash); found {
				log.Error("hash " + hash + " is already exist")
				return true
			}
			ctx, cancel := context.WithCancel(a.ctx)
			user := &User{
				Hash:   hash,
				ctx:    ctx,
				cancel: cancel,
			}
			user.setIPLimit(u.GetIPLimit())
			user.SetSpeedLimit(u.GetSpeedLimit())
			user.setTraffic(u.GetTraffic())
			go user.speedUpdater()
			go user.trafficUpdater(a.pst)
			a.users.Store(hash, user)
			return true
		})
		if err != nil {
			log.Errorf("List user from persistencer: %s", err)
		}
	}
	for _, password := range cfg.Passwords {
		hash := common.SHA224String(password)
		a.AddUser(hash)
	}
	log.Debug("memory authenticator created")
	return a, nil
}

func init() {
	statistic.RegisterAuthenticatorCreator(Name, NewAuthenticator)
}
