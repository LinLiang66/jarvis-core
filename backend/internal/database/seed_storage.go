package database

import (
	"context"
	"log"
	"os"
	"strings"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/service/storage"
	"jarvis-core/backend/internal/store"
)

func seedDefaultStorage(ctx context.Context, cfg *config.Config, s *store.Stores) error {
	var n int64
	if err := s.SysStorage.DB.WithContext(ctx).Model(&model.SysStorage{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	root := cfg.UploadDir
	if err := storage.EnsureLocalDir(root); err != nil {
		return err
	}
	domain := strings.TrimRight(cfg.PublicBaseURL, "/") + strings.TrimRight(cfg.StaticURLPrefix, "/") + "/local"
	row := model.SysStorage{
		Name:       "本地存储",
		Code:       "local",
		Type:       model.StorageTypeLocal,
		BucketName: root,
		Domain:     domain,
		IsDefault:  true,
		Status:     "0",
		Sort:       0,
		Description: "系统默认本地存储",
	}
	if err := s.SysStorage.Create(ctx, &row); err != nil {
		return err
	}
	log.Printf("[seed] default local storage created: code=local path=%s", root)
	return nil
}

func ensureUploadDir(cfg *config.Config) {
	_ = os.MkdirAll(cfg.UploadDir, 0o755)
}
