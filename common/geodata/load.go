package geodata

import (
	"runtime"

	v2router "github.com/v2fly/v2ray-core/v4/app/router"
)

var geoipcache GeoIPCache = make(map[string]*v2router.GeoIP)
var geositecache GeoSiteCache = make(map[string]*v2router.GeoSite)

func LoadIP(filename, country string) ([]*v2router.CIDR, error) {
	geoip, err := geoipcache.Unmarshal(filename, country)
	if err != nil {
		return nil, err
	}
	runtime.GC()
	return geoip.Cidr, nil
}

func LoadGeoIP(country string) ([]*v2router.CIDR, error) {
	return LoadIP("geoip.dat", country)
}

func LoadSite(filename, list string) ([]*v2router.Domain, error) {
	geosite, err := geositecache.Unmarshal(filename, list)
	if err != nil {
		return nil, err
	}
	runtime.GC()
	return geosite.Domain, nil
}

func LoadGeoSite(list string) ([]*v2router.Domain, error) {
	return LoadSite("geosite.dat", list)
}
