package token

import (
	"math/rand"
	"strings"
)

type SessionToken struct {
	SessionTokenId string `redis:"stknid"`
	Uuid           string `redis:"uuid"`
	Device         string `redis:"dev"`
	Expiration     int64  `redis:"exp"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func MakeRandomString(length int32) string {
	var sb strings.Builder
	sb.Grow(int(length))
	for i := 0; i < int(length); i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
