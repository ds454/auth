package rediswrap

import (
	"context"
	"sync"
	"time"

	"github.com/ds454/auth/src/config"
	"github.com/ds454/auth/src/log"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisOnce sync.Once

func GetClient(cfg config.AuthConfig) *redis.Client {
	redisOnce.Do(func() {
		log.Debug("Setting up RedisClient...")

		conf := cfg

		addr := conf.RedisHost + ":" + conf.RedisPort

		redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: conf.RedisPassword,
			DB:       conf.RedisDB,
			Protocol: 2,
		})

		for {
			log.Debug("Testing connection...")
			// context.Background() is fine here since this only gets called during setup
			if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
				log.Error("Unable to connect to Redis at %v:", addr)
				log.Error("%v", err)
				log.Error("Retrying in 5 seconds...")
				time.Sleep(time.Duration(time.Second * 5))
			} else {
				break
			}
		}
		log.Debug("Connection successful!")

	})
	return redisClient
}
