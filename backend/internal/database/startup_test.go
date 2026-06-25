//go:build cgo

package database

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/store"
)

func skipIfSQLiteUnavailable(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "probe.db")), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO") || strings.Contains(err.Error(), "cgo") {
			t.Skip("sqlite requires CGO on this platform")
		}
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
}

func TestStartupOpenSQLite(t *testing.T) {
	skipIfSQLiteUnavailable(t)
	dir := t.TempDir()
	t.Setenv("DB_PATH", filepath.Join(dir, "startup.db"))
	t.Setenv("REDIS_ENABLE", "false")
	t.Setenv("DB_AUTO_MIGRATE", "true")

	cfg := config.Load()
	start := time.Now()
	app, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	first := time.Since(start)
	if app.Driver != "sqlite" {
		t.Fatalf("expected sqlite, got %s", app.Driver)
	}
	t.Logf("cold Open=%v", first)
}

func TestStartupLegacyVsOptimized(t *testing.T) {
	skipIfSQLiteUnavailable(t)
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "bench.db")
	t.Setenv("DB_PATH", dbPath)
	t.Setenv("REDIS_ENABLE", "false")

	legacy := measureStartupSteps(t, dbPath, true)
	optimized := measureStartupSteps(t, dbPath, false)
	saved := legacy - optimized
	pct := float64(saved) / float64(legacy) * 100
	t.Logf("legacy=%v optimized=%v saved=%v (%.0f%%)", legacy, optimized, saved, pct)
	if optimized >= legacy {
		t.Fatalf("expected optimized faster than legacy, legacy=%v optimized=%v", legacy, optimized)
	}
}

func measureStartupSteps(t *testing.T, dbPath string, legacy bool) time.Duration {
	t.Helper()
	_ = os.Remove(dbPath)
	cfg := config.Load()
	cfg.DBPath = dbPath
	cfg.RedisEnable = false
	cfg.DBAutoMigrate = true

	db, _, err := openGORM(cfg)
	if err != nil {
		t.Fatal(err)
	}
	stores := store.NewStores(db)
	ctx := context.Background()

	start := time.Now()
	if legacy {
		if err := migrateSys(ctx, stores); err != nil {
			t.Fatal(err)
		}
		if err := migrateSys(ctx, stores); err != nil {
			t.Fatal(err)
		}
		if err := seedIncrementalMenus(ctx, stores); err != nil {
			t.Fatal(err)
		}
		if err := seedSystem(ctx, stores); err != nil {
			t.Fatal(err)
		}
		if err := seedDefaultStorage(ctx, cfg, stores); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := migrateSys(ctx, stores); err != nil {
			t.Fatal(err)
		}
		syncSeed, err := needsSyncSeed(ctx, stores)
		if err != nil {
			t.Fatal(err)
		}
		if syncSeed {
			if err := seedSystem(ctx, stores); err != nil {
				t.Fatal(err)
			}
		}
	}
	return time.Since(start)
}
