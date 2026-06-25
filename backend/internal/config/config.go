package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr            string
	DBPath          string
	JWTSecret       string
	JWTExpire       time.Duration
	JWTRefreshDays  int
	RedisEnable     bool
	RedisRequired   bool
	RedisAddr       string
	LogLevel        string
	UploadDir              string
	StaticURLPrefix        string
	PublicBaseURL          string
	ImageCompressEnable    bool
	ImageCompressMaxDim    int
	ImageCompressQuality   int
	ImageCompressMinBytes  int
	ImageCompressMaxInput  int
}

func Load() *Config {
	jwtHours := getEnvInt("JWT_EXPIRE_HOURS", 24)
	return &Config{
		Addr:            getEnv("SERVER_ADDR", ":8000"),
		DBPath:          getEnv("DB_PATH", "./data/app.db"),
		JWTSecret:       getEnv("JWT_SECRET", "jarvis-core-dev-secret"),
		JWTExpire:       time.Duration(jwtHours) * time.Hour,
		JWTRefreshDays:  getEnvInt("JWT_REFRESH_DAYS", 7),
		RedisEnable:     getEnvBool("REDIS_ENABLE", true),
		RedisRequired:   getEnvBool("REDIS_REQUIRED", false),
		RedisAddr:       getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		UploadDir:             getEnv("UPLOAD_DIR", "./data/uploads"),
		StaticURLPrefix:       getEnv("STATIC_URL_PREFIX", "/static"),
		PublicBaseURL:         getEnv("PUBLIC_BASE_URL", "http://127.0.0.1:8000"),
		ImageCompressEnable:   getEnvBool("IMAGE_COMPRESS_ENABLE", true),
		ImageCompressMaxDim:   getEnvInt("IMAGE_COMPRESS_MAX_DIM", 1920),
		ImageCompressQuality:  getEnvInt("IMAGE_COMPRESS_QUALITY", 85),
		ImageCompressMinBytes: getEnvInt("IMAGE_COMPRESS_MIN_BYTES", 100*1024),
		ImageCompressMaxInput: getEnvInt("IMAGE_COMPRESS_MAX_INPUT", 20*1024*1024),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
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
