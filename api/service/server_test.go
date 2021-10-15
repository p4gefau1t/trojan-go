package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
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
	user.AddSentTraffic(1234)
	user.AddRecvTraffic(5678)
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
			user.AddSentTraffic(200)
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			user.AddRecvTraffic(300)
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
	common.Must(err)
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

var serverCert = `
-----BEGIN CERTIFICATE-----
MIIC+TCCAeGgAwIBAgIQAZ1MkNXl76ABOPPQ6ci25zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTIwMDkwNjAzMTM1NVoXDTIxMDkwNjAzMTM1
NVowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBAJrgYRlmw0851xp1/OWN4b/gQsKc1TmTaMcN+gDX8w2RWfgOOGymeWDJ
QaTu1G6XHjuvW3sqZGixlRtUJKnmldiwpX0ZY00Ce15fgQHZ85uc3rnFjdkeaYFj
KXN7Xx0QZTQjR5N3W5oVvwKRXe9ATwtOKregSJxCMv8P6OWYH8SwR8GsZnUyvKYR
7JodTXpw7pIL4yNx+QETg537y0TXVFpVt0/H9OoKmY/vsIWVWkKOY4nre9XxNf/p
ABWxYy5n1CKTssLWblJs/lSSPfRxCKUnrBcHwr8ZvLwZSvVktLWr0DnurdfSXOSy
nGvF19q7BpB47ZDTca4V95UtqgfquwUCAwEAAaNLMEkwDgYDVR0PAQH/BAQDAgWg
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFAYDVR0RBA0wC4IJ
bG9jYWxob3N0MA0GCSqGSIb3DQEBCwUAA4IBAQBzaEBQs2bjx0trJxDoKK5xFDUX
mhhVOlparYS04WG3q18r9qfcvXDv3DOmzJDAnSldGmHad/ba6uLDuGEtuIYdMK9u
CpQVaLsNsjIeSika7l0fbQ7XBAJzIHkQHF8dGS3qyzagyCLiRuV2qT5v+p6X4tbp
PY2raoobm5hiscLk540mAAboz+IM1nTGuxD+XUh9znnGJhiKVoNnWhhXLHQK3Lwd
Mct/q+LkMaVHgT/r5LBMbk/jPluvgN0VJ6FnEw1JmotduJd+f80Syp4qccZmupEe
zNXfXCPNcNXeSbAwWnsFeiUrU5YNqPobhaiZXMGnoFb4Cufb57AbNPNDch0x
-----END CERTIFICATE-----
`

var serverKey = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCa4GEZZsNPOdca
dfzljeG/4ELCnNU5k2jHDfoA1/MNkVn4DjhspnlgyUGk7tRulx47r1t7KmRosZUb
VCSp5pXYsKV9GWNNAnteX4EB2fObnN65xY3ZHmmBYylze18dEGU0I0eTd1uaFb8C
kV3vQE8LTiq3oEicQjL/D+jlmB/EsEfBrGZ1MrymEeyaHU16cO6SC+MjcfkBE4Od
+8tE11RaVbdPx/TqCpmP77CFlVpCjmOJ63vV8TX/6QAVsWMuZ9Qik7LC1m5SbP5U
kj30cQilJ6wXB8K/Gby8GUr1ZLS1q9A57q3X0lzkspxrxdfauwaQeO2Q03GuFfeV
LaoH6rsFAgMBAAECggEBAIEbGtZ5+8ZHiTDdunwB0naJFB33bygX4fhNhmK9ojdl
O4K1GAQ6omQ0YSyEi0HFZ8aJX9FEfX9oycuGUSnwtml0l/+48jZ4Iy+AnaJVdeX6
1xA1xxF/cKQTbbJ+3cL0r+jOoBQmI45HInuZgpy3Fy1tc96vFthrtuc49ASw04q5
vIgA+oX6dt7ex7WpXJexqO/9wVsFdiy01gF+e3n2UX5C7F5mm3m0ZJI0A9LCFIim
caLqgqSkFXujw0JurIwLolc/qRn/HG/gfMKBpf1ESpj3dDoZFaRftqtUjzXswD82
eZ1PbfpEZ6iUr4K0scUcDdYrupth6U2tDiPz5y1kduECgYEAwrvqC4ulazroqe+e
LrzftOwg7J3gGyMl+ZTG8Fa4Gd2sAJ0R5kGmcaVU4LW4Ysm5lhgXlKLAGLdPCreS
pruSz1SNgXgYBEnj4Pz0zluQbPmdgGNQ/pOxtI8pr1NGkLwq76M0M8pGHjytIV8N
w1FGikm0Zk4ZSFAFVU5GKCL4hYsCgYEAy5pLi+AejuQuSlR+9aZLwisw3snqGxxe
ECKtPaHAjp/OI43/TGXQihoZJyAYdlDbIIwf03xV14Vv8MPjgtsCFf913YAWSp+y
x1Ul9kGYtVL8QeMcPs1Tb+0BU9VrTDegLNuDNIsxl3pERXIjwotDvQGiTIW7rTY5
SiPOhrlec68CgYAxf/jfVHEJD+FiiRFpigNHhxpba0ozO70Ec0gagcCseoelZEfP
gvKfQsqPkEG9gs+VEqyz0KcJ4VbLP5ycm2OXJkQOHAvm0y2E3GgSKH5O5SifIR/O
hpaOcjHDamSul9ZGMfMsEwe92eicagAinP9UWaXst39/vS+N3qbAvxrzPwKBgQCS
eumLMq0JhKTBGVVWClRK16QLRR1Gb/xEg4473xmYAuTds5VPM5j7IpeiDHdM+BMO
sYFcOAHSUtAcWfJe/I3dobL8ruBaw9ZtjpcHOl5RZejSxkBV9obm6Y6g79SIOyTj
4PHeZZ5CKtbfV6TenC8Z1gkcIMLLdU12R5iYWNjZRQKBgFrFy2jVQHrKVas2Fu+o
HYLaMfoodHq4RWLMf64jSpXkJt8jB1A8vI0ekMe2gTXaldRvinYjuhzU/zJIkWuA
LYIN/nRkP0BLRwfZklUbdO3h1lvvlxM533luvX5mo41Gjg/b2f36yRXTa01Q+QML
NYpAJoagHIeNLGo4aJFwiVsZ
-----END PRIVATE KEY-----
`

var clientKey = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDtp0CreAAU+8K9
rz3e+XYcbf+WgVWOwD1+80g3ixHjUTnBDx4n+rAD9UvVNMfMPBt65vxB3jBFk7BN
eI6kRPSp1Lb3qS1EPfhgyG4kM54xP3QZw4KQY0Qqjqfbi+Pg7XRjyB0q9pYJ8ZMD
F9pg7NreJaJvGmjSrm5RsbQaCTz+bW9ckulaCWGMqOM1uQZ1hKQcCH+QeMyNHYcO
pwqNcq1E47nMTJoApXxftTtbDZJKcxn5wxwNB91N26H8dwXqUh+YK5znhYvsghrf
KRZBfUYEfTln6JHDwPPzsgEZ3jbcMRpkO/FdpOYdV1wWYSmq/4HyGXE0O9UQzVTY
HSpC0MnNAgMBAAECggEAeS2kKwqIOCrbdK8LhEt9LyfjgFG4V46sjLPuKewuldNP
+KIFxWrtH0ePgEpmajxn4rYvAEMUKBYTep0zVo2Wl5ZQKV5JJ5fVszvf9XOggQoS
4CQxyf/jvTN6YdclvgY2J77dKJANl0pnpNcf0fZT75wPBEnaEztAI0XSSMhXIn+d
EPeuCZ5FHCQShVBpggDpvjoAfHpvTIvxjz6U3ojD+yLV4tcBme4+HGDdBX1WC3Lo
8ByL/T/sdNFBEykI/Qh70sf/uqpSMh849iE8U8X4dSpolko5HynxUWSVARWRHCQ7
Zs7FO8Bn1S/cuXQJq7EVuNIU6rVRfuK29jTstG9LfQKBgQD4zskYB6b/igWJ0UAJ
x2Dr41BYqxHknugT81lv4GiOSi2sWeNFi49xIrAOvTfgUR6Qfg73nKKsAUwpt+P7
cGkntE3sEOrb3i5f5CsiEWzkxKiDGvpdF6e4OcRsaCu3IgyuiDrtRqRKwL2a+m0q
ejNGL79Y8AF/gLzN5KstfCENcwKBgQD0hew08qcn2q+OLg0OSBBuhhxD3Y7V65Bf
4G3VCt8YG77sbH+HBs985X8AHr1kIm48aHTDleFOlpLHswDpGvlg+Is0Dfff2R6r
1w3HCvH84nCA4l7gkcCvuR3k74Prgj0vN2c/p/B0GIGfowmNpQ1HItsuSQGiot7B
nC6qTCj7vwKBgH1FeEBuEeoVryYtwgVqamU6RUjvkQm/7G+nFb/biCkkNgzSETkB
xI4c/fHd2VVK4o2zuot3RPw/hv52RQZjGb7Q7G7QMb/UBRtowULc7SvdzE5+ddIL
R/nctAY1CNWjAimaE7lF2RB+LLjsH6zEbC6JedkotkhhJC6yVHGJTwb7AoGAEi53
DsTQKwV2skK4U8yF9EHijiVGPp/CX26nnASv6/H8M0YqAVc/TFEgLVkbyftJaRJ3
RCe71gUaKuEjezG3Qz+X0ioLuUhCoJJgAuHMdno71Ul/toD/69D+6QvqKjPH6t/a
vH/3QBqmYMFVr4OLRjPQSlPBXF9x4sGDMsRw868CgYEA6Y82dggvuwMpqn09JkBh
wLy74mR1IPj7aXz983WfstZHCrCsPOUmi3RamBHeC7udHcfa92X0bp1GK6wlNLN9
WPOZ+zXCv9n7WxUeIvHTS9d7OnVkxp2qkUwnC1mmxH40HTQ1c2rUHCxWLe+Qrwi6
X7nFPTe00Vd/5OGXkkv3JL4=
-----END PRIVATE KEY-----
`

var clientCert = `
-----BEGIN CERTIFICATE-----
MIIC+jCCAeKgAwIBAgIRAJCepdWc3B+X+BMsoK9nJlAwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0yMDA5MDYwMzIyMzdaFw0yMTA5MDYwMzIy
MzdaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDtp0CreAAU+8K9rz3e+XYcbf+WgVWOwD1+80g3ixHjUTnBDx4n+rAD
9UvVNMfMPBt65vxB3jBFk7BNeI6kRPSp1Lb3qS1EPfhgyG4kM54xP3QZw4KQY0Qq
jqfbi+Pg7XRjyB0q9pYJ8ZMDF9pg7NreJaJvGmjSrm5RsbQaCTz+bW9ckulaCWGM
qOM1uQZ1hKQcCH+QeMyNHYcOpwqNcq1E47nMTJoApXxftTtbDZJKcxn5wxwNB91N
26H8dwXqUh+YK5znhYvsghrfKRZBfUYEfTln6JHDwPPzsgEZ3jbcMRpkO/FdpOYd
V1wWYSmq/4HyGXE0O9UQzVTYHSpC0MnNAgMBAAGjSzBJMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMBQGA1UdEQQNMAuC
CWxvY2FsaG9zdDANBgkqhkiG9w0BAQsFAAOCAQEAz87Xx7OAja/YZ1jqIk7YOaE9
l9vfpn9MY+3XaasjJga2QOKZaL2nCA15mUNP69M1awoQc0DKdsyTBS0yU44lbgrV
8ibpxMksiOhBoRr8ig9vQmFyTPFODK0FgSXx9Ek9IJqF+4/0ggOiRD9o+lrr9amJ
A/U291tzkuMc0nalNRFZFJJbSeap+NdNLWEGTbH08Dg/e9/p16lYvq4Th2mXryMz
wDxdHr2KFJp+qMbWF2WHIAUrCBr7gTW5BQElnVTyIihUOTAUCrEfFEj4uN3UXwM5
qbPPrmQPgv5prRHCObn0+j6SwV9vV7Q9BI41CloKUDXZmPFTVipP6z5tV2YTOg==
-----END CERTIFICATE-----

`

func init() {
	ioutil.WriteFile("server.crt", []byte(serverCert), 0o777)
	ioutil.WriteFile("server.key", []byte(serverKey), 0o777)
	ioutil.WriteFile("client.crt", []byte(clientCert), 0o777)
	ioutil.WriteFile("client.key", []byte(clientKey), 0o777)
}
