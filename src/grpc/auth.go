package grpcauth

import (
	"context"

	genauth "github.com/ds454/auth/src/generated/ds454/locomotive/auth/v0"
	genproblem "github.com/ds454/auth/src/generated/ds454/locomotive/problem/v0"
	"github.com/ds454/auth/src/kv"
	"github.com/ds454/auth/src/log"
)

type AuthServer struct {
	genauth.UnimplementedAuthServer
	KVStore kv.KVStore
}

// Deletes the current session token from the key value store
func (s *AuthServer) InvalidateSession(ctx context.Context, in *genauth.InvalidateSessionRequest) (*genauth.InvalidateSessionResponse, error) {
	uuid := in.GetUuid()
	log.Debug("Invalidating session token for uuid '%v'", uuid)
	redisResult := s.KVStore.Delete(ctx, uuid)

	if redisResult {
		log.Debug("Session token for uuid '%v' deleted", uuid)
		return &genauth.InvalidateSessionResponse{}, nil
	}

	log.Debug("Unable to delete session token for user '%v'", uuid)
	return &genauth.InvalidateSessionResponse{
		Problem: &genproblem.Problem{
			Problem: *genproblem.ProblemType_UnableToDeleteSessionToken.Enum(),
		},
	}, nil
}

// Login
func (s *AuthServer) Login(ctx context.Context, in *genauth.LoginRequest) (*genauth.LoginResponse, error) {
	return nil, nil
}
