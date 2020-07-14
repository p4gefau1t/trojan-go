package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestServerAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = config.WithConfig(ctx, memory.Name,
		&memory.Config{
			Passwords: []string{},
		})
	port := common.PickPort("tcp", "127.0.0.1")
	ctx = config.WithConfig(ctx, Name, &Config{
		APIConfig{
			Enabled: true,
			APIHost: "127.0.0.1",
			APIPort: port,
		},
	})
	auth, err := memory.NewAuthenticator(ctx)
	common.Must(err)
	go RunServerAPI(ctx, auth)
	time.Sleep(time.Second * 3)
	common.Must(auth.AddUser("hash1234"))
	_, user := auth.AuthUser("hash1234")
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
	common.Must(err)
	server := NewTrojanServerServiceClient(conn)
	stream1, err := server.ListUsers(ctx, &ListUsersRequest{})
	common.Must(err)
	for {
		resp, err := stream1.Recv()
		if err != nil {
			break
		}
		fmt.Println(resp.Status.User.Hash)
		if resp.Status.User.Hash != "hash1234" {
			t.Fail()
		}
		fmt.Println(resp.Status.SpeedCurrent)
		fmt.Println(resp.Status.SpeedLimit)
	}
	stream1.CloseSend()
	user.AddTraffic(1234, 5678)
	time.Sleep(time.Millisecond * 1000)
	stream2, err := server.GetUsers(ctx)
	common.Must(err)
	stream2.Send(&GetUsersRequest{
		User: &User{
			Hash: "hash1234",
		},
	})
	resp2, err := stream2.Recv()
	common.Must(err)
	if resp2.Status.TrafficTotal.DownloadTraffic != 1234 || resp2.Status.TrafficTotal.UploadTraffic != 5678 {
		t.Fatal("wrong traffic")
	}
	if resp2.Status.SpeedCurrent.DownloadSpeed != 1234 || resp2.Status.TrafficTotal.UploadTraffic != 5678 {
		t.Fatal("wrong speed")
	}

	stream3, err := server.SetUsers(ctx)
	stream3.Send(&SetUsersRequest{
		Status: &UserStatus{
			User: &User{
				Hash: "hash1234",
			},
		},
		Operation: SetUsersRequest_Delete,
	})
	resp3, err := stream3.Recv()
	if err != nil || !resp3.Success {
		t.Fatal("user not exists")
	}
	valid, _ := auth.AuthUser("hash1234")
	if valid {
		t.Fatal("failed to auth")
	}
	stream3.Send(&SetUsersRequest{
		Status: &UserStatus{
			User: &User{
				Hash: "newhash",
			},
		},
		Operation: SetUsersRequest_Add,
	})
	resp3, err = stream3.Recv()
	if err != nil || !resp3.Success {
		t.Fatal("failed to read")
	}
	valid, user = auth.AuthUser("newhash")
	if !valid {
		t.Fatal("failed to auth 2")
	}
	stream3.Send(&SetUsersRequest{
		Status: &UserStatus{
			User: &User{
				Hash: "newhash",
			},
			SpeedLimit: &Speed{
				DownloadSpeed: 5000,
				UploadSpeed:   3000,
			},
			TrafficTotal: &Traffic{
				DownloadTraffic: 1,
				UploadTraffic:   1,
			},
		},
		Operation: SetUsersRequest_Modify,
	})
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			user.AddTraffic(200, 0)
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			user.AddTraffic(0, 300)
		}
	}()
	time.Sleep(time.Second * 3)
	for i := 0; i < 3; i++ {
		stream2.Send(&GetUsersRequest{
			User: &User{
				Hash: "newhash",
			},
		})
		resp2, err = stream2.Recv()
		fmt.Println(resp2.Status.SpeedCurrent)
		fmt.Println(resp2.Status.SpeedLimit)
		time.Sleep(time.Second)
	}
	stream2.CloseSend()
	cancel()
}

func TestTLS(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	cfg := &Config{
		API: APIConfig{
			Enabled: true,
			APIHost: "127.0.0.1",
			APIPort: port,
			SSL: SSLConfig{
				Enabled:        true,
				CertPath:       "server.crt",
				KeyPath:        "server.key",
				VerifyClient:   false,
				ClientCertPath: []string{"client.crt"},
			},
		},
	}

	ctx := config.WithConfig(context.Background(), Name, cfg)
	ctx = config.WithConfig(ctx, memory.Name,
		&memory.Config{
			Passwords: []string{},
		})

	auth, err := memory.NewAuthenticator(ctx)
	common.Must(err)
	go func() {
		common.Must(RunServerAPI(ctx, auth))
	}()
	time.Sleep(time.Second)
	pool := x509.NewCertPool()
	certBytes, err := ioutil.ReadFile("server.crt")
	pool.AppendCertsFromPEM(certBytes)

	certificate, err := tls.LoadX509KeyPair("client.crt", "client.key")
	common.Must(err)
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   "localhost",
		RootCAs:      pool,
		Certificates: []tls.Certificate{certificate},
	})
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithTransportCredentials(creds))
	common.Must(err)
	server := NewTrojanServerServiceClient(conn)
	stream, err := server.ListUsers(ctx, &ListUsersRequest{})
	common.Must(err)
	stream.CloseSend()
	conn.Close()
}

func init() {
	var serverCert = `
-----BEGIN CERTIFICATE-----
MIIDZTCCAk0CFFphZh018B5iAD9F5fV4y0AlD0LxMA0GCSqGSIb3DQEBCwUAMG8x
CzAJBgNVBAYTAlVTMQ0wCwYDVQQIDARNYXJzMRMwEQYDVQQHDAppVHJhbnN3YXJw
MRMwEQYDVQQKDAppVHJhbnN3YXJwMRMwEQYDVQQLDAppVHJhbnN3YXJwMRIwEAYD
VQQDDAlsb2NhbGhvc3QwHhcNMjAwMzMxMTAwMDUxWhcNMzAwMzI5MTAwMDUxWjBv
MQswCQYDVQQGEwJVUzENMAsGA1UECAwETWFyczETMBEGA1UEBwwKaVRyYW5zd2Fy
cDETMBEGA1UECgwKaVRyYW5zd2FycDETMBEGA1UECwwKaVRyYW5zd2FycDESMBAG
A1UEAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
ml44fThYMkCcT627o7ibEs7mq2WOhImjDwYijYJ1684BatrCsHJNcw8PJGTuP+tg
GdngmALjA3l+RipjaE/UK4FJrAjruphA/hOCjZfWqk8KBR4qk0OltxCMWJlp/XCM
9ny1ogFdWUlBbqThs4NWSOUESgxf/Be2njeiOrngGR31qxSiLCLBvafIhKqq/4av
Rlx0Ht770uvF97MlAj1ASAvzTZICHAfUZxEdWl0J4MBbG7SNcnMBbyAF+s60eFTa
4RGMfRGnUa2Fzz/gfjhvfSIGeLQ3JRG6sl6jkc5xe0PZzhq3UNpK0gtQ48yy9CSP
neZnrynoKks7XC2bizsr3QIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQAHS/xuG5+F
yGU3N6V4kv+HbKqHaXNOq4zKVsCc1k7vg4MFFpKUJKxtJYooCI8n2ypp5XRUTIGQ
bmEbVcIPqm9Rf/4vHtF0falNCwieAbXDkiEHoykRmmU1UE/ccPA7X8NO9aVLJAJO
N2Li8MH0Ixgs02pQH56eyGKoRBWPR5C3ETQ9Leqvazg6Dn1iJWvmfF0mOte5228s
mZJOntF9t8MZOJdIWGdrUHn6euRfhd0btkmL/NUDzeCTwJcuPORLxkBbCP5mTC6G
GnLS5Z4oRYgCgvT2pLtcM0r48hYjwgjXFQ4zalkW6YI9LPpqwwMhhOzINlXjBaDi
Haz8uKI4EciU
-----END CERTIFICATE-----
`

	var serverKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAml44fThYMkCcT627o7ibEs7mq2WOhImjDwYijYJ1684BatrC
sHJNcw8PJGTuP+tgGdngmALjA3l+RipjaE/UK4FJrAjruphA/hOCjZfWqk8KBR4q
k0OltxCMWJlp/XCM9ny1ogFdWUlBbqThs4NWSOUESgxf/Be2njeiOrngGR31qxSi
LCLBvafIhKqq/4avRlx0Ht770uvF97MlAj1ASAvzTZICHAfUZxEdWl0J4MBbG7SN
cnMBbyAF+s60eFTa4RGMfRGnUa2Fzz/gfjhvfSIGeLQ3JRG6sl6jkc5xe0PZzhq3
UNpK0gtQ48yy9CSPneZnrynoKks7XC2bizsr3QIDAQABAoIBAFpYUo9W7qdakSFA
+NS1Mm0rkm01nteLBlfAq3BOrl030DSNm+xQuWthoOcX+yiFxVTb40qURfC+plzC
ajOepPphTJDXF7+5ZDBPktTzzLsYTzD3mstdiBtAICOqhhHCUX3hNxx91/htm1H6
Re4eK921y3DbFUIhTswCm3vrVXDc4yTXtURGllVzo40K/1Of39CpufKFdpJ81HV+
h/VW++h3o+sFV4KqcqIjClxBfDxoJpBaRlOCunTiHqZNvqO+EPqPR5zdn34werjU
xQEvPzmz+ClwnaEXQxYWgIcYQii9VNsHogDxEw4R31S7lVrUt0f0atDmGJip1lPb
E7IomAECgYEAzKQ3PzBV46nUNfVO9SODpf14Z+xYfLKouPC+Qnepwp0V0JS6zY1+
Wzskyb80drjnoQraWSEvGsX+tEWeLcnjN7JuMu/U8DPKRcQ+Q2dsVo/q4sfBOgvl
VhPNMZLfa7NIkRUx2KXku++Ep0Xtak0dskrfQrZnvhymRPyWuIMM6IECgYEAwRwL
Gt/ZZdUueE/hwT3c1hNn6igeDLOwK2t6frib+Ofw5oCAQxtTROvP1ljlnWUPkeIS
uzTusmqucalcK3lCHIsyHLwApOI/B31M971pxMVBRZ0wIbBaoarCGND7gi8JUPFR
VErGcAB5YnpRlmfLPEgw2o7DpjsDc2KmdE9oNV0CgYEAmfNEWLYtNztxGTK1treD
96ELLutf2lexlIgQKgLJ5E22tpbdPXwfvdRtpZTBjDsojj+S6hCL1lFzfv0MtZe2
5xTF0G4avKXJmti6moy4tRpJ81ehZuDCJBJ7gLrkd6qFghf2yuxqenQDUK/Lnvfq
ylGHSjHdM+lrsGRxotd8I4ECgYBoo4GA9nseqv2bQ+3YgGUBu1I7l7FwwI1decfO
ksoxfb0Tqd3WfyAH4J+mTlVdjD17lzz/JBeTpisQe+ztwa8JOIPW/ih7L/1nWYYz
V/fQH/LWfe5u0tjJcXXrbJJcYJBzw8+GFV6hoiAkNJOxJF0ENToDtAhgMuoTxAje
TYjyIQKBgQCmHkLLq0Bj3FpIOVrwo2gNvQteNPa7jkkGp4lljO8JQUHhCHDGWKEH
MUJ0EFsxS/EaQa+rW6jHhs3GyBA2TxmC783stAOOEX+hO/zpcbzdCWgp6eZ0aGMW
WS94/5WE/lwHJi8ZPSjH1AURCzXhUi4fGvBrNBtry95e+jcEvP5c0g==
-----END RSA PRIVATE KEY-----
`
	var clientKey = `
-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDDZ6UMZRadrMN7
T8xdgFfM5F8XjF5WyQt9mVqcre+47Us1Z72K3wv5kgC8cm4oqC5RqbAndJN3lDLf
i/nvdPIB0gpQ8/jE9KunnicmF7OpCDg7I9wPaZFZNDkW97vXQz8QRyvDPRA311ZY
1ltB+nZ7kpQs4WWRUWfJNAGPqo1rY5pgcGCEwTgocZhOKyGHxO4dlVXHzmlUF3nP
cuc8ge6tIx+S/EJvAF7JJWXdm3I0ckPFw6RJN81HTr7vSgm/UWOMv9ZblMNeR4dl
errZHhqAOG9c1t0GeuYu3idkaXqITjHaz9w165v33rkVXOrsdGcZOOXEB7rycv1Q
x6+YcCyHAgMBAAECggEAYM83qTVoCAQw8SXu1SAh38QBDLShhJOkWrygdOGD0/XU
fggAkw3AbAwWy0ZSJ1hzYkgUmueZq/PDZJd/40/oGljKfaLxy/qAFNI5CRlTDFqj
KUGx4/zkYvKJmkRwTszlMJZiKx9UqqXIBMlmewCwtLZBLR8aZ+2R4tAeOeRAnkPj
5kKjOFSByyNUtyqqBZ/M5idySwhbITIc0/1kJ1ULAeMlezNdSoEjkK30ck/enUMF
ouFQMMVFnGMpwe83mPDzj4Yw6bL6u3A4wiUTzckjJW+KDdU9V2QFe0fXFGpUckGJ
hMEV8t6+4/ATaHKw8opcgg9ua/9rXOjlo+sF7zXdsQKBgQDla3wf8FH3kmH/CKn4
NeC+XR+fuNVsZrP/OFXxtXPu/kJoPropD6+5DJkSEnms0rOpHedShYMgk1cPYGUI
Ol8ISj+y7qvxTgFZJGMUzZSkTKkjUDpCcN5kENCwN9gHwf+gBv9iR/BUljcjPwmi
AZCYHRyffitZu0rKuGH/3UMFyQKBgQDaC0lYZB7wvD6LObLS85Aygv9y6SldnCEt
qzl/6NwSYK/COwd5cf9t4G0Ylku3gxU1KPgCTeyen8LwDVp0z/J0fHoQXJjulhDN
K5O3960vye8xEtqZDMQe57HYb0BluKSvcGcMP90pgoJmSDESNJ9tJrv2VEJUpSKU
XyqsLGYHzwKBgESWash9p3O1fripVW9QZD1lR9QPhTbgSYXOyNr3XY6g0yepQSyP
dQCExKqDfX7uiynPN94S7k3p3shJEEtycADhecO72QnOQVbuKvUINR0dkh9tl81P
Qx11bX6RY3OGSy8DiIxQZ4hSVG+kI/QcNadUZL9GEB3GgaizkRDWjHgJAoGAYK4u
eF30hiPBy7PqwbSzlGIXaFlQOSyYXYqVdUzH//IVHJdV6hiM/KhNV2CU9CrQRYED
7umkaHVIV25kVHU7+UCUUxrryKaLjp2q4yCUDyOHxoeom8JYV6e+aMxzjmb/xrad
SoYqx1QSA84wy/S/WAObxk54FtYd7hIAdtU87GsCgYB8CQLw7WwA5T9XW7evNFr5
kkVWJmGfxORNC7lo0L+TdpdF/OQzuMiNohq0kRy+KqsWH4EU7P6WS3svQkBNR4WY
QTC/ydb3mkaonVWZk2I1FPYyewLklwJ8lqCJ9kQ5EPNXso7NbEbeGe8M+82BSE5K
xdug4Ym572dEWNno+36LRA==
-----END PRIVATE KEY-----
`
	var clientCert = `
-----BEGIN CERTIFICATE-----
MIIDTTCCAjUCFDXIjx65Pg1HFcdJJmv1JjQFzUFkMA0GCSqGSIb3DQEBCwUAMG0x
CzAJBgNVBAYTAlVTMQswCQYDVQQIDAJOWTELMAkGA1UEBwwCTlkxDTALBgNVBAsM
BHJvb3QxFDASBgNVBAMMC0dvb2dsZSBMdGQuMR8wHQYJKoZIhvcNAQkBFhBhZG1p
bkBnb29nbGUuY29tMB4XDTIwMDcwNDAzMjMwMFoXDTIwMDgwMzAzMjMwMFowWTEL
MAkGA1UEBhMCVVMxCzAJBgNVBAgMAk5ZMQswCQYDVQQHDAJOWTERMA8GA1UEAwwI
YXNkZi5jb20xHTAbBgkqhkiG9w0BCQEWDmFkbWluQGFzZGYuY29tMIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAw2elDGUWnazDe0/MXYBXzORfF4xeVskL
fZlanK3vuO1LNWe9it8L+ZIAvHJuKKguUamwJ3STd5Qy34v573TyAdIKUPP4xPSr
p54nJhezqQg4OyPcD2mRWTQ5Fve710M/EEcrwz0QN9dWWNZbQfp2e5KULOFlkVFn
yTQBj6qNa2OaYHBghME4KHGYTishh8TuHZVVx85pVBd5z3LnPIHurSMfkvxCbwBe
ySVl3ZtyNHJDxcOkSTfNR06+70oJv1FjjL/WW5TDXkeHZXq62R4agDhvXNbdBnrm
Lt4nZGl6iE4x2s/cNeub9965FVzq7HRnGTjlxAe68nL9UMevmHAshwIDAQABMA0G
CSqGSIb3DQEBCwUAA4IBAQBrwrqLoo0/26m/2xmI+F6Le7SFZB0wx0/tAgjF5T/L
W2kwLZ2pA/3FRZ+sWfwl2eiM2z3J049Yiyb/opxqPOMQsXJwKUnMlZFb7Fr2Cryf
4l5PSXTrW2itB184aHzx2DCSGHuSbZ8079r4X6JYprMb3ZbSdEeBhOGpol0F+b48
8Nuuz5u3aCiGGo8BHTb+uoKOjaux6n9scuzVBTqgDvdWZl7V/Z8rVf9y738b3m8v
cB1wthuMku55gQSCjLELWUEqkNQUQmvuCragCNY0nv7DM8aUH7ggF/QuFP6g2+M+
Gimb9UORkV8SmPTyJuOXKO33PN5o1P4ixfAy1Z/WsX4K
-----END CERTIFICATE-----
`
	ioutil.WriteFile("server.crt", []byte(serverCert), 0777)
	ioutil.WriteFile("server.key", []byte(serverKey), 0777)
	ioutil.WriteFile("client.crt", []byte(clientCert), 0777)
	ioutil.WriteFile("client.key", []byte(clientKey), 0777)
}
