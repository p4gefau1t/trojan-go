package geodata_test

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/common/geodata"
)

const (
	geoipURL   = "https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat"
	geositeURL = "https://raw.githubusercontent.com/v2fly/domain-list-community/release/dlc.dat"
)

func init() {
	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "test", "temp")
	os.Setenv("TROJAN_GO_LOCATION_ASSET", tempPath)

	geoipPath := common.GetAssetLocation("geoip.dat")
	geositePath := common.GetAssetLocation("geosite.dat")

	if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0755))
		geoipBytes, err := common.FetchHTTPContent(geoipURL)
		common.Must(err)
		common.Must(common.WriteFile(geoipPath, geoipBytes))
	}
	if _, err := os.Stat(geositePath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0755))
		geositeBytes, err := common.FetchHTTPContent(geositeURL)
		common.Must(err)
		common.Must(common.WriteFile(geositePath, geositeBytes))
	}
}

func TestDecodeGeoIP(t *testing.T) {
	filename := common.GetAssetLocation("geoip.dat")
	result, err := geodata.Decode(filename, "test")
	if err != nil {
		t.Error(err)
	}

	expected := []byte{10, 4, 84, 69, 83, 84, 18, 8, 10, 4, 127, 0, 0, 0, 16, 8}
	if !bytes.Equal(result, expected) {
		t.Errorf("failed to load geoip:test, expected: %v, got: %v", expected, result)
	}
}

func TestDecodeGeoSite(t *testing.T) {
	filename := common.GetAssetLocation("geosite.dat")
	result, err := geodata.Decode(filename, "test")
	if err != nil {
		t.Error(err)
	}

	expected := []byte{10, 4, 84, 69, 83, 84, 18, 20, 8, 3, 18, 16, 116, 101, 115, 116, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109}
	if !bytes.Equal(result, expected) {
		t.Errorf("failed to load geosite:test, expected: %v, got: %v", expected, result)
	}
}
