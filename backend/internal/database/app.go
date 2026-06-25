package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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
	prof := newStartupProfiler()

	t := time.Now()
	db, driver, err := openGORM(cfg)
	if err != nil {
		return nil, err
	}
	prof.step("mysql", t)

	stores := store.NewStores(db)
	app := &App{
		DB:      db,
		Stores:  stores,
		Driver:  driver,
		Session: store.NewSessionCache(nil, cfg.JWTExpire),
	}

	type redisResult struct {
		client *infraredis.Client
		err    error
	}
	redisCh := make(chan redisResult, 1)
	if cfg.RedisEnable {
		go func() {
			rt := time.Now()
			rdb, err := infraredis.InitWithContext(ctx, infraredis.ConfigFromEnv())
			prof.step("redis", rt)
			redisCh <- redisResult{client: rdb, err: err}
		}()
	} else {
		close(redisCh)
	}

	t = time.Now()
	if cfg.DBAutoMigrate {
		if err := migrateAll(ctx, stores); err != nil {
			return nil, err
		}
	}
	prof.step("migrate", t)

	t = time.Now()
	ensureUploadDir(cfg)
	if err := seedSystemRequired(ctx, stores); err != nil {
		return nil, err
	}
	prof.step("seed", t)

	if cfg.RedisEnable {
		res := <-redisCh
		if res.err != nil {
			if cfg.RedisRequired {
				return nil, fmt.Errorf("redis init: %w", res.err)
			}
			log.Printf("[warn] redis unavailable: %v", res.err)
		} else {
			app.Redis = res.client
			app.Session = store.NewSessionCache(res.client, cfg.JWTExpire)
			log.Printf("redis connected: %s", infraredis.ConfigFromEnv().Addr)
		}
	}

	startBackgroundSeed(cfg, stores)
	prof.finish()
	return app, nil
}

func startBackgroundSeed(cfg *config.Config, s *store.Stores) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		t := time.Now()
		if err := seedIncrementalMenus(ctx, s); err != nil {
			log.Printf("[seed] incremental menus: %v", err)
		}
		if err := seedDefaultStorage(ctx, cfg, s); err != nil {
			log.Printf("[seed] default storage: %v", err)
		}
		log.Printf("[startup] background-seed done in %v", time.Since(t))
	}()
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
