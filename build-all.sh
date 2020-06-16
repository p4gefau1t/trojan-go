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

PACKAGE_NAME="github.com/p4gefau1t/trojan-go"
VERSION=`git describe`
COMMIT=`git rev-parse HEAD`

VAR_SETTING=""
VAR_SETTING="$VAR_SETTING -X $PACKAGE_NAME/constant.Version=$VERSION"
VAR_SETTING="$VAR_SETTING -X $PACKAGE_NAME/constant.Commit=$COMMIT"

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  ZIP_FILENAME="trojan-go-${GOOS}-${GOARCH}.zip"
  CMD="CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -tags \"full\" -o temp -ldflags=\"-s -w ${VAR_SETTING}\""
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
  zip -j release/$ZIP_FILENAME temp/* ./*.dat
  zip release/$ZIP_FILENAME example/*
  sha1sum release/$ZIP_FILENAME >> release/checksum.sha1
  rm temp/*
done

# arm
PLATFORMS_ARM="windows linux freebsd netbsd"
for GOOS in $PLATFORMS_ARM; do
  # build for each ARM version
  GOARCH="arm"
  for GOARM in 7 6 5; do
    ZIP_FILENAME="trojan-go-${GOOS}-${GOARCH}v${GOARM}.zip"
    CMD="CGO_ENABLED=0 GOARM=${GOARM} GOOS=${GOOS} GOARCH=${GOARCH} go build -tags \"full\" -o temp -ldflags \"-s -w ${VAR_SETTING}\""
    echo "${CMD}"
    eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}v${GOARM}" 
    zip -j release/$ZIP_FILENAME temp/* ./*.dat
    zip release/$ZIP_FILENAME example/*
    sha1sum release/$ZIP_FILENAME >> release/checksum.sha1
    rm temp/*
  done
done

#android
NDK_VER="r20"
NDK_DL="android-ndk-${NDK_VER}-linux-x86_64.zip"

PREPARE_NDK() {
  [ -d "${NDK_TOOLS}" ] || {
    echo ">>> Start Android NDK env..."
    curl -LOs https://dl.google.com/android/repository/${NDK_DL}
    unzip -q ${NDK_DL} -d ${HOME}
    NDK_TOOLS=${HOME}/android-ndk-${NDK_VER}
    export PATH=${PATH}:${NDK_TOOLS}/toolchains/llvm/prebuilt/linux-x86_64/bin
  }

  unset ANDROID_TOOLCHAIN_PREFIX
  case ${1} in
    "arm" )   ANDROID_TOOLCHAIN_PREFIX="armv7a-linux-androideabi19-";;
    "386" )   ANDROID_TOOLCHAIN_PREFIX="i686-linux-android19-";;
    "arm64" ) ANDROID_TOOLCHAIN_PREFIX="aarch64-linux-android21-";;
    "amd64" ) ANDROID_TOOLCHAIN_PREFIX="x86_64-linux-android21-";;
    * )       echo "Skiped architech: [${1}]";;
  esac
  [ -z "${ANDROID_TOOLCHAIN_PREFIX}" ] && return
  export CC=${ANDROID_TOOLCHAIN_PREFIX}clang
  export STRIP=${ANDROID_TOOLCHAIN_PREFIX}strip
  export CXX=${ANDROID_TOOLCHAIN_PREFIX}c++
}

ANDROID_ARCH="arm 386 arm64 amd64"
for GOARCH in $ANDROID_ARCH; do
  # build for android version 7 only on ARM
  go clean
  unset GOARM CC CXX
  GOOS="android"
  [ "$GOARCH" == "arm" ] && GOARM="GOARM=7"
  PREPARE_NDK "$GOARCH"
  ZIP_FILENAME="trojan-go-${GOOS}-${GOARCH}${GOARM:+v7}.zip"
  CMD="env CGO_ENABLED=1 ${GOARM} GOOS=${GOOS} GOARCH=${GOARCH} go build -tags \"full\" -o temp -ldflags \"-s -w ${VAR_SETTING}\""
  echo "${CMD}"
  eval ${CMD} || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM:+v7}"
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
