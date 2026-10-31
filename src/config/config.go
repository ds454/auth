package config

import (
	"encoding/json"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/ds454/auth/src/log"
)

type AuthConfig struct {
	TtlS             int64  `env:"AUTH_STORE_TTL_S,required"`
	SessionKeyLen    int32  `env:"AUTH_SESSION_TOKEN_LENGTH,required"`
	SessionTokenKey  string `env:"SESSION_TOKEN_KEY,required"`
	EnableReflection int32  `env:"ENABLE_REFLECTION,required"`
	GrpcListenPort   int    `env:"GRPC_PORT,required"`
	UuidMetadataKey  string `env:"UUID_METADATA_KEY,required"`
	RedisPort        string `env:"REDIS_PORT,required"`
	RedisHost        string `env:"REDIS_HOST,required"`
	RedisDB          int    `env:"REDIS_DB,required"`
	RedisPassword    string `env:"REDIS_PASSWORD,required"`
}

var config AuthConfig
var configOnce sync.Once

func GetConfig() *AuthConfig {
	configOnce.Do(func() {
		log.Debug("Loading config from environment variables...")

		err := env.Parse(&config)
		if err != nil {
			log.Error("Unable to load server config from environment:")
			log.Abort("%v", err)
		}

		log.Debug("Done:")
		val, err := json.MarshalIndent(config, "", "\t")
		if err != nil {
			log.Abort("Problem encoding server config as a string")
		}

		log.Info("Config: %s", val)
	})

	return &config
}
