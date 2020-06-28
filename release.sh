#!/bin/bash

PLATFORMS="$PLATFORMS darwin-amd64"
PLATFORMS="$PLATFORMS linux-386"
PLATFORMS="$PLATFORMS linux-amd64"
PLATFORMS="$PLATFORMS linux-armv5"
PLATFORMS="$PLATFORMS linux-armv6"
PLATFORMS="$PLATFORMS linux-armv7"
PLATFORMS="$PLATFORMS linux-armv8"
PLATFORMS="$PLATFORMS linux-mips-softfloat"
PLATFORMS="$PLATFORMS linux-mips-hardfloat"
PLATFORMS="$PLATFORMS linux-mipsle-softfloat"
PLATFORMS="$PLATFORMS linux-mipsle-hardfloat"
PLATFORMS="$PLATFORMS linux-mips64"
PLATFORMS="$PLATFORMS linux-mips64le"
PLATFORMS="$PLATFORMS freebsd-386"
PLATFORMS="$PLATFORMS freebsd-amd64"
PLATFORMS="$PLATFORMS windows-386"
PLATFORMS="$PLATFORMS windows-amd64"

rm -rdf release
rm -rdf temp
mkdir release
mkdir temp

wget https://github.com/v2ray/domain-list-community/raw/release/dlc.dat -O temp/geosite.dat
wget https://github.com/v2ray/geoip/raw/release/geoip.dat -O temp/geoip.dat

for PLATFORM in $PLATFORMS; do
	make clean
	eval "make $PLATFORM" || FAILURES="${PLATFORM}"
	if [[ "${FAILURES}" != "" ]]; then
		echo "failed on: ${FAILURES}"
		exit 1
	fi

	ZIP_FILENAME=trojan-go-$PLATFORM.zip
	zip -j release/$ZIP_FILENAME bin/trojan-go
	zip release/$ZIP_FILENAME example/*
	zip -j release/$ZIP_FILENAME temp/*
done

make clean
rm -rdf temp