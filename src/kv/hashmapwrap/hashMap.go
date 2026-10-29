package hashmapwrap

import (
	"context"
	"sync"

	"github.com/ds454/auth/src/kv"
	"github.com/ds454/auth/src/log"
)

var store LocalKVStore
var storeOnce sync.Once

type LocalKVStore struct {
	mutex sync.RWMutex
	store map[string]string
}

func GetLocalKVStore() *LocalKVStore {
	storeOnce.Do(func() {
		log.Info("Setting up LocalKVStore...")

		store.store = make(map[string]string)

		log.Info("Done!")
	})
	return &store
}

func (c *LocalKVStore) Get(ctx context.Context, key string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	val, exists := c.store[key]

	if !exists {
		return "", kv.KeyNotFound
	}

	return val, nil
}

func (c *LocalKVStore) Exists(ctx context.Context, key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.store[key]

	return exists
}

func (c *LocalKVStore) Put(ctx context.Context, key string, value string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.store[key] = value

	return true
}

func (c *LocalKVStore) Delete(ctx context.Context, key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.store, key)

	return true
}
