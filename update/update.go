package update

import (
	"io"
	"net/http"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
)

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func downloadGeoSite() error {
	filename := "geosite.dat"
	url := "https://github.com/v2ray/domain-list-community/raw/release/dlc.dat"
	return downloadFile(url, filename)
}

func downloadGeoIP() error {
	filename := "geoip.dat"
	url := "https://raw.githubusercontent.com/v2ray/geoip/release/geoip.dat"
	return downloadFile(url, filename)
}

type updateOption struct {
	args *string
	common.OptionHandler
}

func init() {

}
