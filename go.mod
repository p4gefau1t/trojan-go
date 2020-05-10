module github.com/p4gefau1t/trojan-go

go 1.14

require (
	github.com/LiamHaworth/go-tproxy v0.0.0-20190726054950-ef7efd7f24ed
	github.com/cenkalti/backoff/v4 v4.0.2 // indirect
	github.com/go-acme/lego/v3 v3.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.1
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mediocregopher/radix/v3 v3.5.0
	github.com/miekg/dns v1.1.29 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
	github.com/refraction-networking/utls v0.0.0-20190909200633-43c36d3c1f57
	github.com/smartystreets/goconvey v1.6.4
	github.com/xtaci/smux v1.5.12
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.starlark.net v0.0.0-20200330013621-be5394c419b6 // indirect
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f
	golang.org/x/sys v0.0.0-20200509044756-6aff5f38e54f
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	google.golang.org/genproto v0.0.0-20200507105951-43844f6eee31 // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	v2ray.com/core v4.19.1+incompatible
)

replace v2ray.com/core => github.com/v2ray/v2ray-core v4.23.1+incompatible
