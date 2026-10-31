package grpcauth

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/ds454/auth/src/config"
	genauth "github.com/ds454/auth/src/generated/ds454/locomotive/auth/v0"
	genproblem "github.com/ds454/auth/src/generated/ds454/locomotive/problem/v0"
	"github.com/ds454/auth/src/log"
	"github.com/ds454/auth/src/token"
	"github.com/ds454/auth/src/userservice"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const sessionTokenPrefix string = "session:"
const userSessionsPrefix string = "user:"

type AuthServer struct {
	genauth.UnimplementedAuthServer
	authv3.UnimplementedAuthorizationServer
	Redis *redis.Client
}

func (s *AuthServer) LogoutEverywhere(ctx context.Context, req *genauth.LogoutEverywhereRequest) (*genauth.LogoutEverywhereResponse, error) {
	uuid, err := uuidFromContext(ctx)
	if err != nil {
		return nil, err
	}

	refs, problem := getAllUsersSessionTokens(ctx, s, uuid)
	if problem != nil {
		return &genauth.LogoutEverywhereResponse{
			Problem: problem,
		}, nil
	}

	pipeline := s.Redis.TxPipeline()

	// Delete the user set
	pipeline.Del(ctx, userSessionsPrefix+uuid).Result()

	// Clean up the metadata
	for _, sessionToken := range refs {
		pipeline.Del(ctx, sessionTokenPrefix+sessionToken)
	}

	_, err = pipeline.Exec(ctx)
	if err != nil {
		log.Error("Problem clearing session metadata:")
		log.Error("%v", err)
		return &genauth.LogoutEverywhereResponse{
			Problem: &genproblem.Problem{
				Problem: genproblem.ProblemType_ServerError,
			},
		}, nil
	}

	return &genauth.LogoutEverywhereResponse{}, nil
}

func (s *AuthServer) LogoutSomewhere(ctx context.Context, req *genauth.LogoutSomewhereRequest) (*genauth.LogoutSomewhereResponse, error) {
	uuid, err := uuidFromContext(ctx)
	if err != nil {
		return nil, err
	}

	refs, problem := getAllUsersSessionTokens(ctx, s, uuid)
	if problem != nil {
		return &genauth.LogoutSomewhereResponse{
			Problem: problem,
		}, nil
	}

	resultsSlice := make([]*redis.StringCmd, len(refs))

	// A quick pipeline lowers network overhead
	pipeline := s.Redis.Pipeline()

	for index, sessionToken := range refs {
		// From token/session.go. Same order as refs, so we can map session tokens --> session token ids
		resultsSlice[index] = pipeline.HGet(ctx, sessionTokenPrefix+sessionToken, "stknid")
	}

	_, err = pipeline.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error("Problem getting session hashes:")
		log.Error("%v", err)
		return &genauth.LogoutSomewhereResponse{
			Problem: &genproblem.Problem{
				Problem: genproblem.ProblemType_ServerError,
			},
		}, nil
	}

	sessionTokenId := req.GetSessionTokenId()
	for index, redisResult := range resultsSlice {
		result, err := redisResult.Result()

		// Someone else already deleted this session token
		if errors.Is(err, redis.Nil) {
			continue
		} else if err != nil {
			log.Error("Problem getting session hash '%s':", refs[index])
			log.Error("%v", err)
			return &genauth.LogoutSomewhereResponse{
				Problem: &genproblem.Problem{
					Problem: genproblem.ProblemType_ServerError,
				},
			}, nil
		}

		if result == sessionTokenId {
			transaction := s.Redis.TxPipeline()
			transaction.Del(ctx, sessionTokenPrefix+refs[index])
			transaction.SRem(ctx, userSessionsPrefix+uuid, refs[index])
			_, err := transaction.Exec(ctx)
			if err != nil {
				log.Error("Problem deleting session token %s:", refs[index])
				log.Error("%v", err)
				return &genauth.LogoutSomewhereResponse{
					Problem: &genproblem.Problem{
						Problem: genproblem.ProblemType_ServerError,
					},
				}, nil
			}

			return &genauth.LogoutSomewhereResponse{}, nil
		}
	}

	// Assume someone else already deleted the session
	return &genauth.LogoutSomewhereResponse{}, nil
}

func (s *AuthServer) GetActiveSessions(ctx context.Context, req *genauth.GetActiveSessionsRequest) (*genauth.GetActiveSessionsResponse, error) {
	uuid, err := uuidFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userSessionKeySetKey := userSessionsPrefix + uuid

	refs, problem := getAllUsersSessionTokens(ctx, s, uuid)
	if problem != nil {
		return &genauth.GetActiveSessionsResponse{
			Result: &genauth.GetActiveSessionsResponse_Problem{
				Problem: problem,
			},
		}, nil
	}

	tokenMetadataSlice := make([]token.SessionToken, 0, len(refs))
	for _, sessionToken := range refs {

		sessionTokenKey := sessionTokenPrefix + sessionToken

		var tokenMetadata token.SessionToken

		// Load the token metadata into a struct
		cmd := s.Redis.HGetAll(ctx, sessionTokenKey)

		fields, err := cmd.Result()
		if err != nil {
			return &genauth.GetActiveSessionsResponse{
				Result: &genauth.GetActiveSessionsResponse_Problem{
					Problem: &genproblem.Problem{
						Problem: genproblem.ProblemType_ServerError,
					},
				},
			}, nil
		}

		// If the token metadata has been deleted or has expired, we can remove it from our users set
		if len(fields) == 0 {
			_, err = s.Redis.SRem(ctx, userSessionKeySetKey, sessionToken).Result()

			// Not mission critical if this fails, since login checks if the token has expired or not.
			if err != nil {
				log.Error("Problem cleaning up expired session keys:")
				log.Error("%v", err)
			}

			continue
		}

		if err := cmd.Scan(&tokenMetadata); err != nil {
			log.Error("Problem converting token metadata into struct:")
			log.Error("%v", err)
			continue
		}

		tokenMetadataSlice = append(tokenMetadataSlice, tokenMetadata)
	}

	slices.SortFunc(tokenMetadataSlice, func(a, b token.SessionToken) int {
		return cmp.Compare(a.Expiration, b.Expiration)
	})

	var result genauth.Sessions
	result.Sessions = make([]*genauth.Session, 0, len(tokenMetadataSlice))
	for _, value := range tokenMetadataSlice {
		result.Sessions = append(result.Sessions, &genauth.Session{
			SessionTokenId: value.SessionTokenId,
			Device:         &value.Device,
			Expiration:     value.Expiration,
		})
	}

	return &genauth.GetActiveSessionsResponse{
		Result: &genauth.GetActiveSessionsResponse_Sessions{
			Sessions: &result,
		},
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *genauth.LoginRequest) (*genauth.LoginResponse, error) {

	username := req.GetUsername()
	password := req.GetPassword()

	uuid, isValid, err := userservice.UsernamePasswordAreValid(ctx, username, password)

	if err != nil {
		return &genauth.LoginResponse{
			Response: &genauth.LoginResponse_Problem{
				Problem: &genproblem.Problem{
					Problem: genproblem.ProblemType_ServerError,
				},
			},
		}, nil
	}

	if !isValid {
		// Invalid username or password
		return &genauth.LoginResponse{
			Response: &genauth.LoginResponse_Problem{
				Problem: &genproblem.Problem{
					Problem: genproblem.ProblemType_InvalidUsernamePassword,
				},
			},
		}, nil
	}

	var device string
	if req.DeviceType == nil {
		device = "Unknown"
	} else {
		device = req.GetDeviceType()
	}

	sessionToken := token.MakeRandomString(32)
	expiration := config.GetConfig().TtlS + time.Now().Unix()
	transaction := s.Redis.TxPipeline()

	// Add the session token to the user's session token set
	transaction.SAdd(ctx, userSessionsPrefix+uuid, sessionToken)

	// Add the session token and its metadata
	transaction.HSet(ctx, sessionTokenPrefix+sessionToken, token.SessionToken{
		SessionTokenId: token.MakeRandomString(8),
		Uuid:           uuid,
		Device:         device,
		Expiration:     expiration,
	})

	// Set the hash to automatically expire
	transaction.Expire(ctx, sessionTokenPrefix+sessionToken, time.Duration(config.GetConfig().TtlS)*time.Second)

	// This will clear redis over time if the user doesn't login again
	transaction.Expire(ctx, userSessionsPrefix+uuid, time.Duration(config.GetConfig().TtlS)*time.Second)

	_, err = transaction.Exec(ctx)

	if err != nil {
		log.Error("Problem logging in uuid %s:", uuid)
		log.Error("%v", err)
		return &genauth.LoginResponse{
			Response: &genauth.LoginResponse_Problem{
				Problem: &genproblem.Problem{
					Problem: genproblem.ProblemType_UnableToLogin,
				},
			},
		}, nil
	}

	log.Debug("Logging in uuid %s", uuid)
	return &genauth.LoginResponse{
		Response: &genauth.LoginResponse_SessionToken{
			SessionToken: sessionToken,
		},
	}, nil
}

func getAllUsersSessionTokens(ctx context.Context, s *AuthServer, uuid string) ([]string, *genproblem.Problem) {
	refs, err := s.Redis.SMembers(ctx, userSessionsPrefix+uuid).Result()
	if err != nil {
		log.Error("Unable to connect to redis:")
		log.Error("%v", err)
		return nil, &genproblem.Problem{
			Problem: genproblem.ProblemType_ServerError,
		}
	}
	return refs, nil
}

// Generated by Claude
func uuidFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "")
	}

	uuids := md.Get(config.GetConfig().UuidMetadataKey)
	if len(uuids) != 1 {
		return "", status.Error(codes.Unauthenticated, "")
	}
	return uuids[0], nil
}
