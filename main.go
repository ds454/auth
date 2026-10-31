package main

import (
	"fmt"
	"net"

	"github.com/ds454/auth/src/config"
	genauth "github.com/ds454/auth/src/generated/ds454/locomotive/auth/v0"
	grpcauth "github.com/ds454/auth/src/grpc"
	"github.com/ds454/auth/src/log"
	"github.com/ds454/auth/src/rediswrap"
	"google.golang.org/grpc"
)

func main() {
	log.Info("Starting auth service...")

	config := config.GetConfig()

	log.Debug("Setting up redis client...")
	redis := rediswrap.GetClient(*config)

	server := grpc.NewServer()
	genauth.RegisterAuthServer(server, &grpcauth.AuthServer{
		Redis: redis,
	})

	log.Debug("Setting up tcp listener...")
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
