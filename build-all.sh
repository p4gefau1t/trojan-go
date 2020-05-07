#!/bin/bash

PLATFORMS="darwin/amd64 darwin/386"
PLATFORMS="$PLATFORMS windows/amd64 windows/386"
PLATFORMS="$PLATFORMS windows/arm"
PLATFORMS="$PLATFORMS linux/amd64 linux/386"
PLATFORMS="$PLATFORMS linux/ppc64 linux/ppc64le"
PLATFORMS="$PLATFORMS linux/mips64 linux/mips64le"
PLATFORMS="$PLATFORMS linux/mips linux/mipsle"
PLATFORMS="$PLATFORMS linux/arm64 linux/arm"
PLATFORMS="$PLATFORMS linux/s390x"
PLATFORMS="$PLATFORMS dragonfly/amd64"
PLATFORMS="$PLATFORMS openbsd/arm64 openbsd/arm"
PLATFORMS="$PLATFORMS openbsd/amd64 openbsd/386"
PLATFORMS="$PLATFORMS freebsd/amd64 freebsd/386"
PLATFORMS="$PLATFORMS freebsd/arm64 freebsd/arm"

type setopt >/dev/null 2>&1

rm -rd release
rm -rd temp
rm ./*.dat

mkdir release
mkdir temp

wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O geosite.dat
wget https://github.com/v2ray/geoip/raw/release/geoip.dat -O geoip.dat


SCRIPT_NAME=`basename "$0"`
FAILURES=""

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  ZIP_FILENAME="trojan-go-${GOOS}-${GOARCH}.zip"
  CMD="CGO_ENABLE=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -tags \"full\" -o temp $@ -ldflags=\"-s -w\""
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
  zip -j release/$ZIP_FILENAME temp/* ./*.dat
  zip release/$ZIP_FILENAME example/*
  sha1sum release/$ZIP_FILENAME > release/$ZIP_FILENAME.sha1
  rm temp/*
done

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "${SCRIPT_NAME} failed on: ${FAILURES}"
fi
