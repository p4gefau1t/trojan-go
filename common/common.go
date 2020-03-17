package common

import (
	"crypto/sha256"
	"fmt"
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
