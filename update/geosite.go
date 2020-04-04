package update

func downloadGeoSite(filename string) error {
	url := "https://github.com/v2ray/domain-list-community/raw/release/dlc.dat"
	return downloadFile(url, filename)
}
