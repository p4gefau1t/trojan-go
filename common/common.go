package common

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	//_ "github.com/mattn/go-sqlite3"
)

type Runnable interface {
	Run() error
	Close() error
}

func NewBufReadWriter(rw io.ReadWriter) *bufio.ReadWriter {
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

const (
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
)

func HumanFriendlyTraffic(bytes int) string {
	if bytes <= KiB {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes <= MiB {
		return fmt.Sprintf("%.2f KiB", float32(bytes)/KiB)
	}
	if bytes <= GiB {
		return fmt.Sprintf("%.2f MiB", float32(bytes)/MiB)
	}
	return fmt.Sprintf("%.2f TiB", float32(bytes)/GiB)
}

func ConnectDatabase(driverName, username, password, ip string, port int, dbName string) (*sql.DB, error) {
	path := strings.Join([]string{username, ":", password, "@tcp(", ip, ":", fmt.Sprintf("%d", port), ")/", dbName, "?charset=utf8"}, "")
	return sql.Open(driverName, path)
}

func ConnectSQLite(dbName string) (*sql.DB, error) {
	//for debug only
	return sql.Open("sqlite3", dbName)
}
