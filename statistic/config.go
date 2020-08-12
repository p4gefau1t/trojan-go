package statistic

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type StatisticsConfig struct {
	TrackUserIp bool `json:"track_user_ip" yaml:"track-user-ip"`
}

type Config struct {
	Statistics StatisticsConfig `json:"statistics" yaml:"statistics"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			Statistics: StatisticsConfig{
				TrackUserIp: false,
			},
		}
	})
}
