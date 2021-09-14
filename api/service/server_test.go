package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
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
	time.Sleep(time.Second * 1)
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

	stream3, err := server.SetUsers(ctx)
	common.Must(err)
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
		common.Must(err)
		fmt.Println(resp2.Status.SpeedCurrent)
		fmt.Println(resp2.Status.SpeedLimit)
		time.Sleep(time.Second)
	}
	stream2.CloseSend()
	cancel()
}

func TestTLSRSA(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	cfg := &Config{
		API: APIConfig{
			Enabled: true,
			APIHost: "127.0.0.1",
			APIPort: port,
			SSL: SSLConfig{
				Enabled:        true,
				CertPath:       "server-rsa2048.crt",
				KeyPath:        "server-rsa2048.key",
				VerifyClient:   false,
				ClientCertPath: []string{"client-rsa2048.crt"},
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
	certBytes, err := os.ReadFile("server-rsa2048.crt")
	common.Must(err)
	pool.AppendCertsFromPEM(certBytes)

	certificate, err := tls.LoadX509KeyPair("client-rsa2048.crt", "client-rsa2048.key")
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

func TestTLSECC(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	cfg := &Config{
		API: APIConfig{
			Enabled: true,
			APIHost: "127.0.0.1",
			APIPort: port,
			SSL: SSLConfig{
				Enabled:        true,
				CertPath:       "server-ecc.crt",
				KeyPath:        "server-ecc.key",
				VerifyClient:   false,
				ClientCertPath: []string{"client-ecc.crt"},
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
	certBytes, err := os.ReadFile("server-ecc.crt")
	common.Must(err)
	pool.AppendCertsFromPEM(certBytes)

	certificate, err := tls.LoadX509KeyPair("client-ecc.crt", "client-ecc.key")
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

var serverRSA2048Cert = `
-----BEGIN CERTIFICATE-----
MIIC5TCCAc2gAwIBAgIJAJqNVe6g/10vMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0yMTA5MTQwNjE1MTFaFw0yNjA5MTMwNjE1MTFaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAK7bupJ8tmHM3shQ/7N730jzpRsXdNiBxq/Jxx8j+vB3AcxuP5bjXQZqS6YR
5W5vrfLlegtq1E/mmaI3Ht0RfIlzev04Dua9PWmIQJD801nEPknbfgCLXDh+pYr2
sfg8mUh3LjGtrxyH+nmbTjWg7iWSKohmZ8nUDcX94Llo5FxibMAz8OsAwOmUueCH
jP3XswZYHEy+OOP3K0ZEiJy0f5T6ZXk9OWYuPN4VQKJx1qrc9KzZtSPHwqVdkGUi
ase9tOPA4aMutzt0btgW7h7UrvG6C1c/Rr1BxdiYq1EQ+yypnAlyToVQSNbo67zz
wGQk4GeruIkOgJOLdooN/HjhbHMCAwEAAaM6MDgwFAYDVR0RBA0wC4IJbG9jYWxo
b3N0MAsGA1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATANBgkqhkiG9w0B
AQsFAAOCAQEASsBzHHYiWDDiBVWUEwVZAduTrslTLNOxG0QHBKsHWIlz/3QlhQil
ywb3OhfMTUR1dMGY5Iq5432QiCHO4IMCOv7tDIkgb4Bc3v/3CRlBlnurtAmUfNJ6
pTRSlK4AjWpGHAEEd/8aCaOE86hMP8WDht8MkJTRrQqpJ1HeDISoKt9nepHOIsj+
I2zLZZtw0pg7FuR4MzWuqOt071iRS46Pupryb3ZEGIWNz5iLrDQod5Iz2ZGSRGqE
rB8idX0mlj5AHRRanVR3PAes+eApsW9JvYG/ImuCOs+ZsukY614zQZdR+SyFm85G
4NICyeQsmiypNHHgw+xZmGqZg65bXNGoyg==
-----END CERTIFICATE-----
`

var serverRSA2048Key = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCu27qSfLZhzN7I
UP+ze99I86UbF3TYgcavyccfI/rwdwHMbj+W410GakumEeVub63y5XoLatRP5pmi
Nx7dEXyJc3r9OA7mvT1piECQ/NNZxD5J234Ai1w4fqWK9rH4PJlIdy4xra8ch/p5
m041oO4lkiqIZmfJ1A3F/eC5aORcYmzAM/DrAMDplLngh4z917MGWBxMvjjj9ytG
RIictH+U+mV5PTlmLjzeFUCicdaq3PSs2bUjx8KlXZBlImrHvbTjwOGjLrc7dG7Y
Fu4e1K7xugtXP0a9QcXYmKtREPssqZwJck6FUEjW6Ou888BkJOBnq7iJDoCTi3aK
Dfx44WxzAgMBAAECggEBAKYhib/H0ZhWB4yWuHqUxG4RXtrAjHlvw5Acy5zgmHiC
+Sh7ztrTJf0EXN9pvWwRm1ldgXj7hMBtPaaLbD1pccM9/qo66p17Sq/LjlyyeTOe
affOHIbz4Sij2zCOdkR9fr0EztTQScF3yBhl4Aa/4cO8fcCeWxm86WEldq9x4xWJ
s5WMR4CnrOJhDINLNPQPKX92KyxEQ/RfuBWovx3M0nl3fcUWfESY134t5g/UBFId
In19tZ+pGIpCkxP0U1AZWrlZRA8Q/3sO2orUpoAOdCrGk/DcCTMh0c1pMzbYZ1/i
cYXn38MpUo8QeG4FElUhAv6kzeBIl2tRBMVzIigo+AECgYEA3No1rHdFu6Ox9vC8
E93PTZevYVcL5J5yx6x7khCaOLKKuRXpjOX/h3Ll+hlN2DVAg5Jli/JVGCco4GeK
kbFLSyxG1+E63JbgsVpaEOgvFT3bHHSPSRJDnIU+WkcNQ2u4Ky5ahZzbNdV+4fj2
NO2iMgkm7hoJANrm3IqqW8epenMCgYEAyq+qdNj5DiDzBcDvLwY+4/QmMOOgDqeh
/TzhbDRyr+m4xNT7LLS4s/3wcbkQC33zhMUI3YvOHnYq5Ze/iL/TSloj0QCp1I7L
J7sZeM1XimMBQIpCfOC7lf4tU76Fz0DTHAL+CmX1DgmRJdYO09843VsKkscC968R
4cwL5oGxxgECgYAM4TTsH/CTJtLEIfn19qOWVNhHhvoMlSkAeBCkzg8Qa2knrh12
uBsU3SCIW11s1H40rh758GICDJaXr7InGP3ZHnXrNRlnr+zeqvRBtCi6xma23B1X
F5eV0zd1sFsXqXqOGh/xVtp54z+JEinZoForLNl2XVJVGG8KQZP50kUR/QKBgH4O
8zzpFT0sUPlrHVdp0wODfZ06dPmoWJ9flfPuSsYN3tTMgcs0Owv3C+wu5UPAegxB
X1oq8W8Qn21cC8vJQmgj19LNTtLcXI3BV/5B+Aghu02gr+lq/EA1bYuAG0jjUGlD
kyx0bQzl9lhJ4b70PjGtxc2z6KyTPdPpTB143FABAoGAQDoIUdc77/IWcjzcaXeJ
8abak5rAZA7cu2g2NVfs+Km+njsB0pbTwMnV1zGoFABdaHLdqbthLWtX7WOb1PDD
MQ+kbiLw5uj8IY2HEqJhDGGEdXBqxbW7kyuIAN9Mw+mwKzkikNcFQdxgchWH1d1o
lVkr92iEX+IhIeYb4DN1vQw=
-----END PRIVATE KEY-----
`

var clientRSA2048Cert = `
-----BEGIN CERTIFICATE-----
MIIC5TCCAc2gAwIBAgIJAKD1wSl+Mnk7MA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0yMTA5MTQwNjE2MDBaFw0yNjA5MTMwNjE2MDBaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAJeDu630louuf2V4sw396cGiAnxmTseVRMG+m3PnZ831puAsApm3IWSEcOqI
UMk6s1pgSLysg6GxRZhX4L/ljErjMO4+y8riZjqqR0wd0GnhuNxuXaSUsEmKdDZb
cICqXkZeZRn4jw/7L0xgdAdM2w3LXR6aq6CwveFY3/JncEZFQHH5mnorZdpbheR6
rhvIL6AAI0YEY9uzuQBSrzOml3f7D+x5Xll14HoMN0kCysWt8jSP/An5yP8pL5RO
pn5kNBc8Bx8lykuV1uS8ogncSM7JzmpP1SeAViOq8CqXlJtUbUqVPckMmdfMMtbI
qIO7R5/8imrdhLMi25fAOnfmDzcCAwEAAaM6MDgwFAYDVR0RBA0wC4IJbG9jYWxo
b3N0MAsGA1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATANBgkqhkiG9w0B
AQsFAAOCAQEAFu2QPE3x1Sm3SfnHzAhvdjviYkbWvM8rQziIlIevbvA9Nl+vxDBf
N5aRR6Hpxq02J2G/w7tzrKB9IluWdMU1+tilph5bCnwx3QUh/GR4oTsFiTvTZ5br
SNf3xfTyIsL+Hf6iLvEgSt15ziY/334wu9NmQrU0FNZ+Lcc7Mx0OgvuP9Zim+6oo
/FW80R3pUSzUZcUQgsI4Sz7/6nJTxhsc+kqtnOXIQLPC9GA06kP8eN6XjTsavP7f
eZq/yozddOk0dqx8uwmKUOb1Rg+pS8VIhQBRv3UPb4L/07AWSTZSMZLf1+CMgzMY
Jtsxa1MLqPkB7fiAR6SFUFW7Q36gDp/Mdw==
-----END CERTIFICATE-----
`

var clientRSA2048Key = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCXg7ut9JaLrn9l
eLMN/enBogJ8Zk7HlUTBvptz52fN9abgLAKZtyFkhHDqiFDJOrNaYEi8rIOhsUWY
V+C/5YxK4zDuPsvK4mY6qkdMHdBp4bjcbl2klLBJinQ2W3CAql5GXmUZ+I8P+y9M
YHQHTNsNy10emqugsL3hWN/yZ3BGRUBx+Zp6K2XaW4Xkeq4byC+gACNGBGPbs7kA
Uq8zppd3+w/seV5ZdeB6DDdJAsrFrfI0j/wJ+cj/KS+UTqZ+ZDQXPAcfJcpLldbk
vKIJ3EjOyc5qT9UngFYjqvAql5SbVG1KlT3JDJnXzDLWyKiDu0ef/Ipq3YSzItuX
wDp35g83AgMBAAECggEASOv4CjMrubKUUgwTcWqBdNY6iBDdXaVz4COSweffx/qx
BDdqUP0Yrz4m8loFN7Ru2dJ5b4VAHTQqoLW6z+D08p4B0MicYNsyBI4rnnDC/BLN
XBoqK6n8Zoiigf7kWKimkwufcS51/GUSUJojfdf5ndwAx1f9vmsSGEEkF5C9MrQo
Sa4eyySuXuDS9swrXuTn9FcVcbUoIblgL8GIlmX4S1Xl9NaVS/VAVt42FgpNSYy7
2Qf9Medg3ApimjwkLiDSh0RElirlUwzSg9dx2U+hHwWQWimb2AhA85uK3bpFB/x9
b2agS1uxTar4mk4LFppQqVUuXlpj2hW7HxTdcxGHYQKBgQDGA+Wv3b8FMeszYqqx
BbI5+FeXmQ5AoYNdHyPfCH8f2LX1FTnnQbUvFMJ3UQZl/GGHocgyFNvfDo6YyYsg
2XgcNO/JWMbKEw0HkfMgkaIa3Jfq/PTB386NhqBq5FHiBUrvSHzhaIzpuaokBrRk
jFlcqONK+uj77iRgci4wR59POQKBgQDD4fDZzOGQCy/AWrMX/adk8ymoxhWQj6PN
zhy2GZo9jGwyskQr5neIDQtxRbgokMspxpyPdQG2SxbG/zygyFEaCsp/Xb3iB3aA
2dcktkV9agx+g1WrflTpGG9quW5vQJgQv9FtlVXJdiihN9CQvXAcYRjUA/zkadIL
GsrVF9rh7wKBgFoMLaiDU7neEJKGnQ7xgzI/kD29ebDEgkOXxK1JZN4ro9t3MqTK
ycVGUIUIELvSQNv4I107BR3ztb8fcCiZHLjfDehnecctULCPm5vE/o3uoRtYu0lr
KLhNb6gMenwpYgFc2oV7ERG8v/WwItrSxFSR7QMNBWSD0IEXi4+jEnxpAoGAMJTS
5VG5B76egzh7fpG8eH8Ob/tg0c+uMpbR7CABbw5qr1AjNDgeoTGLCvbdq8HtgVju
7213lTyeU5Bt+vpzkt/mRRx8wZhUPbTJdSN3rJkmrCHql3Pnn0AeMfv3dcQxcsYA
LQuCkUqq3QE4yw0QxxkVzU+H4yaTn4lvkNYvxSUCgYEAgD85MM3DaiNcrevQv6Wb
vSo7jBhd8Q/NawH53V32eIlF9Kkm2mqTmPIQ2OxOtdZ3xm+JmPd20jEVg7NcPL58
t6BoSH10pLaI0CP2NdcYfJTIiHAhuQe7PxPCA0nlHTXgSv97QR7y4qMhbf2j1umc
hKym0tdk2QjR3qqglaD572A=
-----END PRIVATE KEY-----
`

var serverECCCert = `
-----BEGIN CERTIFICATE-----
MIICTDCCAfKgAwIBAgIQDtCrO8cNST2eY2tA/AGrsDAKBggqhkjOPQQDAjBeMQsw
CQYDVQQGEwJDTjEOMAwGA1UEChMFTXlTU0wxKzApBgNVBAsTIk15U1NMIFRlc3Qg
RUNDIC0gRm9yIHRlc3QgdXNlIG9ubHkxEjAQBgNVBAMTCU15U1NMLmNvbTAeFw0y
MTA5MTQwNjQ1MzNaFw0yNjA5MTMwNjQ1MzNaMCExCzAJBgNVBAYTAkNOMRIwEAYD
VQQDEwlsb2NhbGhvc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASvYy/r7XR1
Y39lC2JpRJh582zR2CTNynbuolK9a1jsbXaZv+hpBlHkgzMHsWu7LY9Pnb/Dbp4i
1lRASOddD/rLo4HOMIHLMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEF
BQcDAQYIKwYBBQUHAwIwHwYDVR0jBBgwFoAUWxGyVxD0fBhTy3tH4eKznRFXFCYw
YwYIKwYBBQUHAQEEVzBVMCEGCCsGAQUFBzABhhVodHRwOi8vb2NzcC5teXNzbC5j
b20wMAYIKwYBBQUHMAKGJGh0dHA6Ly9jYS5teXNzbC5jb20vbXlzc2x0ZXN0ZWNj
LmNydDAUBgNVHREEDTALgglsb2NhbGhvc3QwCgYIKoZIzj0EAwIDSAAwRQIgDQUa
GEdmKstLMHUmmPMGm/P9S4vvSZV2VHsb3+AEyIUCIQCdJpbyTCz+mEyskhwrGOw/
blh3WBONv6MBtqPpmgE1AQ==
-----END CERTIFICATE-----
`

var serverECCKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIB8G2suYKuBLoodNIwRMp3JPN1fcZxCt3kcOYIx4nbcPoAoGCCqGSM49
AwEHoUQDQgAEr2Mv6+10dWN/ZQtiaUSYefNs0dgkzcp27qJSvWtY7G12mb/oaQZR
5IMzB7Fruy2PT52/w26eItZUQEjnXQ/6yw==
-----END EC PRIVATE KEY-----
`

var clientECCCert = `
-----BEGIN CERTIFICATE-----
MIICTDCCAfKgAwIBAgIQb5FLuCggTiWFtt/dPh+2bTAKBggqhkjOPQQDAjBeMQsw
CQYDVQQGEwJDTjEOMAwGA1UEChMFTXlTU0wxKzApBgNVBAsTIk15U1NMIFRlc3Qg
RUNDIC0gRm9yIHRlc3QgdXNlIG9ubHkxEjAQBgNVBAMTCU15U1NMLmNvbTAeFw0y
MTA5MTQwNjQ2MTFaFw0yNjA5MTMwNjQ2MTFaMCExCzAJBgNVBAYTAkNOMRIwEAYD
VQQDEwlsb2NhbGhvc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQME+DnYtcK
lbcZmc33oEtoeRWH61DYdl4ei/bM+vkv01MkBB+YTZl0yofJIFJYsfU5pMFK+uyw
D4qdcklKPGIKo4HOMIHLMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEF
BQcDAQYIKwYBBQUHAwIwHwYDVR0jBBgwFoAUWxGyVxD0fBhTy3tH4eKznRFXFCYw
YwYIKwYBBQUHAQEEVzBVMCEGCCsGAQUFBzABhhVodHRwOi8vb2NzcC5teXNzbC5j
b20wMAYIKwYBBQUHMAKGJGh0dHA6Ly9jYS5teXNzbC5jb20vbXlzc2x0ZXN0ZWNj
LmNydDAUBgNVHREEDTALgglsb2NhbGhvc3QwCgYIKoZIzj0EAwIDSAAwRQIgfjaU
sCfgklPsjHzs3fUtSRGfWoRLFsRBO66RtHJSzrYCIQDAxgx0FB0mAXbflj4mXJVA
9/SjaCI40D6MhMnJhQS7Zg==
-----END CERTIFICATE-----
`

var clientECCKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEINgbap6WOGXm8V+ghIyporGcGWnggbjjP9xNsxhn+0sqoAoGCCqGSM49
AwEHoUQDQgAEDBPg52LXCpW3GZnN96BLaHkVh+tQ2HZeHov2zPr5L9NTJAQfmE2Z
dMqHySBSWLH1OaTBSvrssA+KnXJJSjxiCg==
-----END EC PRIVATE KEY-----
`

func init() {
	os.WriteFile("server-rsa2048.crt", []byte(serverRSA2048Cert), 0o777)
	os.WriteFile("server-rsa2048.key", []byte(serverRSA2048Key), 0o777)
	os.WriteFile("client-rsa2048.crt", []byte(clientRSA2048Cert), 0o777)
	os.WriteFile("client-rsa2048.key", []byte(clientRSA2048Key), 0o777)

	os.WriteFile("server-ecc.crt", []byte(serverECCCert), 0o777)
	os.WriteFile("server-ecc.key", []byte(serverECCKey), 0o777)
	os.WriteFile("client-ecc.crt", []byte(clientECCCert), 0o777)
	os.WriteFile("client-ecc.key", []byte(clientECCKey), 0o777)
}
