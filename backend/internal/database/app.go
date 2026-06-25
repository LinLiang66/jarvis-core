package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/infra/mysql"
	infraredis "jarvis-core/backend/internal/infra/redis"
	"jarvis-core/backend/internal/store"
)

type App struct {
	DB      *gorm.DB
	Stores  *store.Stores
	Session *store.SessionCache
	Driver  string
	Redis   *infraredis.Client
}

func Open(ctx context.Context, cfg *config.Config) (*App, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	boot := newBootTimer()

	db, driver, err := openGORM(cfg)
	if err != nil {
		return nil, err
	}
	boot.mark("mysql_connect")

	stores := store.NewStores(db)
	ensureUploadDir(cfg)

	var migrateErr error
	var rdb *infraredis.Client
	var redisErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if !cfg.DBAutoMigrate {
			return
		}
		migrateErr = migrateAll(ctx, stores)
	}()
	go func() {
		defer wg.Done()
		if !cfg.RedisEnable {
			return
		}
		rdb, redisErr = infraredis.InitWithContext(ctx, infraredis.ConfigFromEnv())
	}()
	wg.Wait()
	boot.mark("migrate_and_redis_parallel")

	if migrateErr != nil {
		return nil, migrateErr
	}

	app := &App{
		DB:      db,
		Stores:  stores,
		Driver:  driver,
		Session: store.NewSessionCache(nil, cfg.JWTExpire),
	}

	if cfg.RedisEnable {
		if redisErr != nil {
			if cfg.RedisRequired {
				return nil, fmt.Errorf("redis init: %w", redisErr)
			}
			log.Printf("[warn] redis unavailable: %v", redisErr)
		} else {
			app.Redis = rdb
			app.Session = store.NewSessionCache(rdb, cfg.JWTExpire)
			log.Printf("redis connected: %s", infraredis.ConfigFromEnv().Addr)
		}
	}

	syncSeed, err := needsSyncSeed(ctx, stores)
	if err != nil {
		return nil, err
	}
	if syncSeed {
		if err := seedSystem(ctx, stores); err != nil {
			return nil, err
		}
		if err := seedDefaultStorage(ctx, cfg, stores); err != nil {
			return nil, err
		}
		boot.mark("seed_sync")
	} else {
		startBackgroundSeed(cfg, stores)
		boot.mark("seed_async")
	}

	boot.summary()
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
