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
		var dataType string
		err := db.WithContext(ctx).Raw(`
			SELECT DATA_TYPE FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_menus' AND COLUMN_NAME = 'icon'
			LIMIT 1
		`).Scan(&dataType).Error
		if err != nil {
			return err
		}
		switch strings.ToLower(dataType) {
		case "text", "longtext", "mediumtext":
			return nil
		}
		return db.WithContext(ctx).Exec("ALTER TABLE sys_menus MODIFY COLUMN icon TEXT").Error
	}
	if dialector == "postgres" {
		var dataType string
		err := db.WithContext(ctx).Raw(`
			SELECT data_type FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = 'sys_menus' AND column_name = 'icon'
			LIMIT 1
		`).Scan(&dataType).Error
		if err != nil {
			return err
		}
		if strings.EqualFold(dataType, "text") {
			return nil
		}
		return db.WithContext(ctx).Exec("ALTER TABLE sys_menus ALTER COLUMN icon TYPE TEXT").Error
	}
	if strings.Contains(dialector, "sqlite") {
		return nil
	}
	return nil
}
