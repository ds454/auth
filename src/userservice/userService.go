package userservice

import (
	"context"
	"fmt"
	"strconv"

	genuser "github.com/ds454/auth/src/generated/ds454/locomotive/user/v0"
	"github.com/ds454/auth/src/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userServiceChannel *grpc.ClientConn

func InitChannel(host string, port int) {
	connection, err := grpc.NewClient(host+":"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Unable to connect to user service:")
		log.Abort("%v", err)
	} else {
		userServiceChannel = connection
	}
}

// TODO: Call user service
func UsernamePasswordAreValid(ctx context.Context, username string, password string) (string, string, bool, error) {
	client := genuser.NewInternalUserServiceClient(userServiceChannel)
	var req = genuser.ValidateCredentialRequest{
		Username: username,
		Password: password,
	}
	result, err := client.ValidateCredentials(ctx, &req)
	if err != nil {
		log.Error("Problem checking if credentials are valid:")
		log.Error("%v", err)
		return "", "", false, err
	}
	if result.GetProblem() != nil {
		log.Debug("Failed user authentication request")
		return "", "", false, fmt.Errorf("%s", result.GetProblem().GetMessage())
	}

	return result.GetIds().Uuid, result.GetIds().Steamid, true, nil
}
