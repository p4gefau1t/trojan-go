package service

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/p4gefau1t/trojan-go/api"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
)

type ServerAPI struct {
	auth   statistic.Authenticator
	apiKey string
	server *http.Server
}

// 密钥验证中间件
func (s *ServerAPI) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		if apiKey == "" || apiKey != s.apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// 设置响应头
		w.Header().Set("Content-Type", "application/json")
		log.Debugf("request for %s, apiKey: %s", r.URL.Path, apiKey)
		next(w, r)
	}
}

// 获取流量统计
func (s *ServerAPI) getTraffics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// https://t43.165.190.29.radio238.com/
	isClear := r.URL.Query().Get("clear")
	users := s.auth.ListUsers()
	trafficData := make(map[string]interface{})
	for _, user := range users {
		sent, recv := user.GetTraffic()
		trafficData[user.Hash()] = map[string]interface{}{
			"rx": recv,
			"tx": sent,
		}
	}

	if isClear == "1" {
		for _, user := range users {
			user.ResetTraffic()
		}
	}
	log.Infof("trafficData: %v", trafficData)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trafficData)
}

// 创建 HTTP 服务器
func newHTTPServer(cfg *Config, auth statistic.Authenticator) (*ServerAPI, error) {
	apiKey := cfg.API.APIKey
	if apiKey == "" {
		return nil, common.NewError("API key is required")
	}

	server := &ServerAPI{
		auth:   auth,
		apiKey: apiKey,
	}

	// 设置路由
	mux := http.NewServeMux()

	// 应用认证中间件
	mux.HandleFunc("/traffic", server.authMiddleware(server.getTraffics))

	// 创建 HTTP 服务器
	httpServer := &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.server = httpServer
	return server, nil
}

func RunServerAPI(ctx context.Context, auth statistic.Authenticator) error {
	cfg := config.FromContext(ctx, Name).(*Config)
	if !cfg.API.Enabled {
		return nil
	}

	server, err := newHTTPServer(cfg, auth)
	if err != nil {
		return err
	}

	// 解析地址
	addr, err := net.ResolveIPAddr("ip", cfg.API.APIHost)
	if err != nil {
		return common.NewError("api found invalid addr").Base(err)
	}

	server.server.Addr = (&net.TCPAddr{
		IP:   addr.IP,
		Port: cfg.API.APIPort,
		Zone: addr.Zone,
	}).String()

	log.Info("HTTP API server is listening on", server.server.Addr)

	// 启动服务器
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.server.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Debug("HTTP API server closed")
		server.server.Shutdown(context.Background())
		return nil
	}
}

func init() {
	api.RegisterHandler(trojan.Name+"_SERVER", RunServerAPI)
}
