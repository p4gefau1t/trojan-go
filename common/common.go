package common

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	Version = "v0.6.0"
)

type Runnable interface {
	Run() error
	Close() error
}

func NewBufioReadWriter(rw io.ReadWriter) *bufio.ReadWriter {
	if bufrw, ok := rw.(*bufio.ReadWriter); ok {
		return bufrw
	}
	return bufio.NewReadWriter(bufio.NewReader(rw), bufio.NewWriter(rw))
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
