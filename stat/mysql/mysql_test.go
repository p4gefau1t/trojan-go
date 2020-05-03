package mysql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/stat/memory"
	_ "github.com/proullon/ramsql/driver"
)

func TestDBAuth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	db, err := sql.Open("ramsql", "TestLoadUserAddresses")
	common.Must(err)
	common.Must2(db.Exec(`
		CREATE TABLE users (
		password CHAR(56) NOT NULL,
		quota BIGINT NOT NULL DEFAULT 0,
		download BIGINT NOT NULL DEFAULT 0,
		upload BIGINT NOT NULL DEFAULT 0,
		);
	`))
	common.Must2(db.Exec(`INSERT INTO users (password, quota, download, upload) VALUES ("hashhash", 20000, 0, 0);`))
	memoryAuth, err := memory.NewMemoryAuth(ctx, &conf.GlobalConfig{})
	auth := &DBAuth{
		db:                  db,
		ctx:                 ctx,
		updateDuration:      time.Second,
		MemoryAuthenticator: memoryAuth.(*memory.MemoryAuthenticator),
	}
	go auth.updater()
	time.Sleep(time.Second * 5)
	valid, _ := auth.AuthUser("hashhash")
	if !valid {
		t.Fail()
	}
	time.Sleep(time.Second * 5)
	valid, _ = auth.AuthUser("hashhash")
	valid, _ = auth.AuthUser("hashhash")
	common.Must2(db.Exec(`DELETE FROM users WHERE password="hashhash"`))
	time.Sleep(time.Second * 5)
	valid, _ = auth.AuthUser("hashhash")
	if valid {
		t.Fail()
	}
	cancel()
}
