NAME := trojan-go
PACKAGE_NAME := github.com/p4gefau1t/trojan-go
VERSION := `echo git describe --dirty`
COMMIT := `echo git rev-parse HEAD`

GOPATH ?=
ifneq ($(strip $(GOPATH)), )
	GO_DIR := $(GOPATH)/bin/
endif

MAKEDEPEND = $(GODIR)go
GO_MINIMUM := go1.14
GO_VERSION != $(GO_DIR)go version | cut -d' ' -f3
GOMETALINTER := $(GO_DIR)gometalinter

PLATFORM := linux
BIN_DIR := bin
VAR_SETTING := -X $(PACKAGE_NAME)/constant.Version=$(VERSION) -X $(PACKAGE_NAME)/constant.Commit=$(COMMIT)
GOBUILD = env CGO_ENABLED=0 $(GO_DIR)go build -tags "full" -ldflags="-s -w $(VAR_SETTING)" -o $(BIN_DIR)

.PHONY: depends trojan-go release
normal: clean trojan-go

depends: geosite.dat geoip.dat
	$(info GO_DIR: $(GO_DIR))
	$(info Current Go Verison: $(GO_VERSION))
ifneq ($(GO_VERSION), $(lastword $(sort $(GO_MINIMUM) $(GO_VERSION))))
	$(error Requires $(GO_MINIMUM) for module fingerprint/tls)
endif

realclean: distclean
	rm -f *.dat

distclean: clean
	rm -f *.zip

clean:
	rm -rf $(BIN_DIR)

geoip.dat:
	wget https://github.com/v2ray/geoip/raw/release/geoip.dat

geosite.dat:
	wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O geosite.dat

lint:
	$(GO_DIR)go get -u github.com/alecthomas/gometalinter
	$(GOMETALINTER) --install &> /dev/null
	$(GOMETALINTER) ./... --vendor

test: depends
	@$(GO_DIR)go test ./...

trojan-go: depends
	mkdir -p $(BIN_DIR)
	$(GOBUILD)

install: $(BIN_DIR)/$(NAME) geoip.dat geosite.dat
	mkdir -p /etc/$(NAME)
	mkdir -p /usr/share/$(NAME)
	cp example/*.json /etc/$(NAME)
	cp $(BIN_DIR)/$(NAME) /usr/bin/$(NAME)
	cp example/$(NAME).service /usr/lib/systemd/system/
	cp example/$(NAME)@.service /usr/lib/systemd/system/
	systemctl daemon-reload
	cp geosite.dat /usr/share/$(NAME)/geosite.dat
	cp geoip.dat /usr/share/$(NAME)/geoip.dat
	ln -fs /usr/share/$(NAME)/geoip.dat /usr/bin/
	ln -fs /usr/share/$(NAME)/geosite.dat /usr/bin/

uninstall:
	rm /usr/lib/systemd/system/$(NAME).service
	rm /usr/lib/systemd/system/$(NAME)@.service
	systemctl daemon-reload
	rm /usr/bin/$(NAME)
	rm -rd /etc/$(NAME)
	rm -rd /usr/share/$(NAME)
	rm /usr/bin/geoip.dat
	rm /usr/bin/geosite.dat

%.zip: % geosite.dat geoip.dat
	@zip -du $(NAME)-$@ -j $(BIN_DIR)/$</*
	@zip -du $(NAME)-$@ example/*
	@-zip -du $(NAME)-$@ *.dat
	@echo "<<< ---- $(NAME)-$@"

release: depends darwin-amd64.zip linux-386.zip linux-amd64.zip \
	linux-arm.zip linux-armv5.zip linux-armv6.zip linux-armv7.zip linux-armv8.zip \
	linux-mips-softfloat.zip linux-mips-hardfloat.zip linux-mipsle-softfloat.zip linux-mipsle-hardfloat.zip \
	linux-mips64.zip linux-mips64le.zip freebsd-386.zip freebsd-amd64.zip \
	windows-386.zip windows-amd64.zip

darwin-amd64:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=amd64 GOOS=darwin $(GOBUILD)/$@

linux-386:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=386 GOOS=linux $(GOBUILD)/$@

linux-amd64:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=amd64 GOOS=linux $(GOBUILD)/$@

linux-arm:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=arm GOOS=linux $(GOBUILD)/$@

linux-armv5:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=arm GOOS=linux GOARM=5 $(GOBUILD)/$@

linux-armv6:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=arm GOOS=linux GOARM=6 $(GOBUILD)/$@

linux-armv7:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=arm GOOS=linux GOARM=7 $(GOBUILD)/$@

linux-armv8:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=arm64 GOOS=linux $(GOBUILD)/$@

linux-mips-softfloat:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD)/$@

linux-mips-hardfloat:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=mips GOMIPS=hardfloat GOOS=linux $(GOBUILD)/$@

linux-mipsle-softfloat:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD)/$@

linux-mipsle-hardfloat:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=mipsle GOMIPS=hardfloat GOOS=linux $(GOBUILD)/$@

linux-mips64:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=mips64 GOOS=linux $(GOBUILD)/$@

linux-mips64le:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=mips64le GOOS=linux $(GOBUILD)/$@

freebsd-386:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=386 GOOS=freebsd $(GOBUILD)/$@

freebsd-amd64:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=amd64 GOOS=freebsd $(GOBUILD)/$@

windows-386:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=386 GOOS=windows $(GOBUILD)/$@

windows-amd64:
	mkdir -p $(BIN_DIR)/$@
	GOARCH=amd64 GOOS=windows $(GOBUILD)/$@
