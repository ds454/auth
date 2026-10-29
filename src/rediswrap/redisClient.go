package rediswrap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ds454/auth/src/config"
	"github.com/ds454/auth/src/log"
	"github.com/redis/go-redis/v9"
)

var client RedisClient
var redisOnce sync.Once

type RedisClient struct {
	client redis.Client
}

func GetClient() *RedisClient {
	redisOnce.Do(func() {
		log.Info("Setting up redis client...")

		conf := config.GetConfig()

		addr := conf.RedisHost + ":" + conf.RedisPort

		redisClient := *redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: conf.RedisPassword,
			DB:       conf.RedisDB,
			Protocol: 2,
		})

		client.client = redisClient

		log.Info("Done!")
	})
	return &client
}

func (c RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		log.Error("Problem getting key '%v' from Redis:", key)
		log.Error("%v", err)
		return "", err
	}

	if val == "" {
		return "", fmt.Errorf("Key '%v' does not exist", key)
	}

	return val, nil
}

func (c RedisClient) Put(ctx context.Context, key string, value string, ttl time.Duration) bool {
	_, err := c.client.Set(ctx, key, value, ttl).Result()
	if err != nil {
		log.Error("Problem setting '%v' = '%v':", key, value)
		log.Error("%v", err)
		return false
	}
	return true
}

func (c RedisClient) Delete(ctx context.Context, key string) bool {
	// Val is the number of deleted keys
	val, err := c.client.Del(ctx, key).Result()
	if err != nil {
		log.Error("Problem removing key '%v':", key)
		log.Error("%v", err)
		return false
	}

	switch val {
	case 0:
		log.Debug("Tried to delete a missing key, ignoring")
		return true
	case 1:
		return true
	default:
		// val != 1
		log.Warn("Deleted %d keys intead of the expected 1, ignoring", val)
		return true
	}
}

func (c RedisClient) Exists(ctx context.Context, key string, value string, ttl time.Duration) bool {
	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		log.Error("Problem setting '%v' = '%v':", key, value)
		log.Error("%v", err)
		return false
	}

	switch val {
	case 0:
		return false
	case 1:
		return true
	default:
		log.Warn("Received unexpected number of existing keys: %d", val)
		return true
	}
}

func (c RedisClient) TextConnection(ctx context.Context) bool {
	val, err := c.client.Ping(ctx).Result()
	if err != nil {
		log.Error("Problem connecting to redis!")
		log.Error("%v", err)
		return false
	}

	switch val {
	case "PONG":
		log.Info("Connected to redis")
		return true
	default:
		log.Error("Unable to connect to redis!")
		return false
	}
}
