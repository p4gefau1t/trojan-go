module github.com/p4gefau1t/trojan-go

go 1.14

require (
	github.com/LiamHaworth/go-tproxy v0.0.0-20190726054950-ef7efd7f24ed
	github.com/cenkalti/backoff/v4 v4.0.2 // indirect
	github.com/go-acme/lego/v3 v3.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mediocregopher/radix/v3 v3.5.1
	github.com/miekg/dns v1.1.29 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
	github.com/refraction-networking/utls v0.0.0-20190909200633-43c36d3c1f57
	github.com/smartystreets/goconvey v1.6.4
	github.com/xtaci/smux v1.5.15-0.20200523091831-637399ad4398
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.starlark.net v0.0.0-20200519165436-0aa95694c768 // indirect
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	golang.org/x/sys v0.0.0-20200523222454-059865788121
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	v2ray.com/core v4.19.1+incompatible
)

replace v2ray.com/core => github.com/v2ray/v2ray-core v4.23.4+incompatible
