package redis

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/mediocregopher/radix/v3"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
	"github.com/p4gefau1t/trojan-go/stat/memory"
)

type DBAuth struct {
	*memory.MemoryAuthenticator
	db             *radix.Pool
	updateDuration time.Duration
	ctx            context.Context
}

func validateHash(i string) bool {
	re := regexp.MustCompile("^[0-9a-fA-F]{56}$")
	return re.MatchString(i)
}

func (a *DBAuth) updater() {
	for {
		users := a.ListUsers()
		for _, user := range users {
			// fetch user flow
			hash := user.Hash()
			sent, recv := user.GetAndReset()

			// check if user exists in DB
			var exist bool
			if err := a.db.Do(radix.Cmd(&exist, "EXISTS", hash)); err != nil {
				log.Error(common.NewError("failed to check user in DB").Base(err))
			}

			// remove the user if not
			if !exist {
				a.DelUser(hash)
				continue
			}

			// update flow to DB
			pipe := radix.Pipeline(
				radix.Cmd(nil, "HINCRBY", hash, "upload", strconv.FormatUint(recv, 10)),
				radix.Cmd(nil, "HINCRBY", hash, "download", strconv.FormatUint(sent, 10)),
			)
			if err := a.db.Do(pipe); err != nil {
				log.Error(common.NewError("failed to execute pipeline").Base(err))
			}
		}
		log.Info("buffered data has been written into the database")

		//update memory
		var keys []string
		if err := a.db.Do(radix.Cmd(&keys, "KEYS", "*")); err != nil {
			log.Error(common.NewError("failed to pull data from the database").Base(err))
			time.Sleep(a.updateDuration)
			continue
		}
		for _, key := range keys {
			if validateHash(key) {
				a.AddUser(key)
			}
		}

		select {
		case <-time.After(a.updateDuration):
		case <-a.ctx.Done():
			log.Debug("db daemon exiting...")
			return
		}
	}
}

func NewDBAuth(ctx context.Context, config *conf.GlobalConfig) (stat.Authenticator, error) {
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

	memoryAuth, err := memory.NewMemoryAuth(ctx, config)
	if err != nil {
		return nil, err
	}
	a := &DBAuth{
		db:                  db,
		ctx:                 ctx,
		updateDuration:      time.Duration(config.Redis.CheckRate) * time.Second,
		MemoryAuthenticator: memoryAuth.(*memory.MemoryAuthenticator),
	}
	go a.updater()
	return a, nil
}

func init() {
	stat.RegisterAuthCreator("redis", NewDBAuth)
}
