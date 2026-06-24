# -*- coding: utf-8 -*-
"""Fix UTF-8 Chinese in seed and openplatform meta files."""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

ACTIONS_META = """package openplatform

func init() {
	registerBuiltinActionMetas()
}

func registerBuiltinActionMetas() {
	RegisterActionMetaWithTypes(ActionMeta{
		Action: ActionGetPublicKey, Title: "获取 Token 与公钥", Category: "握手",
		Description: "获取网关会话 token 与服务端 RSA 公钥",
		Encrypted:   false, Billable: false, Sort: 10,
	}, docPublicKeyReq{Timestamp: "", Data: "{}"}, docPublicKeyResp{})

	RegisterActionMetaWithTypes(ActionMeta{
		Action: ActionCreateSecretKey, Title: "3DES 密钥交换", Category: "握手",
		Description: "使用应用 RSA 私钥加密 clientPart，交换 serverPart，合成 3DES 会话密钥",
		Encrypted:   false, Billable: false, Sort: 20,
	}, docSecretKeyReq{}, docSecretKeyResp{})

	RegisterActionMetaWithTypes(ActionMeta{
		Action: ActionEcho, Title: "Echo 演示", Category: "演示",
		Description: "加密回显请求 JSON，用于联调与示例",
		Encrypted:   true, Billable: true, Sort: 30,
	}, docEchoReq{}, docEchoResp{Action: ActionEcho, Message: "pong"})
}
"""

SEED_PATCH = """package database

import (
	"context"
	"log"

	"jarvis/backend/internal/model"
	"jarvis/backend/internal/store"
)

type menuPatch struct {
	Dir      model.SysMenu
	Children []model.SysMenu
}

// seedIncrementalMenus 按路由 path 增量补全菜单（旧库升级时使用）
func seedIncrementalMenus(ctx context.Context, s *store.Stores) error {
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
		if _, exists := byPath[childTpl.Path]; exists {
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
	}
}
"""

SEED_SYS = """package database

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"jarvis/backend/internal/model"
	"jarvis/backend/internal/store"
)

func seedSystem(ctx context.Context, s *store.Stores) error {
	if err := migrateSys(ctx, s); err != nil {
		return err
	}
	if err := seedIncrementalMenus(ctx, s); err != nil {
		return err
	}
	var userN, menuN int64
	s.SysUser.DB.WithContext(ctx).Model(&model.SysUser{}).Count(&userN)
	s.SysMenu.DB.WithContext(ctx).Model(&model.SysMenu{}).Count(&menuN)
	if userN > 0 && menuN > 0 {
		return nil
	}
	if userN > 0 && menuN == 0 {
		return seedMenusOnly(ctx, s)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	adminRole := model.SysRole{Code: model.RoleAdminCode, Name: "超级管理员", Status: "0", Sort: 0, IsSystem: true}
	userRole := model.SysRole{Code: "role_user", Name: "普通用户", Status: "0", Sort: 1}
	if err := s.SysRole.Create(ctx, &adminRole); err != nil {
		return err
	}
	if err := s.SysRole.Create(ctx, &userRole); err != nil {
		return err
	}
	admin := model.SysUser{
		Username: "admin", Password: string(hash), Nickname: "管理员",
		Status: "0", IsSuperAdmin: true,
	}
	if err := s.SysUser.Create(ctx, &admin); err != nil {
		return err
	}
	_ = s.SysUser.ReplaceRoles(ctx, admin.ID, []int64{adminRole.ID})

	var menuIDs []int64
	createDir := func(m model.SysMenu, children []model.SysMenu) error {
		m.Status = "0"
		if err := s.SysMenu.Create(ctx, &m); err != nil {
			return err
		}
		menuIDs = append(menuIDs, m.ID)
		for _, c := range children {
			c.ParentID = m.ID
			c.Status = "0"
			if err := s.SysMenu.Create(ctx, &c); err != nil {
				return err
			}
			menuIDs = append(menuIDs, c.ID)
		}
		return nil
	}
	if err := seedMenuTree(createDir); err != nil {
		return err
	}
	_ = s.SysRole.ReplaceMenus(ctx, adminRole.ID, menuIDs)

	dt := model.SysDictType{Name: "通用状态", Code: "common_status", Status: "0", Sort: 1, IsSystem: true}
	if err := s.SysDict.CreateType(ctx, &dt); err != nil {
		return err
	}
	for _, d := range []model.SysDictData{
		{TypeID: dt.ID, Label: "正常", Value: "0", Status: "0", Sort: 1},
		{TypeID: dt.ID, Label: "停用", Value: "1", Status: "0", Sort: 2},
	} {
		if err := s.SysDict.CreateData(ctx, &d); err != nil {
			return err
		}
	}
	statusAlias := model.SysDictType{Name: "状态", Code: "STATUS", Status: "0", Sort: 1, IsSystem: true}
	_ = s.SysDict.CreateType(ctx, &statusAlias)
	_ = s.SysDict.CreateData(ctx, &model.SysDictData{TypeID: statusAlias.ID, Label: "正常", Value: "0", Status: "0", Sort: 1})
	_ = s.SysDict.CreateData(ctx, &model.SysDictData{TypeID: statusAlias.ID, Label: "停用", Value: "1", Status: "0", Sort: 2})

	gender := model.SysDictType{Name: "性别", Code: "GENDER", Status: "0", Sort: 2}
	_ = s.SysDict.CreateType(ctx, &gender)
	_ = s.SysDict.CreateData(ctx, &model.SysDictData{TypeID: gender.ID, Label: "男", Value: "1", Status: "0", Sort: 1})
	_ = s.SysDict.CreateData(ctx, &model.SysDictData{TypeID: gender.ID, Label: "女", Value: "2", Status: "0", Sort: 2})
	return nil
}

func seedMenusOnly(ctx context.Context, s *store.Stores) error {
	var adminRole model.SysRole
	if err := s.SysRole.DB.WithContext(ctx).Where("code = ?", model.RoleAdminCode).First(&adminRole).Error; err != nil {
		return err
	}
	var menuIDs []int64
	createDir := func(m model.SysMenu, children []model.SysMenu) error {
		m.Status = "0"
		if err := s.SysMenu.Create(ctx, &m); err != nil {
			return err
		}
		menuIDs = append(menuIDs, m.ID)
		for _, c := range children {
			c.ParentID = m.ID
			c.Status = "0"
			if err := s.SysMenu.Create(ctx, &c); err != nil {
				return err
			}
			menuIDs = append(menuIDs, c.ID)
		}
		return nil
	}
	if err := seedMenuTree(createDir); err != nil {
		return err
	}
	return s.SysRole.ReplaceMenus(ctx, adminRole.ID, menuIDs)
}

func seedMenuTree(createDir func(m model.SysMenu, children []model.SysMenu) error) error {
	if err := createDir(
		model.SysMenu{Type: 1, Title: "工作台", Path: "/dashboard", Component: "Layout", Redirect: "/dashboard/index", Icon: "HomeFilled", Sort: 1, Affix: true},
		[]model.SysMenu{{Type: 2, Title: "工作台", Path: "/dashboard/index", Component: "dashboard/index", Icon: "HomeFilled", Sort: 1, KeepAlive: true, Affix: true}},
	); err != nil {
		return err
	}
	if err := createDir(
		model.SysMenu{Type: 1, Title: "开放平台", Path: "/openplatform", Component: "Layout", Redirect: "/openplatform/app", Icon: "Connection", Sort: 5, AlwaysShow: true},
		[]model.SysMenu{
			{Type: 2, Title: "应用管理", Path: "/openplatform/app", Component: "openplatform/app/index", Permission: "openapp:query", Sort: 1, KeepAlive: true},
			{Type: 2, Title: "接口管理", Path: "/openplatform/api", Component: "openplatform/api/index", Permission: "openaction:query", Sort: 2, KeepAlive: true},
			{Type: 2, Title: "接口文档", Path: "/openplatform/docs", Component: "openplatform/docs/index", Permission: "openaction:query", Sort: 3, KeepAlive: true},
			{Type: 2, Title: "调用统计", Path: "/openplatform/stat", Component: "openplatform/stat/index", Permission: "openstat:query", Sort: 4, KeepAlive: true},
		},
	); err != nil {
		return err
	}
	return createDir(
		model.SysMenu{Type: 1, Title: "系统管理", Path: "/system", Component: "Layout", Redirect: "/system/user", Icon: "Setting", Sort: 10, AlwaysShow: true},
		[]model.SysMenu{
			{Type: 2, Title: "用户管理", Path: "/system/user", Component: "system/user/index", Permission: "module_system:user:query", Sort: 1, KeepAlive: true},
			{Type: 2, Title: "角色管理", Path: "/system/role", Component: "system/role/index", Permission: "module_system:role:query", Sort: 2, KeepAlive: true},
			{Type: 2, Title: "菜单管理", Path: "/system/menu", Component: "system/menu/index", Permission: "module_system:menu:query", Sort: 3, KeepAlive: true},
			{Type: 2, Title: "字典管理", Path: "/system/dict", Component: "system/dict/index", Permission: "module_system:dict:query", Sort: 4, KeepAlive: true},
		},
	)
}

func migrateSys(ctx context.Context, s *store.Stores) error {
	if err := errorsJoin(
		s.SysUser.AutoMigrate(ctx),
		s.SysRole.AutoMigrate(ctx),
		s.SysMenu.AutoMigrate(ctx),
		s.SysDict.AutoMigrate(ctx),
		s.OpenApp.AutoMigrate(ctx),
		s.OpenAPIStat.AutoMigrate(ctx),
		s.OpenAPIAction.AutoMigrate(ctx),
	); err != nil {
		return err
	}
	return applySchemaPatches(ctx, s.SysMenu.DB)
}
"""


def write(rel: str, content: str) -> None:
    path = ROOT / rel
    path.write_text(content.strip() + "\n", encoding="utf-8", newline="\n")
    print("fixed", rel)


def main() -> None:
    write("backend/internal/service/openplatform/actions_meta.go", ACTIONS_META)
    write("backend/internal/database/seed_patch_menus.go", SEED_PATCH)
    write("backend/internal/database/seed_sys.go", SEED_SYS)


if __name__ == "__main__":
    main()
