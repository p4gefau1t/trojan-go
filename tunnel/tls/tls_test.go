package tls

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel/freedom"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
)

var rsa2048Cert = `
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

var rsa2048Key = `
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

var eccCert = `
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

var eccKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIB8G2suYKuBLoodNIwRMp3JPN1fcZxCt3kcOYIx4nbcPoAoGCCqGSM49
AwEHoUQDQgAEr2Mv6+10dWN/ZQtiaUSYefNs0dgkzcp27qJSvWtY7G12mb/oaQZR
5IMzB7Fruy2PT52/w26eItZUQEjnXQ/6yw==
-----END EC PRIVATE KEY-----
`

func TestDefaultTLSRSA2048(t *testing.T) {
	os.WriteFile("server-rsa2048.crt", []byte(rsa2048Cert), 0o777)
	os.WriteFile("server-rsa2048.key", []byte(rsa2048Key), 0o777)
	serverCfg := &Config{
		TLS: TLSConfig{
			VerifyHostName: true,
			CertCheckRate:  1,
			KeyPath:        "server-rsa2048.key",
			CertPath:       "server-rsa2048.crt",
		},
	}
	clientCfg := &Config{
		TLS: TLSConfig{
			Verify:      false,
			SNI:         "localhost",
			Fingerprint: "",
		},
	}
	sctx := config.WithConfig(context.Background(), Name, serverCfg)
	cctx := config.WithConfig(context.Background(), Name, clientCfg)

	port := common.PickPort("tcp", "127.0.0.1")
	transportConfig := &transport.Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  port,
		RemoteHost: "127.0.0.1",
		RemotePort: port,
	}
	ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
	ctx = config.WithConfig(ctx, freedom.Name, &freedom.Config{})
	tcpClient, err := transport.NewClient(ctx, nil)
	common.Must(err)
	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)
	common.Must(err)
	s, err := NewServer(sctx, tcpServer)
	common.Must(err)
	c, err := NewClient(cctx, tcpClient)
	common.Must(err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	var conn1, conn2 net.Conn
	go func() {
		conn2, err = s.AcceptConn(nil)
		common.Must(err)
		wg.Done()
	}()
	conn1, err = c.DialConn(nil, nil)
	common.Must(err)

	common.Must2(conn1.Write([]byte("12345678\r\n")))
	wg.Wait()
	buf := [10]byte{}
	conn2.Read(buf[:])
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
	conn1.Close()
	conn2.Close()
}

func TestDefaultTLSECC(t *testing.T) {
	os.WriteFile("server-ecc.crt", []byte(eccCert), 0o777)
	os.WriteFile("server-ecc.key", []byte(eccKey), 0o777)
	serverCfg := &Config{
		TLS: TLSConfig{
			VerifyHostName: true,
			CertCheckRate:  1,
			KeyPath:        "server-ecc.key",
			CertPath:       "server-ecc.crt",
		},
	}
	clientCfg := &Config{
		TLS: TLSConfig{
			Verify:      false,
			SNI:         "localhost",
			Fingerprint: "",
		},
	}
	sctx := config.WithConfig(context.Background(), Name, serverCfg)
	cctx := config.WithConfig(context.Background(), Name, clientCfg)

	port := common.PickPort("tcp", "127.0.0.1")
	transportConfig := &transport.Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  port,
		RemoteHost: "127.0.0.1",
		RemotePort: port,
	}
	ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
	ctx = config.WithConfig(ctx, freedom.Name, &freedom.Config{})
	tcpClient, err := transport.NewClient(ctx, nil)
	common.Must(err)
	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)
	common.Must(err)
	s, err := NewServer(sctx, tcpServer)
	common.Must(err)
	c, err := NewClient(cctx, tcpClient)
	common.Must(err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	var conn1, conn2 net.Conn
	go func() {
		conn2, err = s.AcceptConn(nil)
		common.Must(err)
		wg.Done()
	}()
	conn1, err = c.DialConn(nil, nil)
	common.Must(err)

	common.Must2(conn1.Write([]byte("12345678\r\n")))
	wg.Wait()
	buf := [10]byte{}
	conn2.Read(buf[:])
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
	conn1.Close()
	conn2.Close()
}

func TestUTLSRSA2048(t *testing.T) {
	os.WriteFile("server-rsa2048.crt", []byte(rsa2048Cert), 0o777)
	os.WriteFile("server-rsa2048.key", []byte(rsa2048Key), 0o777)
	fingerprints := []string{
		"chrome",
		"firefox",
		"ios",
	}
	for _, s := range fingerprints {
		serverCfg := &Config{
			TLS: TLSConfig{
				CertCheckRate: 1,
				KeyPath:       "server-rsa2048.key",
				CertPath:      "server-rsa2048.crt",
			},
		}
		clientCfg := &Config{
			TLS: TLSConfig{
				Verify:      false,
				SNI:         "localhost",
				Fingerprint: s,
			},
		}
		sctx := config.WithConfig(context.Background(), Name, serverCfg)
		cctx := config.WithConfig(context.Background(), Name, clientCfg)

		port := common.PickPort("tcp", "127.0.0.1")
		transportConfig := &transport.Config{
			LocalHost:  "127.0.0.1",
			LocalPort:  port,
			RemoteHost: "127.0.0.1",
			RemotePort: port,
		}
		ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
		ctx = config.WithConfig(ctx, freedom.Name, &freedom.Config{})
		tcpClient, err := transport.NewClient(ctx, nil)
		common.Must(err)
		tcpServer, err := transport.NewServer(ctx, nil)
		common.Must(err)

		s, err := NewServer(sctx, tcpServer)
		common.Must(err)
		c, err := NewClient(cctx, tcpClient)
		common.Must(err)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var conn1, conn2 net.Conn
		go func() {
			conn2, err = s.AcceptConn(nil)
			common.Must(err)
			wg.Done()
		}()
		conn1, err = c.DialConn(nil, nil)
		common.Must(err)

		common.Must2(conn1.Write([]byte("12345678\r\n")))
		wg.Wait()
		buf := [10]byte{}
		conn2.Read(buf[:])
		if !util.CheckConn(conn1, conn2) {
			t.Fail()
		}
		conn1.Close()
		conn2.Close()
		s.Close()
		c.Close()
	}
}

func TestUTLSECC(t *testing.T) {
	os.WriteFile("server-ecc.crt", []byte(eccCert), 0o777)
	os.WriteFile("server-ecc.key", []byte(eccKey), 0o777)
	fingerprints := []string{
		"chrome",
		"firefox",
		"ios",
	}
	for _, s := range fingerprints {
		serverCfg := &Config{
			TLS: TLSConfig{
				CertCheckRate: 1,
				KeyPath:       "server-ecc.key",
				CertPath:      "server-ecc.crt",
			},
		}
		clientCfg := &Config{
			TLS: TLSConfig{
				Verify:      false,
				SNI:         "localhost",
				Fingerprint: s,
			},
		}
		sctx := config.WithConfig(context.Background(), Name, serverCfg)
		cctx := config.WithConfig(context.Background(), Name, clientCfg)

		port := common.PickPort("tcp", "127.0.0.1")
		transportConfig := &transport.Config{
			LocalHost:  "127.0.0.1",
			LocalPort:  port,
			RemoteHost: "127.0.0.1",
			RemotePort: port,
		}
		ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
		ctx = config.WithConfig(ctx, freedom.Name, &freedom.Config{})
		tcpClient, err := transport.NewClient(ctx, nil)
		common.Must(err)
		tcpServer, err := transport.NewServer(ctx, nil)
		common.Must(err)

		s, err := NewServer(sctx, tcpServer)
		common.Must(err)
		c, err := NewClient(cctx, tcpClient)
		common.Must(err)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var conn1, conn2 net.Conn
		go func() {
			conn2, err = s.AcceptConn(nil)
			common.Must(err)
			wg.Done()
		}()
		conn1, err = c.DialConn(nil, nil)
		common.Must(err)

		common.Must2(conn1.Write([]byte("12345678\r\n")))
		wg.Wait()
		buf := [10]byte{}
		conn2.Read(buf[:])
		if !util.CheckConn(conn1, conn2) {
			t.Fail()
		}
		conn1.Close()
		conn2.Close()
		s.Close()
		c.Close()
	}
}

func TestMatch(t *testing.T) {
	if !isDomainNameMatched("*.google.com", "www.google.com") {
		t.Fail()
	}

	if isDomainNameMatched("*.google.com", "google.com") {
		t.Fail()
	}

	if !isDomainNameMatched("localhost", "localhost") {
		t.Fail()
	}
}
