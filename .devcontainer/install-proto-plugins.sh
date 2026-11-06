#! /bin/sh

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
export PATH="$PATH:$(go env GOPATH)/bin"
