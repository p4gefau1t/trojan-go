FROM golang:alpine AS builder
WORKDIR /
RUN apk add --update git &&\
    git clone --depth=1 https://github.com/p4gefau1t/trojan-go.git &&\
    cd trojan-go && mkdir release && go build -tags "full" -ldflags "-s -w" -o release &&\
    wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O release/geosite.dat &&\
    wget https://github.com/v2ray/geoip/raw/release/geoip.dat -O release/geoip.dat

FROM alpine
WORKDIR /
COPY --from=builder /trojan-go/release /usr/local/bin/
COPY example/server.json /etc/trojan-go/config.json


ENTRYPOINT ["/usr/local/bin/trojan-go", "-config"]
<<<<<<< HEAD
CMD ["/etc/trojan-go/server.json"]
=======
CMD ["/etc/trojan-go/config.json"]
>>>>>>> 3005f4c60a79faf2f7024a1d67fc75328cbdf433
