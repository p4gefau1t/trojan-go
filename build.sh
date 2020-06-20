#!/bin/bash

PACKAGE_NAME="github.com/p4gefau1t/trojan-go"
VERSION=`git describe`
COMMIT=`git rev-parse HEAD`

VAR_SETTING=""
VAR_SETTING="$VAR_SETTING -X $PACKAGE_NAME/constant.Version=$VERSION"
VAR_SETTING="$VAR_SETTING -X $PACKAGE_NAME/constant.Commit=$COMMIT"

CGO_ENABLED=0 go build -tags "full" -ldflags="-s -w $VAR_SETTING"
