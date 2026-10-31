#!/bin/sh

# Install the grpc go plugins
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12

# Add the plugins to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

go mod download
go mod tidy

alias apply='/workspaces/auth/local-dev/apply-proto-go.sh'
