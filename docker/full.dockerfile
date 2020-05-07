FROM golang:latest AS builder

WORKDIR /
RUN git clone --depth=1 https://github.com/p4gefau1t/trojan-go.git
WORKDIR /trojan-go
RUN mkdir release && go build -tags "full" -ldflags "-s -w" -o release
RUN wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O release/geosite.dat \
&& wget https://github.com/v2ray/geoip/raw/release/geoip.dat -O release/geoip.dat

FROM debian:buster-slim
WORKDIR /
COPY --from=builder /trojan-go/release /trojan-go/

ENTRYPOINT ["/trojan-go/trojan-go", "-config"]
CMD ["/config/config.json"]