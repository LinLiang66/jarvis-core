package redis

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr        string
	Password    string
	DB          int
	PoolSize    int
	ReadTimeout time.Duration
}

func ConfigFromEnv() Config {
	cfg := Config{
		Addr:     strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_DB")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.DB = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_POOL_SIZE")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.PoolSize = n
		}
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 10
	}
	cfg.ReadTimeout = 3 * time.Second
	if v := strings.TrimSpace(os.Getenv("REDIS_READ_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ReadTimeout = d
		}
	}
	return cfg
}
