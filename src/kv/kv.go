package kv

import (
	"context"
	"time"
)

type KVStore interface {
	Get(context.Context, string) string
	Put(context.Context, string, string, time.Duration) bool
	Delete(context.Context, string) bool
	Exists(context.Context, string) bool
	TestConnection(context.Context) bool
}
