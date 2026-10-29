package token

import (
	"math/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func MakeSessionToken(length int32) string {
	var sb strings.Builder
	sb.Grow(int(length))
	for i := 0; i < int(length); i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
