package guide

import "fmt"

func CertGuide() {
	logger.Info("Guide mode: cert")
	fmt.Println("Your domain name:")
	var domain, email string
	fmt.Scanf("%s", &domain)
	fmt.Println("Your email:")
	fmt.Scanf("%s", &email)

	err := CreateCert(domain, email)
	if err != nil {
		logger.Error("Failed to create cert")
		logger.Error(err)
		return
	}

	logger.Info("Done. Certificates and keys have been saved.")
	logger.Info("BACKUP SERVER.KEY, SERVER.CRT AND USER.KEY TO A SAFE PLACE")
}
