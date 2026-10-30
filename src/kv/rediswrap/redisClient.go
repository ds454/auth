package rediswrap

import (
	"context"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/ds454/auth/src/config"
	"github.com/ds454/auth/src/log"
	"github.com/redis/go-redis/v9"

	"github.com/ds454/auth/src/kv"
)

var client RedisClient
var redisOnce sync.Once

type RedisClient struct {
	client *redis.Client
	locker *redislock.Client
}

func GetClient(cfg config.AuthConfig) *RedisClient {
	redisOnce.Do(func() {
		log.Debug("Setting up RedisClient...")

		conf := cfg

		addr := conf.RedisHost + ":" + conf.RedisPort

		redisClient := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: conf.RedisPassword,
			DB:       conf.RedisDB,
			Protocol: 2,
		})

		client.client = redisClient

		for {
			log.Debug("Testing connection...")
			// context.Background() is fine here since this only gets called during setup
			if !client.TestConnection(context.Background()) {
				log.Error("Unable to connect to Redis at %v", addr)
				log.Error("Retrying in 5 seconds...")
				time.Sleep(time.Duration(time.Second * 5))
			} else {
				break
			}
		}
		log.Debug("Connection successful!")

		log.Debug("Setting up locker...")
		client.locker = redislock.New(redisClient)
		log.Debug("Done")
	})
	return &client
}

// Error is set if key does not exist
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		log.Error("Problem getting key '%v' from Redis:", key)
		log.Error("%v", err)
		return "", err
	}

	if val == "" {
		return "", kv.KeyNotFound
	}

	return val, nil
}

func (c *RedisClient) Put(ctx context.Context, key string, value string) bool {
	_, err := c.client.Set(ctx, key, value, time.Duration(config.GetConfig().TtlMs)*time.Millisecond).Result()
	if err != nil {
		log.Error("Problem setting '%v' = '%v':", key, value)
		log.Error("%v", err)
		return false
	}
	return true
}

func (c *RedisClient) TryLock(ctx context.Context, key string, ttlMs int) bool {
	_, err := c.locker.Obtain(ctx, key, time.Duration(ttlMs)*time.Millisecond, nil)
	if err != nil {
		log.Debug("Did not acquire redis lock: %v", err)
		return false
	}
	log.Debug("Acquired lock")
	return true
}

func (c *RedisClient) Link(ctx context.Context, key string, value string) bool {

	transaction := c.client.TxPipeline()

	transaction.Set(ctx, key, value, time.Duration(config.GetConfig().TtlMs)*time.Millisecond)
	transaction.Set(ctx, value, key, time.Duration(config.GetConfig().TtlMs)*time.Millisecond)

	_, err := transaction.Exec(ctx)
	if err != nil {
		log.Error("Problem linking '%v' <---> '%v':", key, value)
		log.Error("%v", err)
		return false
	}
	return true
}

func (c *RedisClient) Unlink(ctx context.Context, key string, value string) bool {

	transaction := c.client.TxPipeline()

	transaction.Del(ctx, key)
	transaction.Del(ctx, value)

	_, err := transaction.Exec(ctx)
	if err != nil {
		log.Error("Problem unlinking '%v' <---> '%v':", key, value)
		log.Error("%v", err)
		return false
	}
	return true
}

func (c *RedisClient) Delete(ctx context.Context, key string) bool {
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
		// val > 1
		log.Warn("Deleted %d keys intead of the expected 1, ignoring", val)
		return true
	}
}

func (c *RedisClient) Exists(ctx context.Context, key string) bool {
	// Val is the number of existing keys
	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		log.Error("Problem checking if '%v' exists:", key)
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

func (c *RedisClient) TestConnection(ctx context.Context) bool {
	// Val is "PONG" if successful
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
		log.Error("Unable to connect to redis! Received '%v' instead of expected 'PONG'", val)
		return false
	}
}
