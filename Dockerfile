FROM golang:alpine AS builder
WORKDIR /
RUN apk add git make &&\
    git clone https://github.com/p4gefau1t/trojan-go.git &&\
    cd trojan-go &&\
    make &&\
    wget https://github.com/v2fly/domain-list-community/raw/release/dlc.dat -O build/geosite.dat &&\
    wget https://github.com/v2fly/geoip/raw/release/geoip.dat -O build/geoip.dat

FROM alpine
WORKDIR /
RUN apk add --no-cache tzdata ca-certificates
COPY --from=builder /trojan-go/build /usr/local/bin/
COPY --from=builder /trojan-go/example/server.json /etc/trojan-go/config.json

ENTRYPOINT ["/usr/local/bin/trojan-go", "-config"]
CMD ["/etc/trojan-go/config.json"]
