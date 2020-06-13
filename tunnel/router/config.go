package router

import (
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"os"
)

type Config struct {
	Router RouterConfig `json,yaml:"router"`
}

type RouterConfig struct {
	Enabled         bool     `json,yaml:"enabled"`
	Bypass          []string `json,yaml:"bypass"`
	Proxy           []string `json,yaml:"proxy"`
	Block           []string `json,yaml:"block"`
	DomainStrategy  string   `json:"domain_strategy" yaml:"domain-strategy"`
	DefaultPolicy   string   `json:"default_policy" yaml:"default-policy"`
	GeoIPFilename   string   `json,yaml:"geoip"`
	GeoSiteFilename string   `json,yaml:"geosite"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		cfg := &Config{
			Router: RouterConfig{
				DefaultPolicy:   "proxy",
				DomainStrategy:  "as_is",
				GeoIPFilename:   common.GetProgramDir() + "/geoip.dat",
				GeoSiteFilename: common.GetProgramDir() + "/geosite.dat",
			},
		}
		if path := os.Getenv("TROJAN_GO_LOCATION_ASSET"); path != "" {
			cfg.Router.GeoIPFilename = path + "/geoip.dat"
			cfg.Router.GeoSiteFilename = path + "/geosite.dat"
			log.Debug("env set:", path)
		}
		return cfg
	})
}
