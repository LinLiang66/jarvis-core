package database

import (
	"context"
	"log"
	"strings"

	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/store"
)

type menuPatch struct {
	Dir      model.SysMenu
	Children []model.SysMenu
}

// seedIncrementalMenus 按路由 path 增量补全菜单（旧库升级时使用，启动后后台执行）。
func seedIncrementalMenus(ctx context.Context, s *store.Stores) error {
	paths := incrementalMenuPatchPaths()
	upToDate, err := incrementalMenusApplied(ctx, s, paths)
	if err != nil {
		return err
	}
	if upToDate {
		return nil
	}

	all, err := s.SysMenu.All(ctx)
	if err != nil {
		return err
	}
	byPath := make(map[string]model.SysMenu, len(all))
	for _, m := range all {
		if m.Path != "" {
			byPath[m.Path] = m
		}
	}

	var newIDs []int64
	for _, patch := range incrementalMenuPatches() {
		ids, err := ensureMenuPatch(ctx, s, byPath, patch)
		if err != nil {
			return err
		}
		newIDs = append(newIDs, ids...)
	}
	if len(newIDs) == 0 {
		return nil
	}

	var adminRole model.SysRole
	if err := s.SysRole.DB.WithContext(ctx).Where("code = ?", model.RoleAdminCode).First(&adminRole).Error; err != nil {
		log.Printf("[seed] skip role menu append: admin role not found: %v", err)
		return nil
	}
	if err := s.SysRole.AppendMenuIDs(ctx, adminRole.ID, newIDs); err != nil {
		return err
	}
	log.Printf("[seed] incremental menus added: %d item(s), linked to role %s", len(newIDs), adminRole.Code)
	return nil
}

func ensureMenuPatch(ctx context.Context, s *store.Stores, byPath map[string]model.SysMenu, patch menuPatch) ([]int64, error) {
	var newIDs []int64

	parent, ok := byPath[patch.Dir.Path]
	if !ok {
		dir := patch.Dir
		dir.Status = "0"
		if err := s.SysMenu.Create(ctx, &dir); err != nil {
			return nil, err
		}
		byPath[dir.Path] = dir
		parent = dir
		newIDs = append(newIDs, dir.ID)
	}

	for _, childTpl := range patch.Children {
		if existing, exists := byPath[childTpl.Path]; exists {
			if strings.TrimSpace(existing.Component) == "" && strings.TrimSpace(childTpl.Component) != "" {
				existing.Component = childTpl.Component
				if err := s.SysMenu.Save(ctx, &existing); err != nil {
					return nil, err
				}
				byPath[childTpl.Path] = existing
			}
			continue
		}
		child := childTpl
		child.ParentID = parent.ID
		child.Status = "0"
		if err := s.SysMenu.Create(ctx, &child); err != nil {
			return nil, err
		}
		byPath[child.Path] = child
		newIDs = append(newIDs, child.ID)
	}
	return newIDs, nil
}

func incrementalMenuPatchPaths() []string {
	var paths []string
	for _, patch := range incrementalMenuPatches() {
		if patch.Dir.Path != "" {
			paths = append(paths, patch.Dir.Path)
		}
		for _, child := range patch.Children {
			if child.Path != "" {
				paths = append(paths, child.Path)
			}
		}
	}
	return paths
}

func incrementalMenusApplied(ctx context.Context, s *store.Stores, paths []string) (bool, error) {
	if len(paths) == 0 {
		return true, nil
	}
	var n int64
	err := s.SysMenu.DB.WithContext(ctx).Model(&model.SysMenu{}).Where("path IN ?", paths).Count(&n).Error
	return n >= int64(len(paths)), err
}

func incrementalMenuPatches() []menuPatch {
	return []menuPatch{
		{
			Dir: model.SysMenu{
				Type: 1, Title: "开放平台", Name: "开放平台", Path: "/openplatform",
				Component: "Layout", Redirect: "/openplatform/app", Icon: "Connection",
				Sort: 5, AlwaysShow: true,
			},
			Children: []model.SysMenu{
				{Type: 2, Title: "应用管理", Name: "应用管理", Path: "/openplatform/app", Component: "openplatform/app/index", Permission: "openapp:query", Sort: 1, KeepAlive: true},
				{Type: 2, Title: "接口管理", Name: "接口管理", Path: "/openplatform/api", Component: "openplatform/api/index", Permission: "openaction:query", Sort: 2, KeepAlive: true},
				{Type: 2, Title: "接口文档", Name: "接口文档", Path: "/openplatform/docs", Component: "openplatform/docs/index", Permission: "openaction:query", Sort: 3, KeepAlive: true},
				{Type: 2, Title: "调用统计", Name: "调用统计", Path: "/openplatform/stat", Component: "openplatform/stat/index", Permission: "openstat:query", Sort: 4, KeepAlive: true},
			},
		},
		{
			Dir: model.SysMenu{
				Type: 1, Title: "系统管理", Name: "系统管理", Path: "/system",
				Component: "Layout", Redirect: "/system/user", Icon: "Setting",
				Sort: 10, AlwaysShow: true,
			},
			Children: []model.SysMenu{
				{Type: 2, Title: "存储配置", Name: "存储配置", Path: "/system/storage", Component: "system/storage/index", Permission: "module_system:storage:query", Sort: 5, KeepAlive: true},
				{Type: 2, Title: "文件管理", Name: "文件管理", Path: "/system/file", Component: "system/file/index", Permission: "module_system:file:query", Sort: 6, KeepAlive: true},
			},
		},
	}
}
