FROM golang:alpine AS builder
WORKDIR /
RUN apk add git make &&\
    git clone https://github.com/p4gefau1t/trojan-go.git &&\
    cd trojan-go &&\
    make &&\
    wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O bin/geosite.dat &&\
    wget https://github.com/v2ray/geoip/raw/release/geoip.dat -O bin/geoip.dat

FROM alpine
WORKDIR /
COPY --from=builder /trojan-go/bin /usr/local/bin/
COPY --from=builder /trojan-go/example/server.json /etc/trojan-go/config.json

ENTRYPOINT ["/usr/local/bin/trojan-go", "-config"]
CMD ["/etc/trojan-go/config.json"]