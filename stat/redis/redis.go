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

type RedisTrafficMeter struct {
	stat.TrafficMeter

	hash string
	db   *radix.Pool
	ctx  context.Context
}

func (m *RedisTrafficMeter) Close() error { return nil }

func (m *RedisTrafficMeter) Count(sent, recv int) {
	key := m.hash
	evalScript := radix.NewEvalScript(1, `
		if redis.call('exists', KEYS[1]) == 1
		then
			redis.call('hincrby', KEYS[1], 'upload', ARGV[1])
			redis.call('hincrby', KEYS[1], 'download', ARGV[2])
		end
	`)

	if err := m.db.Do(evalScript.Cmd(nil, key, strconv.Itoa(recv), strconv.Itoa(sent))); err != nil {
		log.Error(common.NewError("failed to update data to user").Base(err))
	}
}

func (m *RedisTrafficMeter) LimitSpeed(send, recv int) {}

func (m *RedisTrafficMeter) GetSpeedLimit() (send, recv int) { return 0, 0 }

func (m *RedisTrafficMeter) Hash() string { return m.hash }

func (m *RedisTrafficMeter) Get() (uint64, uint64) { return 0, 0 }

func (m *RedisTrafficMeter) Reset() {}

func (m *RedisTrafficMeter) GetAndReset() (uint64, uint64) { return 0, 0 }

func (m *RedisTrafficMeter) GetSpeed() (uint64, uint64) { return 0, 0 }

type RedisAuthenticator struct {
	stat.Authenticator
	db  *radix.Pool
	ctx context.Context
}

func (a *RedisAuthenticator) AuthUser(hash string) (bool, stat.TrafficMeter) {
	var exist bool
	if err := a.db.Do(radix.Cmd(&exist, "EXISTS", hash)); err != nil {
		log.Error(common.NewError("failed to check user in DB").Base(err))
	}
	if exist {
		return true, &RedisTrafficMeter{hash: hash, db: a.db, ctx: a.ctx}
	}
	return false, nil
}

func (a *RedisAuthenticator) AddUser(hash string) error { return nil }

func (a *RedisAuthenticator) DelUser(hash string) error { return nil }

func (a *RedisAuthenticator) ListUsers() []stat.TrafficMeter { return []stat.TrafficMeter{} }

func NewRedisAuth(ctx context.Context, config *conf.GlobalConfig) (stat.Authenticator, error) {
	addr := config.Redis.ServerHost + ":" + strconv.Itoa(config.Redis.ServerPort)
	conn := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialAuthPass(config.Redis.Password),
		)
	}
	db, err := radix.NewPool("tcp", addr, 10, radix.PoolConnFunc(conn))
	if err != nil {
		return nil, common.NewError("failed to connect to database server").Base(err)
	}
	return &RedisAuthenticator{db: db, ctx: ctx}, nil
}

func init() {
	stat.RegisterAuthCreator("redis", NewRedisAuth)
}
