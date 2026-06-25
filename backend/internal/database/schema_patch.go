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
		if schemaColumnIsText(ctx, db, "sys_menus", "icon") {
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
		if err != nil || strings.EqualFold(strings.TrimSpace(dataType), "text") {
			return nil
		}
		return db.WithContext(ctx).Exec("ALTER TABLE sys_menus ALTER COLUMN icon TYPE TEXT").Error
	}
	if strings.Contains(dialector, "sqlite") {
		return nil
	}
	return nil
}

func schemaColumnIsText(ctx context.Context, db *gorm.DB, table, column string) bool {
	var dataType string
	err := db.WithContext(ctx).Raw(
		`SELECT DATA_TYPE FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table, column,
	).Scan(&dataType).Error
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "text", "mediumtext", "longtext":
		return true
	default:
		return false
	}
}
