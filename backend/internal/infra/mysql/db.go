package mysql

import (
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局 GORM 实例（参考 tiny-pro-go/utils.Db）。
type DB struct {
	*gorm.DB
}

var (
	global      DB
	once        sync.Once
	initErr     error
	initialized bool
)

// Init 初始化 MySQL；未配置 MYSQL_HOST 时返回 (nil, false, nil)。
func Init(cfg Config) (*DB, bool, error) {
	if !cfg.Enabled() {
		return nil, false, nil
	}
	once.Do(func() {
		logMode := logger.Warn
		if cfg.ShowSQL {
			logMode = logger.Info
		}
		if cfg.LogLevel > 0 {
			logMode = logger.LogLevel(cfg.LogLevel)
		}
		db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logMode),
		})
		if err != nil {
			initErr = fmt.Errorf("mysql open: %w", err)
			return
		}
		sqlDB, err := db.DB()
		if err != nil {
			initErr = err
			return
		}
		if cfg.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		}
		if cfg.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		}
		global = DB{DB: db}
		initialized = true
	})
	if initErr != nil {
		return nil, false, initErr
	}
	if !initialized {
		return nil, false, nil
	}
	return &global, true, nil
}

// MustDB 返回已初始化的 DB。
func MustDB() *DB {
	if !initialized {
		panic("mysql not initialized")
	}
	return &global
}

// Available 是否已连接 MySQL。
func Available() bool {
	return initialized && global.DB != nil
}
