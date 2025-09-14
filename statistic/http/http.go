package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
)

type Authenticator struct {
	*memory.Authenticator
	apiURL  string
	timeout int
}

type AuthRequest struct {
	Auth string `json:"auth"`
	Addr string `json:"addr"`
	Tx   int    `json:"tx"`
}

type AuthResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message",omitempty`
	Error   string `json:"error",omitempty`
}

func (a *Authenticator) AuthUser(hash string) (bool, statistic.User) {
	// requset apiURL
	rawRequest := AuthRequest{
		Auth: hash,
		Addr: "0.0.0.0",
		Tx:   0,
	}
	requestBody, err := json.Marshal(rawRequest)
	if err != nil {
		return false, nil
	}
	request, err := http.NewRequest("POST", a.apiURL, bytes.NewReader(requestBody))
	if err != nil {
		return false, nil
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, nil
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Errorf("http authenticator failed to read body: %s", err)
		return false, nil
	}
	var bodyJson AuthResponse
	err = json.Unmarshal(body, &bodyJson)
	if err != nil {
		log.Errorf("http authenticator failed to unmarshal body: %s", err)
		return false, nil
	}
	if !bodyJson.Ok {
		log.Warnf("http authenticator failed to auth user: %s", hash)
		return false, nil
	}
	err = a.AddUser(hash)
	if err == memory.ErrorHashExist || err == nil {
		user, err := a.GetUser(hash)
		if err != nil {
			return false, nil
		}
		return true, user
	}
	user, err := a.GetUser(hash)
	if err != nil {
		return false, nil
	}
	return true, user
}

func NewAuthenticator(ctx context.Context) (statistic.Authenticator, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	apiURL := cfg.HTTP.APIURL
	timeout := cfg.HTTP.Timeout
	if apiURL == "" {
		return nil, common.NewError("API URL is required")
	}
	memoryAuth, err := memory.NewAuthenticator(ctx)
	if err != nil {
		return nil, err
	}
	a := &Authenticator{
		Authenticator: memoryAuth.(*memory.Authenticator),
		apiURL:        apiURL,
		timeout:       timeout,
	}
	log.Debug("http authenticator created")
	return a, nil
}

func init() {
	statistic.RegisterAuthenticatorCreator(Name, NewAuthenticator)
}
