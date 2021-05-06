package geodata

import (
	"runtime"

	v2router "github.com/v2fly/v2ray-core/v4/app/router"
)

type geodataLoader interface {
	LoadIP(filename, country string) ([]*v2router.CIDR, error)
	LoadSite(filename, list string) ([]*v2router.Domain, error)
	LoadGeoIP(country string) ([]*v2router.CIDR, error)
	LoadGeoSite(list string) ([]*v2router.Domain, error)
}

func GetGeodataLoader() geodataLoader {
	return &geodataCache{
		make(map[string]*v2router.GeoIP),
		make(map[string]*v2router.GeoSite),
	}
}

type geodataCache struct {
	GeoIPCache
	GeoSiteCache
}

func (g *geodataCache) LoadIP(filename, country string) ([]*v2router.CIDR, error) {
	geoip, err := g.GeoIPCache.Unmarshal(filename, country)
	if err != nil {
		return nil, err
	}
	runtime.GC()
	return geoip.Cidr, nil
}

func (g *geodataCache) LoadSite(filename, list string) ([]*v2router.Domain, error) {
	geosite, err := g.GeoSiteCache.Unmarshal(filename, list)
	if err != nil {
		return nil, err
	}
	runtime.GC()
	return geosite.Domain, nil
}

func (g *geodataCache) LoadGeoIP(country string) ([]*v2router.CIDR, error) {
	return g.LoadIP("geoip.dat", country)
}

func (g *geodataCache) LoadGeoSite(list string) ([]*v2router.Domain, error) {
	return g.LoadSite("geosite.dat", list)
}
