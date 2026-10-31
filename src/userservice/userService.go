package userservice

import (
	"context"

	"github.com/ds454/auth/src/log"
)

// TODO: Call user service
func UsernamePasswordAreValid(ctx context.Context, username string, password string) (string, bool, error) {
	log.Error("TODO: Call user service")
	return "averylonguuid", true, nil
}
