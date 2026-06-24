package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/infra/mysql"
	infraredis "jarvis-core/backend/internal/infra/redis"
	"jarvis-core/backend/internal/store"
)

type App struct {
	DB      *gorm.DB
	Redis   *infraredis.Client
	Stores  *store.Stores
	Session *store.SessionCache
	Driver  string
}

func Open(ctx context.Context, cfg *config.Config) (*App, error) {
	db, driver, err := openGORM(cfg)
	if err != nil {
		return nil, err
	}

	stores := store.NewStores(db)
	if err := migrateAll(ctx, stores); err != nil {
		return nil, err
	}
	if err := seedSystem(ctx, stores); err != nil {
		return nil, err
	}

	app := &App{
		DB:      db,
		Stores:  stores,
		Driver:  driver,
		Session: store.NewSessionCache(nil, cfg.JWTExpire),
	}

	if cfg.RedisEnable {
		rdb, err := infraredis.InitWithContext(ctx, infraredis.ConfigFromEnv())
		if err != nil {
			if cfg.RedisRequired {
				return nil, fmt.Errorf("redis init: %w", err)
			}
			log.Printf("[warn] redis unavailable: %v", err)
		} else {
			app.Redis = rdb
			app.Session = store.NewSessionCache(rdb, cfg.JWTExpire)
			log.Printf("redis connected: %s", infraredis.ConfigFromEnv().Addr)
		}
	}

	return app, nil
}

func openGORM(cfg *config.Config) (*gorm.DB, string, error) {
	mysqlCfg := mysql.ConfigFromEnv()
	if mysqlCfg.Enabled() {
		dbWrap, ok, err := mysql.Init(mysqlCfg)
		if err != nil {
			return nil, "", err
		}
		if ok && dbWrap != nil {
			sqlDB, err := dbWrap.DB.DB()
			if err != nil {
				return nil, "", err
			}
			if err := sqlDB.Ping(); err != nil {
				return nil, "", fmt.Errorf("mysql ping: %w", err)
			}
			log.Printf("database: mysql %s/%s", mysqlCfg.Host, mysqlCfg.DBName)
			return dbWrap.DB, "mysql", nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return nil, "", err
	}
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return nil, "", err
	}
	log.Printf("database: sqlite %s", cfg.DBPath)
	return db, "sqlite", nil
}

func migrateAll(ctx context.Context, s *store.Stores) error {
	return migrateSys(ctx, s)
}

func errorsJoin(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func (a *App) Health(ctx context.Context) map[string]any {
	out := map[string]any{
		"status":   "ok",
		"database": a.Driver,
		"mysql":    mysql.Available(),
		"redis":    infraredis.Available(),
	}
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			out["db_ping"] = sqlDB.PingContext(ctx) == nil
		}
	}
	return out
}
