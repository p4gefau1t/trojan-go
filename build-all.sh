#!/bin/bash

PLATFORMS="darwin/amd64 darwin/386"
PLATFORMS="$PLATFORMS windows/amd64 windows/386"
PLATFORMS="$PLATFORMS linux/amd64 linux/386"
PLATFORMS="$PLATFORMS freebsd/amd64 freebsd/386"
PLATFORMS="$PLATFORMS openbsd/amd64 openbsd/386"
PLATFORMS="$PLATFORMS linux/ppc64 linux/ppc64le"
PLATFORMS="$PLATFORMS linux/mips64 linux/mips64le"
PLATFORMS="$PLATFORMS linux/mips linux/mipsle"
PLATFORMS="$PLATFORMS dragonfly/amd64"
PLATFORMS="$PLATFORMS linux/arm64 linux/arm"
PLATFORMS="$PLATFORMS freebsd/arm64 freebsd/arm"
PLATFORMS="$PLATFORMS openbsd/arm64 openbsd/arm"
PLATFORMS="$PLATFORMS linux/s390x"

type setopt >/dev/null 2>&1

rm -rd release

SCRIPT_NAME=`basename "$0"`
FAILURES=""
CURRENT_DIRECTORY=${PWD##*/}
OUTPUT=release/$CURRENT_DIRECTORY

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}"
  if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
  CMD="CGO_ENABLE=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BIN_FILENAME} $@ -ldflags=\"-s -w\""
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
done

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "${SCRIPT_NAME} failed on: ${FAILURES}"
fi

cd release
wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O geosite.dat
wget https://raw.githubusercontent.com/v2ray/geoip/release/geoip.dat -O geoip.dat
cp ../data/* ./

for name in trojan-go*;do
  zip $name.zip client.json server.json trojan-go.service geoip.dat geosite.dat $name
  sha1sum $name.zip > $name.zip.sha1
  rm $name
done