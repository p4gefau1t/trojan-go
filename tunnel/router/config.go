package router

import (
	"os"
	"path/filepath"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
)

type Config struct {
	Router RouterConfig `json:"router" yaml:"router"`
}

type RouterConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	Bypass          []string `json:"bypass" yaml:"bypass"`
	Proxy           []string `json:"proxy" yaml:"proxy"`
	Block           []string `json:"block" yaml:"block"`
	DomainStrategy  string   `json:"domain_strategy" yaml:"domain-strategy"`
	DefaultPolicy   string   `json:"default_policy" yaml:"default-policy"`
	GeoIPFilename   string   `json:"geoip" yaml:"geoip"`
	GeoSiteFilename string   `json:"geosite" yaml:"geosite"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		cfg := &Config{
			Router: RouterConfig{
				DefaultPolicy:   "proxy",
				DomainStrategy:  "as_is",
				GeoIPFilename:   filepath.Join(common.GetProgramDir(), "geoip.dat"),
				GeoSiteFilename: filepath.Join(common.GetProgramDir(), "geosite.dat"),
			},
		}
		if path := os.Getenv("TROJAN_GO_LOCATION_ASSET"); path != "" {
			cfg.Router.GeoIPFilename = filepath.Join(path, "geoip.dat")
			cfg.Router.GeoSiteFilename = filepath.Join(path, "geosite.dat")
			log.Debug("env set:", path)
		}
		return cfg
	})
}
