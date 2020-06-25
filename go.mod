module github.com/p4gefau1t/trojan-go

go 1.14

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.1
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/refraction-networking/utls v0.0.0-20200601200209-ada0bb9b38a0
	github.com/shadowsocks/go-shadowsocks2 v0.1.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/xtaci/smux v1.5.14
	golang.org/x/crypto v0.0.0-20200602180216-279210d13fed
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/sys v0.0.0-20200602225109-6fdc65e7d980 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.24.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0
	v2ray.com/core v0.0.0-20190603071532-16e9d39fff74
)

replace v2ray.com/core => github.com/v2ray/v2ray-core v0.0.0-20200603100350-6b5d2fed91c0
