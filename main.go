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
		log.Warn("Using local kv store for session tracking. This does not automatically delete session tokens, and should not be used outside local development.")
		log.Warn("It does not support transactions or locking. Don't try to test race conditions when in this mode.")
		kvStore = hashmapwrap.GetLocalKVStore()

	default:
		log.Abort("Unable to start: mode must be either 'REDIS' or 'BUILTIN'")
	}

	server := grpc.NewServer()
	genauth.RegisterAuthServer(server, &grpcauth.AuthServer{
		KVStore:       kvStore,
		SessionLength: config.SessionKeyLen,
	})

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcListenPort))
	if err != nil {
		log.Error("Unable to bind to port '%d':", config.GrpcListenPort)
		log.Abort("%v", err)
	}

	log.Debug("Starting grpc server")
	if err := server.Serve(listen); err != nil {
		log.Error("Unable to start GRPC server:")
		log.Abort("%v", err)
	}
}
