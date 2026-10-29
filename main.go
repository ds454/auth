package main

import (
	"fmt"
	"net"

	"github.com/ds454/auth/src/config"
	genauth "github.com/ds454/auth/src/generated/ds454/locomotive/auth/v0"
	grpcauth "github.com/ds454/auth/src/grpc"
	"github.com/ds454/auth/src/kv"
	"github.com/ds454/auth/src/kv/hashmapwrap"
	"github.com/ds454/auth/src/kv/rediswrap"
	"github.com/ds454/auth/src/log"
	"google.golang.org/grpc"
)

func main() {
	log.Info("Starting auth service...")

	config := config.GetConfig()

	var kvStore kv.KVStore

	switch config.Mode {
	case "REDIS":
		kvStore = rediswrap.GetClient(*config)

	case "BUILTIN":
		kvStore = hashmapwrap.GetLocalKVStore()

	default:
		log.Abort("Unable to start: mode must be either 'REDIS' or 'BUILTIN'")
	}

	server := grpc.NewServer()
	genauth.RegisterAuthServer(server, &grpcauth.AuthServer{
		KVStore: kvStore,
	})

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcListenPort))
	if err != nil {
		log.Error("Unable to bind to port '%d':", config.GrpcListenPort)
		log.Abort("%v", err)
	}

	log.Info("Starting grpc server")
	if err := server.Serve(listen); err != nil {
		log.Error("Unable to start GRPC server:")
		log.Abort("%v", err)
	}
}
