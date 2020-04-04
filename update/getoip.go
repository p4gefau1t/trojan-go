package update

import (
	"fmt"

	"net"

	"github.com/golang/protobuf/proto"
	"v2ray.com/core/app/router"
)

func downloadGeoIP(filename string) error {
	url := "https://raw.githubusercontent.com/v2ray/geoip/release/geoip.dat"
	return downloadFile(url, filename)
}

func parseGeoIP(geoip []byte, country string) (string, error) {
	geoIPList := new(router.GeoIPList)
	err := proto.Unmarshal(geoip, geoIPList)
	if err != nil {
		return "", err
	}
	entry := geoIPList.GetEntry()
	result := ""
	for _, e := range entry {
		if e.CountryCode == country {
			cidrList := e.GetCidr()
			for _, cidr := range cidrList {
				addr := net.IPAddr{
					IP: cidr.GetIp(),
				}
				cidrStr := fmt.Sprintf("%s/%d", addr.String(), cidr.GetPrefix())
				_, realCidr, err := net.ParseCIDR(cidrStr)
				if err != nil {
					return "", err
				}
				fmt.Println(realCidr)
				result += realCidr.String() + "\n"
			}
		}
	}
	return result, nil
}
