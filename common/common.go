package common

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/p4gefau1t/trojan-go/log"
)

type Runnable interface {
	Run() error
	Close() error
}

func SHA224String(password string) string {
	hash := sha256.New224()
	hash.Write([]byte(password))
	val := hash.Sum(nil)
	str := ""
	for _, v := range val {
		str += fmt.Sprintf("%02x", v)
	}
	return str
}

func GetProgramDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func GetAssetLocation(file string) string {
	if filepath.IsAbs(file) {
		return file
	}
	if loc := os.Getenv("TROJAN_GO_LOCATION_ASSET"); loc != "" {
		log.Debugf("env set: TROJAN_GO_LOCATION_ASSET=%s", loc)
		return filepath.Join(loc, file)
	}
	return filepath.Join(GetProgramDir(), file)
}
