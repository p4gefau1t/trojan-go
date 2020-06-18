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
PLATFORMS="$PLATFORMS windows/armv5 linux/armv5 freebsd/armv5 netbsd/armv5"
PLATFORMS="$PLATFORMS windows/armv6 linux/armv6 freebsd/armv6 netbsd/armv6"
PLATFORMS="$PLATFORMS windows/armv7 linux/armv7 freebsd/armv7 netbsd/armv7"
PLATFORMS="$PLATFORMS android/armv7 android/386 android/arm64 android/amd64"

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

PACKAGE_NAME="github.com/p4gefau1t/trojan-go"
COMMIT=`git rev-parse HEAD`
VERSION=`git tag --contains ${COMMIT}`

VAR_SETTING=""
VAR_SETTING="$VAR_SETTING -X $PACKAGE_NAME/constant.Version=$VERSION"
VAR_SETTING="$VAR_SETTING -X $PACKAGE_NAME/constant.Commit=$COMMIT"
#android
NDK_VER="r20"
COS=${1}
shift 1
NDK_DL="android-ndk-${NDK_VER}-${COS}.zip"

PREPARE_NDK() {
  [ -d "${NDK_TOOLS}" ] || {
    echo ">>> Start Android NDK env..."
    curl -LOs https://dl.google.com/android/repository/${NDK_DL}
    unzip -q ${NDK_DL} -d ${HOME}
    NDK_TOOLS=${HOME}/android-ndk-${NDK_VER}
    export PATH=${PATH}:${NDK_TOOLS}/toolchains/llvm/prebuilt/linux-x86_64/bin
  }

  unset APREFIX
  case ${1} in
    "arm" )   APREFIX="armv7a-linux-androideabi19-";;
    "386" )   APREFIX="i686-linux-android19-";;
    "arm64" ) APREFIX="aarch64-linux-android21-";;
    "amd64" ) APREFIX="x86_64-linux-android21-";;
    * )       echo "Skiped architech: [${1}]";;
  esac

  export CC=${APREFIX:+${APREFIX}clang}
  export CXX=${APREFIX:+${APREFIX}c++}
  export STRIP=${APREFIX:+${APREFIX}strip}
}

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  ZIP_FILENAME="trojan-go-${GOOS}-${GOARCH}.zip"
  [[ "x${GOARCH}" =~ xarmv[5-7] ]] && GOARM="${GOARCH:4:1}" && GOARCH="arm"
  CGO=0
  [ "x${GOOS}" = "xandroid" ] && PREPARE_NDK "${GOARCH}" && CGO=1

  CMD="env -v CGO_ENABLED=${CGO} ${GOARM:+GOARM=${GOARM}} GOOS=${GOOS} GOARCH=${GOARCH} go build -tags \"full\" -o temp -ldflags=\"-s -w ${VAR_SETTING}\""
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
  zip -j release/$ZIP_FILENAME temp/* ./*.dat
  zip release/$ZIP_FILENAME example/*
  sha1sum release/$ZIP_FILENAME >> release/checksum.sha1
  rm temp/*
done
rm -vf "${NDK_DL}"

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "${SCRIPT_NAME} failed on: ${FAILURES}"
fi
