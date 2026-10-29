package grpcauth

import (
	"context"

	genauth "github.com/ds454/auth/src/generated/ds454/locomotive/auth/v0"
	genproblem "github.com/ds454/auth/src/generated/ds454/locomotive/problem/v0"
	"github.com/ds454/auth/src/kv"
	"github.com/ds454/auth/src/log"
	"github.com/ds454/auth/src/token"
)

type AuthServer struct {
	genauth.UnimplementedAuthServer
	KVStore       kv.KVStore
	SessionLength int32
}

// Deletes the current session token from the key value store
func (s *AuthServer) InvalidateSession(ctx context.Context, in *genauth.InvalidateSessionRequest) (*genauth.InvalidateSessionResponse, error) {
	// Get the uuid we're invalidating
	uuid := in.GetUuid()
	log.Debug("Invalidating session token for uuid '%v'", uuid)

	// Get the session token
	session, err := s.KVStore.Get(ctx, uuid)
	if err != nil {

	}

	redisResult := s.KVStore.Unlink(ctx, session, uuid)

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
	// Get the credentials from the request
	username := in.GetPassword()
	password := in.GetPassword()

	log.Debug("Received login request for %s:%s", username, password)

	var isValid bool
	var uuid string
	var problem genproblem.Problem

	log.Error("TODO: Interact with User Service to vlidate credentials")

	// Validate credentials with User Service
	isValid = true
	uuid = "anotherreallylongcharset"

	// Get problem from User Service
	problem = genproblem.Problem{
		Problem: genproblem.ProblemType_InvalidUsernamePassword,
	}

	// If the username and password are valid...
	if isValid {

		// ...make the session token...
		sessionToken := token.MakeSessionToken(s.SessionLength)

		// ...link them in the kvstore...
		s.KVStore.Link(ctx, sessionToken, uuid)

		// ...and return it to the user
		return &genauth.LoginResponse{
			Response: &genauth.LoginResponse_SessionToken{
				SessionToken: sessionToken,
			},
		}, nil
	}

	// If they aren't valid, we can inspect or obscure the error
	return &genauth.LoginResponse{
		Response: &genauth.LoginResponse_Problem{
			Problem: &problem,
		},
	}, nil
}
