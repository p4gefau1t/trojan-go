package sqlite

import (
	"errors"

	"github.com/p4gefau1t/trojan-go/statistic"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const Name = "sqlite"

type Persistencer struct {
	db *gorm.DB
}

func NewSqlitePersistencer(path string) (*Persistencer, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}
	sp := &Persistencer{
		db: db,
	}

	return sp, nil
}

func (p *Persistencer) SaveUser(u statistic.Metadata) error {
	if u == nil {
		return errors.New("user is nil")
	}
	ls, lr := u.GetSpeedLimit()
	usr := &User{
		Hash:      u.GetHash(),
		MaxIPNum:  u.GetIPLimit(),
		SendLimit: ls,
		RecvLimit: lr,
		Sent:      make([]byte, 8),
		Recv:      make([]byte, 8),
	}
	ts, tr := u.GetTraffic()
	usr.setSent(ts)
	usr.setRecv(tr)
	err := p.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "hash"}},
		UpdateAll: true,
	}).Create(usr).Error
	if err != nil {
		return err
	}
	return nil
}

func (p *Persistencer) LoadUser(hash string) (statistic.Metadata, error) {
	var u User
	err := p.db.First(&u, "hash = ?", hash).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (p *Persistencer) DeleteUser(hash string) error {
	err := p.db.Delete(&User{Hash: hash}).Error
	if err != nil {
		return err
	}
	return nil
}

func (p *Persistencer) ListUser(f func(hash string, u statistic.Metadata) bool) error {
	users := make([]User, 0)
	err := p.db.Find(&users).Error
	if err != nil {
		return err
	}
	for _, u := range users {
		if goOn := f(u.Hash, &u); !goOn {
			break
		}
	}
	return nil
}

func (p *Persistencer) UpdateUserTraffic(hash string, sent, recv uint64) error {
	u := &User{
		Hash: hash,
		Sent: make([]byte, 8),
		Recv: make([]byte, 8),
	}
	u.setSent(sent)
	u.setRecv(recv)
	return p.db.Model(&User{Hash: hash}).Updates(u).Error
}
