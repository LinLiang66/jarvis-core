package mysql

import (
	"os"
	"strconv"
	"strings"
)

// Config MySQL 连接配置（.env / 环境变量）。
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	Charset      string
	ShowSQL      bool
	LogLevel     int
	MaxOpenConns int
	MaxIdleConns int
}

func ConfigFromEnv() Config {
	port := 3306
	if v := strings.TrimSpace(os.Getenv("MYSQL_PORT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			port = n
		}
	}
	logLevel := 2
	if v := strings.TrimSpace(os.Getenv("MYSQL_LOG_LEVEL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			logLevel = n
		}
	}
	maxOpen := 20
	if v := strings.TrimSpace(os.Getenv("MYSQL_MAX_OPEN_CONNS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpen = n
		}
	}
	maxIdle := 10
	if v := strings.TrimSpace(os.Getenv("MYSQL_MAX_IDLE_CONNS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxIdle = n
		}
	}
	charset := strings.TrimSpace(os.Getenv("MYSQL_CHARSET"))
	if charset == "" {
		charset = "utf8mb4"
	}
	return Config{
		Host:         strings.TrimSpace(os.Getenv("MYSQL_HOST")),
		Port:         port,
		User:         strings.TrimSpace(os.Getenv("MYSQL_USER")),
		Password:     os.Getenv("MYSQL_PASSWORD"),
		DBName:       strings.TrimSpace(os.Getenv("MYSQL_DATABASE")),
		Charset:      charset,
		ShowSQL:      strings.EqualFold(os.Getenv("MYSQL_SHOW_SQL"), "true"),
		LogLevel:     logLevel,
		MaxOpenConns: maxOpen,
		MaxIdleConns: maxIdle,
	}
}

func (c Config) Enabled() bool {
	return c.Host != "" && c.DBName != ""
}

func (c Config) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.DBName +
		"?charset=" + c.Charset + "&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s"
}

// DefaultConfig 本地开发默认连接（与旧版 MYSQL_DSN 默认值一致）。
func DefaultConfig() Config {
	return Config{
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "root",
		Password: "root",
		DBName:   "jarvis_scheduler",
		Charset:  "utf8mb4",
	}
}
