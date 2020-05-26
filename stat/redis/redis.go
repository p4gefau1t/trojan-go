package redis

import (
	"context"
	"strconv"

	"github.com/mediocregopher/radix/v3"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
)

type RedisUser struct {
	hash string
	db   *radix.Pool
	ctx  context.Context
}

func (m *RedisUser) Close() error { return nil }

func (m *RedisUser) AddTraffic(sent, recv int) {
	key := m.hash
	evalScript := radix.NewEvalScript(1, `
		if redis.call('exists', KEYS[1]) == 1
		then
			redis.call('hincrby', KEYS[1], 'upload', ARGV[1])
			redis.call('hincrby', KEYS[1], 'download', ARGV[2])
		end
	`)

	if err := m.db.Do(evalScript.Cmd(nil, key, strconv.Itoa(recv), strconv.Itoa(sent))); err != nil {
		log.Error(common.NewError("Failed to update data to user").Base(err))
	}
}

// TODO implement these methods

func (m *RedisUser) SetSpeedLimit(send, recv int) {}

func (m *RedisUser) GetSpeedLimit() (send, recv int) { return 0, 0 }

func (m *RedisUser) Hash() string { return m.hash }

func (m *RedisUser) GetTraffic() (uint64, uint64) { return 0, 0 }

func (m *RedisUser) ResetTraffic() {}

func (m *RedisUser) GetAndResetTraffic() (uint64, uint64) { return 0, 0 }

func (m *RedisUser) GetSpeed() (uint64, uint64) { return 0, 0 }

func (m *RedisUser) AddIP(string) bool { return true }

func (m *RedisUser) DelIP(string) bool { return true }

func (m *RedisUser) SetIPLimit(int) {}

func (m *RedisUser) GetIPLimit() int { return 0 }

type RedisAuthenticator struct {
	stat.Authenticator
	db  *radix.Pool
	ctx context.Context
}

func (a *RedisAuthenticator) AuthUser(hash string) (bool, stat.User) {
	var exist bool
	if err := a.db.Do(radix.Cmd(&exist, "EXISTS", hash)); err != nil {
		log.Error(common.NewError("Failed to check user in DB").Base(err))
	}
	if exist {
		return true, &RedisUser{hash: hash, db: a.db, ctx: a.ctx}
	}
	return false, nil
}

// TODO implement these methods

func (a *RedisAuthenticator) AddUser(hash string) error { return nil }

func (a *RedisAuthenticator) DelUser(hash string) error { return nil }

func (a *RedisAuthenticator) ListUsers() []stat.User { return []stat.User{} }

func NewRedisAuth(ctx context.Context, config *conf.GlobalConfig) (stat.Authenticator, error) {
	addr := config.Redis.ServerHost + ":" + strconv.Itoa(config.Redis.ServerPort)
	conn := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialAuthPass(config.Redis.Password),
		)
	}
	db, err := radix.NewPool("tcp", addr, 10, radix.PoolConnFunc(conn))
	if err != nil {
		return nil, common.NewError("Failed to connect to database server").Base(err)
	}
	return &RedisAuthenticator{db: db, ctx: ctx}, nil
}

func init() {
	stat.RegisterAuthCreator("redis", NewRedisAuth)
}
