package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App  `yaml:"app"`
		HTTP `yaml:"http"`
		Log
		Postgres `yaml:"postgres"`
		Redis
		Elastic
		Kafka
		CORS
		AuthSettings
		ApiKeys
	}

	App struct {
		Name    string `  yaml:"name"`
		Version string `  yaml:"version"`
	}

	HTTP struct {
		Host string `  env:"HTTP_HOST"`
		Port string `  yaml:"port"`
	}

	Log struct {
		Level string `  env:"LOG_LEVEL"`
	}

	Postgres struct {
		PoolMax int    `  yaml:"pool_max"`
		URL     string `                  env:"PG_URL"`
	}

	Redis struct {
		SentinelAddrs  string `  env:"REDIS_SENTINEL_ADDRS"`
		MasterName     string `  env:"REDIS_MASTER_NAME"`
		MasterPassword string `  env:"REDIS_MASTER_PASSWORD"`
	}

	Elastic struct {
		Addrs                 string `  env:"ELASTIC_ADDRS"`
		MaxIdleConnsPerHost   string `  env:"ELASTIC_MAX_IDLE_CONNS_PER_HOST"`
		ResponseHeaderTimeout string `  env:"ELASTIC_RESPONSE_HEADER_TIMEOUT"`
		DialTimeout           string `  env:"ELASTIC_DIAL_TIMEOUT"`
	}

	Kafka struct {
		Brokers string `  env:"KAFKA_BROKERS"`
	}

	CORS struct {
		AllowedOrigins []string `env:"ALLOWED_ORIGINS"`
	}

	AuthSettings struct {
		Salt       string        `  env:"AUTH_SALT"`
		SigningKey string        `  env:"JWT_SIGNING_KEY"`
		TokenTTL   time.Duration `  env:"TOKEN_TTL"`
	}

	ApiKeys struct {
		AppId         string `  env:"APP_ID"`
		AppKey        string `  env:"APP_KEY"`
		PrivateAppId  string `  env:"PRIVATE_APP_ID"`
		PrivateAppKey string `  env:"PRIVATE_APP_KEY"`
	}
)

func NewConfig(configPath string) (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("yml config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
