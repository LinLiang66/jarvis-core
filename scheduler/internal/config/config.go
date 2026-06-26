package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"jarvis-core/scheduler/internal/infra/mysql"
	infraredis "jarvis-core/scheduler/internal/infra/redis"
)

type Config struct {
	ServerAddr     string
	MySQLDSN       string
	MySQL          mysql.Config
	RedisEnable    bool
	RedisRequired  bool
	Redis          infraredis.Config
	AdminToken     string
	WorkerToken    string
	PollTimeout    time.Duration
	PollInterval   time.Duration
	WorkerTTL             time.Duration
	RunningLockTTL        time.Duration
	InstanceClaimTimeout  time.Duration
	InstanceScanInterval  time.Duration
}

func Load() *Config {
	mysqlCfg := mysql.ConfigFromEnv()
	redisCfg := infraredis.ConfigFromEnv()
	return &Config{
		ServerAddr:     getEnv("SERVER_ADDR", ":9000"),
		MySQLDSN:       resolveMySQLDSN(mysqlCfg),
		MySQL:          mysqlCfg,
		RedisEnable:    getEnvBool("REDIS_ENABLE", true),
		RedisRequired:  getEnvBool("REDIS_REQUIRED", true),
		Redis:          redisCfg,
		AdminToken:     getEnv("ADMIN_TOKEN", "sched-admin-dev"),
		WorkerToken:    getEnv("WORKER_TOKEN", "sched-worker-dev"),
		PollTimeout:    time.Duration(getEnvInt("POLL_TIMEOUT_SEC", 30)) * time.Second,
		PollInterval:   time.Duration(getEnvInt("POLL_INTERVAL_MS", 1000)) * time.Millisecond,
		WorkerTTL:            time.Duration(getEnvInt("WORKER_TTL_SEC", 90)) * time.Second,
		RunningLockTTL:       time.Duration(getEnvInt("RUNNING_LOCK_TTL_SEC", 3600)) * time.Second,
		InstanceClaimTimeout: time.Duration(getEnvInt("INSTANCE_CLAIM_TIMEOUT_SEC", 120)) * time.Second,
		InstanceScanInterval: time.Duration(getEnvInt("INSTANCE_SCAN_INTERVAL_SEC", 30)) * time.Second,
	}
}

func resolveMySQLDSN(cfg mysql.Config) string {
	if dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN")); dsn != "" {
		return dsn
	}
	if cfg.Enabled() {
		return cfg.DSN()
	}
	return mysql.DefaultConfig().DSN()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes")
}
