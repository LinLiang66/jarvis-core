package database

import (
	"context"
	"log"
	"time"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/store"
)

// needsSyncSeed 空库必须同步种子；已有数据的库可后台补菜单/存储。
func needsSyncSeed(ctx context.Context, s *store.Stores) (bool, error) {
	var userN int64
	if err := s.SysUser.DB.WithContext(ctx).Model(&model.SysUser{}).Limit(1).Count(&userN).Error; err != nil {
		return true, err
	}
	return userN == 0, nil
}

func startBackgroundSeed(cfg *config.Config, s *store.Stores) {
	go func() {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := runExistingDBSeed(ctx, s, cfg); err != nil {
			log.Printf("[seed] background failed: %v", err)
			return
		}
		log.Printf("[seed] background finished in %s", time.Since(start).Round(time.Millisecond))
	}()
}

func runExistingDBSeed(ctx context.Context, s *store.Stores, cfg *config.Config) error {
	if err := seedIncrementalMenus(ctx, s); err != nil {
		return err
	}
	if err := seedDefaultStorage(ctx, cfg, s); err != nil {
		return err
	}
	var userN, menuN int64
	if err := s.SysUser.DB.WithContext(ctx).Model(&model.SysUser{}).Limit(1).Count(&userN).Error; err != nil {
		return err
	}
	if err := s.SysMenu.DB.WithContext(ctx).Model(&model.SysMenu{}).Limit(1).Count(&menuN).Error; err != nil {
		return err
	}
	if userN > 0 && menuN == 0 {
		return seedMenusOnly(ctx, s)
	}
	return nil
}
