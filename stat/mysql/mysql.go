package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	// MySQL Driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
	"github.com/p4gefau1t/trojan-go/stat/memory"
)

type DBAuth struct {
	*memory.MemoryAuthenticator
	db             *sql.DB
	updateDuration time.Duration
	ctx            context.Context
}

func (a *DBAuth) updater() {
	for {
		for _, user := range a.ListUsers() {
			//swap upload and download for users
			hash := user.Hash()
			sent, recv := user.GetAndReset()

			s, err := a.db.Exec("UPDATE `users` SET `upload`=`upload`+?, `download`=`download`+? WHERE `password`=?;", recv, sent, hash)
			if err != nil {
				log.Error(common.NewError("failed to update data to user").Base(err))
				continue
			}
			if r, err := s.RowsAffected(); err != nil {
				if r == 0 {
					a.DelUser(hash)
				}
			}
		}
		log.Info("buffered data has been written into the database")

		//update memory
		rows, err := a.db.Query("SELECT password,quota,download,upload FROM users")
		if err != nil {
			log.Error(common.NewError("failed to pull data from the database").Base(err))
			time.Sleep(a.updateDuration)
			continue
		}
		for rows.Next() {
			var hash string
			var quota, download, upload int64
			err := rows.Scan(&hash, &quota, &download, &upload)
			if err != nil {
				log.Error(common.NewError("failed to obtain data from the query result").Base(err))
				break
			}
			if download+upload < quota || quota < 0 {
				a.AddUser(hash)
			} else {
				a.DelUser(hash)
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

func connectDatabase(driverName, username, password, ip string, port int, dbName string) (*sql.DB, error) {
	path := strings.Join([]string{username, ":", password, "@tcp(", ip, ":", fmt.Sprintf("%d", port), ")/", dbName, "?charset=utf8"}, "")
	return sql.Open(driverName, path)
}

func NewDBAuth(ctx context.Context, config *conf.GlobalConfig) (stat.Authenticator, error) {
	db, err := connectDatabase(
		"mysql",
		config.MySQL.Username,
		config.MySQL.Password,
		config.MySQL.ServerHost,
		config.MySQL.ServerPort,
		config.MySQL.Database,
	)
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
		updateDuration:      time.Duration(config.MySQL.CheckRate) * time.Second,
		MemoryAuthenticator: memoryAuth.(*memory.MemoryAuthenticator),
	}
	go a.updater()
	return a, nil
}

func init() {
	stat.RegisterAuthCreator("mysql", NewDBAuth)
}
