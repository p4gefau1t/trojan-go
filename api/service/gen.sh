#!/usr/bin/env bash

echo "Processing..."

GOPATH=${GOPATH:-$(go env GOPATH)}
GOBIN=${GOBIN:-$(go env GOBIN)}

if [[ $GOBIN == "" ]]; then
  GOBIN=${GOPATH}/bin
fi

go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

echo "Use protoc-gen-go and protoc-gen-go-grpc in $GOBIN."

protoc --go_out=. \
--go_opt=paths=source_relative \
--go-grpc_out=. \
--go-grpc_opt=paths=source_relative \
--plugin=protoc-gen-go=${GOBIN}/protoc-gen-go \
--plugin=protoc-gen-go-grpc=${GOBIN}/protoc-gen-go-grpc \
api.proto

if [ $? -eq 0 ]; then
  echo "Generated successfully."
fi
