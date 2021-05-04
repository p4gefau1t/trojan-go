package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
)

func HumanFriendlyTraffic(bytes uint64) string {
	if bytes <= KiB {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes <= MiB {
		return fmt.Sprintf("%.2f KiB", float32(bytes)/KiB)
	}
	if bytes <= GiB {
		return fmt.Sprintf("%.2f MiB", float32(bytes)/MiB)
	}
	return fmt.Sprintf("%.2f GiB", float32(bytes)/GiB)
}

func PickPort(network string, host string) int {
	switch network {
	case "tcp":
		for retry := 0; retry < 16; retry++ {
			l, err := net.Listen("tcp", host+":0")
			if err != nil {
				continue
			}
			defer l.Close()
			_, port, err := net.SplitHostPort(l.Addr().String())
			Must(err)
			p, err := strconv.ParseInt(port, 10, 32)
			Must(err)
			return int(p)
		}
	case "udp":
		for retry := 0; retry < 16; retry++ {
			conn, err := net.ListenPacket("udp", host+":0")
			if err != nil {
				continue
			}
			defer conn.Close()
			_, port, err := net.SplitHostPort(conn.LocalAddr().String())
			Must(err)
			p, err := strconv.ParseInt(port, 10, 32)
			Must(err)
			return int(p)
		}
	default:
		return 0
	}
	return 0
}

func WriteAllBytes(writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		n, err := writer.Write(payload)
		if err != nil {
			return err
		}
		payload = payload[n:]
	}
	return nil
}

func WriteFile(path string, payload []byte) error {
	writer, err := os.Create(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	return WriteAllBytes(writer, payload)
}

func FetchHTTPContent(target string) ([]byte, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", target)
	}

	if s := strings.ToLower(parsedTarget.Scheme); s != "http" && s != "https" {
		return nil, fmt.Errorf("invalid scheme: %s", parsedTarget.Scheme)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    parsedTarget,
		Close:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial to %s", target)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response")
	}

	return content, nil
}
