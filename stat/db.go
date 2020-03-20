package stat

import (
	"context"
	"database/sql"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
)

type traffic struct {
	passwordHash string
	download     int
	upload       int
}

type DBTrafficCounter struct {
	TrafficCounter
	db          *sql.DB
	trafficChan chan *traffic
	ctx         context.Context
	cancel      context.CancelFunc
}

const (
	statsUpdateDuration = time.Second * 5
)

func (c *DBTrafficCounter) Count(passwordHash string, upload int, download int) {
	c.trafficChan <- &traffic{
		passwordHash: passwordHash,
		upload:       upload,
		download:     download,
	}
}

func (c *DBTrafficCounter) Close() error {
	c.cancel()
	return c.db.Close()
}

func (c *DBTrafficCounter) dbDaemon() {
	for {
		beginTime := time.Now()
		statBuffer := make(map[string]*traffic)
		for {
			select {
			case u := <-c.trafficChan:
				t, found := statBuffer[u.passwordHash]
				if !found {
					t = &traffic{
						passwordHash: u.passwordHash,
					}
					statBuffer[u.passwordHash] = t
				}
				t.upload += u.upload
				t.download += u.download
			case <-time.After(statsUpdateDuration):
				break
			case <-c.ctx.Done():
				return
			}
			if time.Now().Sub(beginTime) > statsUpdateDuration {
				break
			}
		}
		if len(statBuffer) == 0 {
			continue
		}
		tx, err := c.db.Begin()
		if err != nil {
			logger.Error(common.NewError("cannot begin transactin").Base(err))
			continue
		}
		for _, traffic := range statBuffer {
			s, err := tx.Prepare("UPDATE users SET upload=upload+? WHERE password=?;")
			common.Must(err)
			_, err = s.Exec(traffic.upload, traffic.passwordHash)

			s, err = tx.Prepare("UPDATE users SET download=download+? WHERE password=?;")
			common.Must(err)
			_, err = s.Exec(traffic.download, traffic.passwordHash)

			if err != nil {
				logger.Error(common.NewError("failed to update data to tx").Base(err))
				break
			}
		}
		err = tx.Commit()
		if err != nil {
			logger.Error(common.NewError("failed to commit tx").Base(err))
		}
		logger.Info("buffered data has been written into the database")
	}
}

func NewDBTrafficCounter(db *sql.DB) (TrafficCounter, error) {
	db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    password CHAR(56) NOT NULL,
    quota BIGINT NOT NULL DEFAULT 0,
    download BIGINT UNSIGNED NOT NULL DEFAULT 0,
    upload BIGINT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    INDEX (password)
	);`)
	c := &DBTrafficCounter{
		db:          db,
		trafficChan: make(chan *traffic, 1024),
		ctx:         context.Background(),
	}
	go c.dbDaemon()
	return c, nil
}
