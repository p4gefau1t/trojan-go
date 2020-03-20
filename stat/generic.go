package stat

import (
	"database/sql"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

type ConfigUserAuthenticator struct {
	Config *conf.GlobalConfig
	Authenticator
}

func (a *ConfigUserAuthenticator) CheckHash(hash string) bool {
	_, found := a.Config.Hash[hash]
	return found
}

func (a *ConfigUserAuthenticator) Close() error {
	return nil
}

type MixedAuthenticator struct {
	dbAuth     Authenticator
	configAuth Authenticator
	Authenticator
}

func (a *MixedAuthenticator) CheckHash(hash string) bool {
	if a.configAuth.CheckHash(hash) {
		return true
	} else if a.dbAuth.CheckHash(hash) {
		return true
	}
	return false
}

func (a *MixedAuthenticator) Close() error {
	return a.dbAuth.Close()
}

func NewMixedAuthenticator(config *conf.GlobalConfig, db *sql.DB) (Authenticator, error) {
	dbAuth, err := NewDBAuthenticator(db)
	common.Must(err)
	a := &MixedAuthenticator{
		configAuth: &ConfigUserAuthenticator{
			Config: config,
		},
		dbAuth: dbAuth,
	}
	return a, nil
}
