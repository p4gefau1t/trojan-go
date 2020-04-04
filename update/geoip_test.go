package update

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestDownloadGeoIP(t *testing.T) {
	err := downloadGeoIP("geoip.dat")
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseGeoIP(t *testing.T) {
	geoip, err := ioutil.ReadFile("geoip.dat")
	if err != nil {
		t.Fatal(err)
	}
	result, err := parseGeoIP(geoip, "CN")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
}
