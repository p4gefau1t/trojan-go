package geodata

import (
	"io/ioutil"
	"strings"

	v2router "github.com/v2fly/v2ray-core/v4/app/router"
	"google.golang.org/protobuf/proto"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
)

type geoipCache map[string]*v2router.GeoIP

func (g geoipCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g geoipCache) Get(key string) *v2router.GeoIP {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g geoipCache) Set(key string, value *v2router.GeoIP) {
	if g == nil {
		g = make(map[string]*v2router.GeoIP)
	}
	g[key] = value
}

func (g geoipCache) Unmarshal(filename, code string) (*v2router.GeoIP, error) {
	asset := common.GetAssetLocation(filename)
	idx := strings.ToLower(asset + ":" + code)
	if g.Has(idx) {
		log.Debugf("geoip cache HIT: %s -> %s", code, idx)
		return g.Get(idx), nil
	}

	geoipBytes, err := Decode(asset, code)
	switch err {
	case nil:
		var geoip v2router.GeoIP
		if err := proto.Unmarshal(geoipBytes, &geoip); err != nil {
			return nil, err
		}
		g.Set(idx, &geoip)
		return &geoip, nil

	case ErrCodeNotFound:
		return nil, common.NewError("country code " + code + " not found in " + filename)

	case ErrFailedToReadBytes, ErrFailedToReadExpectedLenBytes,
		ErrInvalidGeodataFile, ErrInvalidGeodataVarintLength:
		log.Warnf("failed to decode geoip file: %s, fallback to the original ReadFile method", filename)
		geoipBytes, err = ioutil.ReadFile(asset)
		if err != nil {
			return nil, err
		}
		var geoipList v2router.GeoIPList
		if err := proto.Unmarshal(geoipBytes, &geoipList); err != nil {
			return nil, err
		}
		for _, geoip := range geoipList.GetEntry() {
			if strings.EqualFold(code, geoip.GetCountryCode()) {
				g.Set(idx, geoip)
				return geoip, nil
			}
		}

	default:
		return nil, err
	}

	return nil, common.NewError("country code " + code + " not found in " + filename)
}

type geositeCache map[string]*v2router.GeoSite

func (g geositeCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g geositeCache) Get(key string) *v2router.GeoSite {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g geositeCache) Set(key string, value *v2router.GeoSite) {
	if g == nil {
		g = make(map[string]*v2router.GeoSite)
	}
	g[key] = value
}

func (g geositeCache) Unmarshal(filename, code string) (*v2router.GeoSite, error) {
	asset := common.GetAssetLocation(filename)
	idx := strings.ToLower(asset + ":" + code)
	if g.Has(idx) {
		log.Debugf("geosite cache HIT: %s -> %s", code, idx)
		return g.Get(idx), nil
	}

	geositeBytes, err := Decode(asset, code)
	switch err {
	case nil:
		var geosite v2router.GeoSite
		if err := proto.Unmarshal(geositeBytes, &geosite); err != nil {
			return nil, err
		}
		g.Set(idx, &geosite)
		return &geosite, nil

	case ErrCodeNotFound:
		return nil, common.NewError("list " + code + " not found in " + filename)

	case ErrFailedToReadBytes, ErrFailedToReadExpectedLenBytes,
		ErrInvalidGeodataFile, ErrInvalidGeodataVarintLength:
		log.Warnf("failed to decode geoip file: %s, fallback to the original ReadFile method", filename)
		geositeBytes, err = ioutil.ReadFile(asset)
		if err != nil {
			return nil, err
		}
		var geositeList v2router.GeoSiteList
		if err := proto.Unmarshal(geositeBytes, &geositeList); err != nil {
			return nil, err
		}
		for _, geosite := range geositeList.GetEntry() {
			if strings.EqualFold(code, geosite.GetCountryCode()) {
				g.Set(idx, geosite)
				return geosite, nil
			}
		}

	default:
		return nil, err
	}

	return nil, common.NewError("list " + code + " not found in " + filename)
}
