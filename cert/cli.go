package cert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
)

type domainInfo struct {
	Domain string
	Email  string
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		logger.Fatal(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

func RequestCertGuide() {
	//caDir = "https://127.0.0.1:14000/dir"
	logger.Info("Guide mode: request cert")

	logger.Warn("To perform a ACME challenge, trojan-go need the ROOT PRIVILEGE to bind port 80 and 443")
	logger.Warn("Please make sure you HAVE sudo this program, and port 80/443 is NOT used by other process at this moment")
	logger.Info("Continue? (y/n)")

	if !askForConfirmation() {
		return
	}

	data, err := ioutil.ReadFile("domain_info.json")
	info := &domainInfo{}

	if err != nil {
		fmt.Println("Your domain name:")
		fmt.Scanf("%s", &info.Domain)
		fmt.Println("Your email:")
		fmt.Scanf("%s", &info.Email)
	} else {
		logger.Info("domain_info.json found")
		if err := json.Unmarshal(data, info); err != nil {
			logger.Error(common.NewError("failed to parse domain_info.json").Base(err))
			return
		}
	}

	fmt.Printf("Domain: %s, Email: %s\n", info.Domain, info.Email)
	fmt.Println("Is that correct? (y/n)")

	if !askForConfirmation() {
		return
	}

	data, err = json.Marshal(info)
	common.Must(err)
	ioutil.WriteFile("domain_info.json", data, os.ModePerm)

	if err := RequestCert(info.Domain, info.Email); err != nil {
		logger.Error(common.NewError("Failed to create cert").Base(err))
		return
	}

	logger.Info("All done. Certificates has been saved to server.crt and server.key")
	logger.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	logger.Warn("BACKUP DOMAIN_INFO.JSON, SERVER.KEY, SERVER.CRT AND USER.KEY TO A SAFE PLACE")
	logger.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
}

func RenewCertGuide() {
	//caDir = "https://127.0.0.1:14000/dir"
	logger.Info("Guide mode: renew cert")

	logger.Warn("To perform a ACME challenge, trojan-go need the ROOT PRIVILEGE to bind port 80 and 443")
	logger.Warn("Please make sure you HAVE sudo this program, and port 80/443 is NOT used by other process at this moment")
	logger.Info("Continue? (y/n)")

	if !askForConfirmation() {
		return
	}

	data, err := ioutil.ReadFile("domain_info.json")
	if err != nil {
		logger.Error(err)
		return
	}

	info := &domainInfo{}
	if err := json.Unmarshal(data, info); err != nil {
		logger.Error(err)
	}

	fmt.Printf("Domain: %s, Email: %s\n", info.Domain, info.Email)
	fmt.Println("Is that correct? (y/n)")

	if !askForConfirmation() {
		return
	}

	if err := RenewCert(info.Domain, info.Email); err != nil {
		logger.Error(common.NewError("Failed to renew cert").Base(err))
		return
	}
	logger.Info("All done")
}
