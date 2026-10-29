package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/ds454/auth/src/log"
)

type AuthConfig struct {
	Mode           string `env:"AUTH_STORE_MODE,required"`
	TtlMs          int64  `env:"AUTH_STORE_TTL_MS,required"`
	GrpcListenPort int    `env:"GRPC_PORT,required"`
	RedisPort      string `env:"REDIS_PORT,required"`
	RedisHost      string `env:"REDIS_HOST,required"`
	RedisDB        int    `env:"REDIS_DB,required"`
	RedisPassword  string `env:"REDIS_PASSWORD,required"`
}

var config AuthConfig
var configOnce sync.Once

func GetConfig() *AuthConfig {
	configOnce.Do(func() {
		log.Info("Loading config from environment variables...")

		err := env.Parse(&config)
		if err != nil {
			log.Error("Unable to load server config from environment:")
			log.Abort("%v", err)
		}

		log.Info("Done!")
	})

	return &config
}
