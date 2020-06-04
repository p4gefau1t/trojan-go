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
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.12.3 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
	github.com/refraction-networking/utls v0.0.0-20200601200209-ada0bb9b38a0
	github.com/smartystreets/goconvey v1.6.4
	github.com/xtaci/smux v1.5.15-0.20200523091831-637399ad4398
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.starlark.net v0.0.0-20200519165436-0aa95694c768 // indirect
	golang.org/x/crypto v0.0.0-20200602180216-279210d13fed
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9
	golang.org/x/sys v0.0.0-20200602225109-6fdc65e7d980
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	v2ray.com/core v0.0.0-20190603071532-16e9d39fff74
)

replace v2ray.com/core => github.com/v2ray/v2ray-core v0.0.0-20200603100350-6b5d2fed91c0
