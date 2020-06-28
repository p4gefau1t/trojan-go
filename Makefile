NAME=trojan-go
PACKAGE_NAME=github.com/p4gefau1t/trojan-go
VERSION=`git describe`
COMMIT=`git rev-parse HEAD`

BIN_DIR=bin
VAR_SETTING=-X $(PACKAGE_NAME)/constant.Version=$(VERSION) -X $(PACKAGE_NAME)/constant.Commit=$(COMMIT)
GOBUILD=CGO_ENABLED=0 go build -tags "full" -ldflags="-s -w $(VAR_SETTING)" -o $(BIN_DIR)/$(NAME)

normal: clean
	$(GOBUILD)

clean:
	-rm -frd $(BIN_DIR)

geoip.dat:
	wget https://github.com/v2ray/geoip/raw/release/geoip.dat

geosite.dat:
	wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O geosite.dat

install: $(BIN_DIR)/$(NAME) geoip.dat geosite.dat
	mkdir -p /etc/$(NAME)
	mkdir -p /usr/share/$(NAME)
	cp example/*.json /etc/$(NAME)
	cp $(BIN_DIR)/$(NAME) /usr/bin/$(NAME)
	cp example/$(NAME).service /etc/systemd/system
	cp example/$(NAME)@.service /etc/systemd/system
	cp geosite.dat /usr/share/$(NAME)/geosite.dat
	cp geoip.dat /usr/share/$(NAME)/geoip.dat
	ln -fs /usr/share/$(NAME)/geoip.dat /usr/bin/
	ln -fs /usr/share/$(NAME)/geosite.dat /usr/bin/

uninstall:
	rm /etc/systemd/system/$(NAME).service
	rm /etc/systemd/system/$(NAME)@.service
	rm /usr/bin/$(NAME)
	rm -rd /etc/$(NAME)
	rm -rd /usr/share/$(NAME)
	rm /usr/bin/geoip.dat
	rm /usr/bin/geosite.dat

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD)

linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD)

linux-armv5:
	GOARCH=arm GOOS=linux GOARM=5 $(GOBUILD)

linux-armv6:
	GOARCH=arm GOOS=linux GOARM=6 $(GOBUILD)

linux-armv7:
	GOARCH=arm GOOS=linux GOARM=7 $(GOBUILD)

linux-armv8:
	GOARCH=arm64 GOOS=linux $(GOBUILD)

linux-mips-softfloat:
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD)

linux-mips-hardfloat:
	GOARCH=mips GOMIPS=hardfloat GOOS=linux $(GOBUILD)

linux-mipsle-softfloat:
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD)

linux-mipsle-hardfloat:
	GOARCH=mipsle GOMIPS=hardfloat GOOS=linux $(GOBUILD)

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GOBUILD)

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GOBUILD)

freebsd-386:
	GOARCH=386 GOOS=freebsd $(GOBUILD)

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GOBUILD)

windows-386:
	GOARCH=386 GOOS=windows $(GOBUILD)

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD)
