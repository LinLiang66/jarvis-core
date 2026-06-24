package database

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// applySchemaPatches 幂等修正已有库结构，对齐模板仓库 schema 变更。
func applySchemaPatches(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return nil
	}
	dialector := db.Dialector.Name()
	if dialector == "mysql" {
		return db.WithContext(ctx).Exec("ALTER TABLE sys_menus MODIFY COLUMN icon TEXT").Error
	}
	if dialector == "postgres" {
		return db.WithContext(ctx).Exec("ALTER TABLE sys_menus ALTER COLUMN icon TYPE TEXT").Error
	}
	if strings.Contains(dialector, "sqlite") {
		// SQLite 字符串列本身无长度限制，AutoMigrate 已足够。
		return nil
	}
	return nil
}
