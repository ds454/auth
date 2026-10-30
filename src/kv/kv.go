package kv

import (
	"context"
	"errors"
)

var KeyNotFound = errors.New("Key not found")

type KVStore interface {

	// Returns KeyNotFound error if key is missing
	Get(ctx context.Context, key string) (string, error)

	// time.Duration is ignored for LocalKVStore
	Put(ctx context.Context, key string, value string) bool
	Link(ctx context.Context, key string, value string) bool
	Unlink(ctx context.Context, key string, value string) bool
	Delete(ctx context.Context, key string) bool
	Exists(ctx context.Context, key string) bool

	TryLock(ctx context.Context, name string, ttlMs int) bool
}
