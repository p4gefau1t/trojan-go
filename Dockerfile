FROM golang:alpine AS builder
WORKDIR /
ARG REF
RUN apk add git make &&\
    git clone https://github.com/p4gefau1t/trojan-go.git
RUN if [[ -z "${REF}" ]]; then \
        echo "No specific commit provided, use the latest one." \
    ;else \
        echo "Use commit ${REF}" &&\
        cd trojan-go &&\
        git checkout ${REF} \
    ;fi
RUN cd trojan-go &&\
    make &&\
    wget https://github.com/v2fly/domain-list-community/raw/release/dlc.dat -O build/geosite.dat &&\
    wget https://github.com/v2fly/geoip/raw/release/geoip.dat -O build/geoip.dat &&\
    wget https://github.com/v2fly/geoip/raw/release/geoip-only-cn-private.dat -O build/geoip-only-cn-private.dat

FROM alpine
WORKDIR /
RUN apk add --no-cache tzdata ca-certificates
COPY --from=builder /trojan-go/build /usr/local/bin/
COPY --from=builder /trojan-go/example/server.json /etc/trojan-go/config.json

ENTRYPOINT ["/usr/local/bin/trojan-go", "-config"]
CMD ["/etc/trojan-go/config.json"]
