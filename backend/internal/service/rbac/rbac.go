package rbac

import (
	"context"

	"gorm.io/gorm"

	"jarvis/backend/internal/model"
)

// PermissionsForUser 聚合用户权限（type=3 按钮 + type=2 带 permission 的菜单）
func PermissionsForUser(ctx context.Context, db *gorm.DB, userID int64) ([]string, error) {
	var user model.SysUser
	if err := db.WithContext(ctx).Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}
	for _, r := range user.Roles {
		if r.Code == model.RoleAdminCode && r.Status == "0" {
			return []string{"*:*:*"}, nil
		}
	}
	var roleIDs []int64
	for _, r := range user.Roles {
		if r.Status == "0" {
			roleIDs = append(roleIDs, r.ID)
		}
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var menus []model.SysMenu
	err := db.WithContext(ctx).
		Joins("JOIN sys_role_menus rm ON rm.sys_menu_id = sys_menus.id").
		Where("rm.sys_role_id IN ? AND sys_menus.status = ?", roleIDs, "0").
		Where("sys_menus.permission <> ''").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var perms []string
	for _, m := range menus {
		if m.Permission == "" {
			continue
		}
		if _, ok := seen[m.Permission]; ok {
			continue
		}
		seen[m.Permission] = struct{}{}
		perms = append(perms, m.Permission)
	}
	return perms, nil
}

func RoleCodes(user model.SysUser) []string {
	var codes []string
	for _, r := range user.Roles {
		if r.Status == "0" {
			codes = append(codes, r.Code)
		}
	}
	return codes
}

func IsSuperAdmin(user model.SysUser) bool {
	if user.IsSuperAdmin {
		return true
	}
	for _, r := range user.Roles {
		if r.Code == model.RoleAdminCode && r.Status == "0" {
			return true
		}
	}
	return false
}
