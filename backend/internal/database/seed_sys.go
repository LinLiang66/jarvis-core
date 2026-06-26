package database

import (
	"context"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/store"
)

func seedSystem(ctx context.Context, s *store.Stores) error {
	var userN int64
	if err := s.SysUser.DB.WithContext(ctx).Model(&model.SysUser{}).Limit(1).Count(&userN).Error; err != nil {
		return err
	}
	if userN > 0 {
		return nil
	}
	var menuN int64
	if err := s.SysMenu.DB.WithContext(ctx).Model(&model.SysMenu{}).Limit(1).Count(&menuN).Error; err != nil {
		return err
	}
	if menuN > 0 {
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
	if err := createDir(
		model.SysMenu{Type: 1, Title: "任务调度", Path: "/scheduler", Component: "Layout", Redirect: "/scheduler/job", Icon: "Timer", Sort: 6, AlwaysShow: true},
		[]model.SysMenu{
			{Type: 2, Title: "任务管理", Path: "/scheduler/job", Component: "scheduler/job/index", Permission: "scheduler:job:query", Sort: 1, KeepAlive: true},
			{Type: 2, Title: "执行记录", Path: "/scheduler/instance", Component: "scheduler/instance/index", Permission: "scheduler:instance:query", Sort: 2, KeepAlive: true},
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
			{Type: 2, Title: "存储配置", Path: "/system/storage", Component: "system/storage/index", Permission: "module_system:storage:query", Sort: 5, KeepAlive: true},
			{Type: 2, Title: "文件管理", Path: "/system/file", Component: "system/file/index", Permission: "module_system:file:query", Sort: 6, KeepAlive: true},
		},
	)
}

func migrateSys(ctx context.Context, s *store.Stores) error {
	tasks := []func() error{
		func() error { return s.SysUser.AutoMigrate(ctx) },
		func() error { return s.SysRole.AutoMigrate(ctx) },
		func() error { return s.SysMenu.AutoMigrate(ctx) },
		func() error { return s.SysDict.AutoMigrate(ctx) },
		func() error { return s.SysStorage.AutoMigrate(ctx) },
		func() error { return s.SysFile.AutoMigrate(ctx) },
		func() error { return s.OpenApp.AutoMigrate(ctx) },
		func() error { return s.OpenAPIStat.AutoMigrate(ctx) },
		func() error { return s.OpenAPIAction.AutoMigrate(ctx) },
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))
	for _, task := range tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task(); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	// RepairDirTypes 依赖 sys_files 表，必须在 AutoMigrate 之后串行执行
	if err := s.SysFile.RepairDirTypes(ctx); err != nil {
		return err
	}
	return applySchemaPatches(ctx, s.SysMenu.DB)
}
