package store

import "gorm.io/gorm"

type Stores struct {
	SysUser       *SysUserRepository
	SysRole       *SysRoleRepository
	SysMenu       *SysMenuRepository
	SysDict       *SysDictRepository
	OpenApp       *OpenAppRepository
	OpenAPIStat   *OpenAPIStatRepository
	OpenAPIAction *OpenAPIActionRepository
}

func NewStores(db *gorm.DB) *Stores {
	return &Stores{
		SysUser:       NewSysUserRepository(db),
		SysRole:       NewSysRoleRepository(db),
		SysMenu:       NewSysMenuRepository(db),
		SysDict:       NewSysDictRepository(db),
		OpenApp:       NewOpenAppRepository(db),
		OpenAPIStat:   NewOpenAPIStatRepository(db),
		OpenAPIAction: NewOpenAPIActionRepository(db),
	}
}
