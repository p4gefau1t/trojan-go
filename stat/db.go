package stat

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
)

type trafficInfo struct {
	passwordHash string
	recv         int
	sent         int
}

type DBTrafficMeter struct {
	TrafficMeter
	db             *sql.DB
	trafficChan    chan *trafficInfo
	ctx            context.Context
	cancel         context.CancelFunc
	updateDuration time.Duration
}

func (c *DBTrafficMeter) Count(passwordHash string, sent int, recv int) {
	c.trafficChan <- &trafficInfo{
		passwordHash: passwordHash,
		sent:         sent,
		recv:         recv,
	}
}

func (c *DBTrafficMeter) Close() error {
	c.cancel()
	return c.db.Close()
}

func (c *DBTrafficMeter) dbDaemon() {
	for {
		beginTime := time.Now()
		statBuffer := make(map[string]*trafficInfo)
		for {
			select {
			case u := <-c.trafficChan:
				t, found := statBuffer[u.passwordHash]
				if !found {
					t = &trafficInfo{
						passwordHash: u.passwordHash,
					}
					statBuffer[u.passwordHash] = t
				}
				t.sent += u.sent
				t.recv += u.recv
			case <-time.After(c.updateDuration):
				break
			case <-c.ctx.Done():
				return
			}
			if time.Now().Sub(beginTime) > c.updateDuration {
				break
			}
		}
		if len(statBuffer) == 0 {
			continue
		}
		tx, err := c.db.Begin()
		if err != nil {
			log.Error(common.NewError("cannot begin transactin").Base(err))
			continue
		}
		for _, traffic := range statBuffer {
			//swap upload and download for users
			s, err := tx.Prepare("UPDATE users SET upload=upload+? WHERE password=?;")
			common.Must(err)
			_, err = s.Exec(traffic.recv, traffic.passwordHash)

			s, err = tx.Prepare("UPDATE users SET download=download+? WHERE password=?;")
			common.Must(err)
			_, err = s.Exec(traffic.sent, traffic.passwordHash)

			if err != nil {
				log.Error(common.NewError("failed to update data to tx").Base(err))
				break
			}
		}
		err = tx.Commit()
		if err != nil {
			log.Error(common.NewError("failed to commit tx").Base(err))
		} else {
			log.Info("buffered data has been written into the database")
		}
	}
}

func NewDBTrafficMeter(config *conf.GlobalConfig, db *sql.DB) (TrafficMeter, error) {
	c := &DBTrafficMeter{
		db:             db,
		trafficChan:    make(chan *trafficInfo, 1024),
		ctx:            context.Background(),
		updateDuration: time.Duration(config.MySQL.CheckRate) * time.Second,
	}
	go c.dbDaemon()
	return c, nil
}

type userInfo struct {
	username     string
	passwordHash string
	download     uint64
	upload       uint64
	quota        uint64
}

type DBAuthenticator struct {
	db             *sql.DB
	validUsers     sync.Map
	ctx            context.Context
	cancel         context.CancelFunc
	updateDuration time.Duration
	Authenticator
}

func (a *DBAuthenticator) CheckHash(hash string) bool {
	_, ok := a.validUsers.Load(hash)
	if !ok {
		return false
	}
	return true
}

func (a *DBAuthenticator) updateDaemon() {
	for {
		rows, err := a.db.Query("SELECT password,quota,download,upload FROM users")
		if err != nil {
			log.Error(common.NewError("failed to pull data from the database").Base(err))
			time.Sleep(a.updateDuration)
			continue
		}
		newValidUsers := make(map[string]string)
		for rows.Next() {
			var passwordHash string
			var quota, download, upload int64
			err := rows.Scan(&passwordHash, &quota, &download, &upload)
			if err != nil {
				log.Error(common.NewError("failed to obtain data from the query result").Base(err))
				break
			}
			if download+upload < quota || quota < 0 {
				newValidUsers[passwordHash] = "valid"
			}
		}
		//delete those out of quota
		a.validUsers.Range(func(key interface{}, val interface{}) bool {
			if _, found := newValidUsers[key.(string)]; !found {
				a.validUsers.Delete(key)
			}
			return true
		})
		for k, v := range newValidUsers {
			a.validUsers.Store(k, v)
		}
		select {
		case <-time.After(a.updateDuration):
			break
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *DBAuthenticator) Close() error {
	a.cancel()
	return a.db.Close()
}

func NewDBAuthenticator(config *conf.GlobalConfig, db *sql.DB) (Authenticator, error) {
	ctx, cancel := context.WithCancel(context.Background())
	a := &DBAuthenticator{
		db:             db,
		cancel:         cancel,
		ctx:            ctx,
		updateDuration: time.Duration(config.MySQL.CheckRate) * time.Second,
	}
	go a.updateDaemon()
	return a, nil
}
