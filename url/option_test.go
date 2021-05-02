package url

import (
	"testing"
	"time"

	_ "github.com/p4gefau1t/trojan-go/proxy/client"
)

func TestUrl_Handle(t *testing.T) {
	urlCases := []string{
		"trojan-go://password@server.com",
		"trojan-go://password@server.com/?type=ws&host=baidu.com&path=%2fwspath",
		"trojan-go://password@server.com/?encryption=ss%3baes-256-gcm%3afuckgfw",
		"trojan-go://password@server.com/?type=ws&host=baidu.com&path=%2fwspath&encryption=ss%3Baes-256-gcm%3Afuckgfw",
	}
	optionCases := []string{
		"mux=true;listen=127.0.0.1:0",
		"mux=false;listen=127.0.0.1:0",
		"mux=false;listen=127.0.0.1:0;api=127.0.0.1:0",
	}

	for _, s := range urlCases {
		for _, option := range optionCases {
			s := s
			option := option
			u := &url{
				url:    &s,
				option: &option,
			}
			u.Name()
			u.Priority()

			errChan := make(chan error, 1)
			go func() {
				errChan <- u.Handle()
			}()

			select {
			case err := <-errChan:
				t.Fatal(err)
			case <-time.After(time.Second * 1):
			}
		}
	}
}
