package model

import (
	"time"

	"gorm.io/gorm"
)

const RoleAdminCode = "role_admin"

// SysUser 系统用户
type SysUser struct {
	ID           int64          `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Username     string         `gorm:"size:64;uniqueIndex" json:"username"`
	Password     string         `gorm:"size:128" json:"-"`
	Nickname     string         `gorm:"size:64" json:"nickname"`
	Phone        string         `gorm:"size:32" json:"phone"`
	Email        string         `gorm:"size:128" json:"email"`
	Avatar       string         `gorm:"size:500" json:"avatar"`
	Remark       string         `gorm:"size:512" json:"remark"`
	Status       string         `gorm:"size:1;default:0" json:"status"`
	Sort         int            `json:"sort"`
	DeptID       *int64         `json:"dept_id"`
	IsSuperAdmin bool           `gorm:"default:false" json:"is_super_admin"`
	Roles        []SysRole      `gorm:"many2many:sys_user_roles;" json:"-"`
}

// SysRole 角色
type SysRole struct {
	ID        int64          `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `gorm:"size:64;uniqueIndex" json:"code"`
	Name      string         `gorm:"size:64" json:"name"`
	Remark    string         `gorm:"size:512" json:"remark"`
	Status    string         `gorm:"size:1;default:0" json:"status"`
	Sort      int            `json:"sort"`
	DataScope int            `gorm:"default:1" json:"data_scope"`
	IsSystem  bool           `gorm:"default:false" json:"is_system"`
	Menus     []SysMenu      `gorm:"many2many:sys_role_menus;" json:"-"`
}

// SysMenu 菜单/权限
type SysMenu struct {
	ID         int64          `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	ParentID   int64          `gorm:"default:0;index" json:"parent_id"`
	Type       int            `gorm:"default:2" json:"type"` // 1目录 2菜单 3按钮
	Title      string         `gorm:"size:64" json:"title"`
	Name       string         `gorm:"size:64" json:"name"`
	Permission string         `gorm:"size:128" json:"permission"`
	Icon       string         `gorm:"type:text" json:"icon"`
	Path       string         `gorm:"size:128" json:"path"`
	Component  string         `gorm:"size:128" json:"component"`
	Redirect   string         `gorm:"size:128" json:"redirect"`
	Sort       int            `gorm:"column:sort_order" json:"sort"`
	Status     string         `gorm:"size:1;default:0" json:"status"`
	Hidden     bool           `json:"hidden"`
	KeepAlive  bool           `json:"keep_alive"`
	Affix      bool           `json:"affix"`
	AlwaysShow bool           `json:"always_show"`
	Breadcrumb bool           `gorm:"default:true" json:"breadcrumb"`
	ShowInTabs bool           `gorm:"default:true" json:"show_in_tabs"`
	ActiveMenu string         `gorm:"size:128" json:"active_menu"`
	IsSystem   bool           `json:"is_system"`
}

// SysDictType 字典类型
type SysDictType struct {
	ID        int64          `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"size:64" json:"name"`
	Code      string         `gorm:"size:64;uniqueIndex" json:"code"`
	Status    string         `gorm:"size:1;default:0" json:"status"`
	Sort      int            `gorm:"column:sort_order" json:"sort"`
	Remark    string         `gorm:"size:512" json:"remark"`
	IsSystem  bool           `json:"is_system"`
}

// SysDictData 字典数据
type SysDictData struct {
	ID        int64          `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	TypeID    int64          `gorm:"index" json:"type_id"`
	Label     string         `gorm:"size:64" json:"label"`
	Value     string         `gorm:"size:128" json:"value"`
	Status    string         `gorm:"size:1;default:0" json:"status"`
	Sort      int            `gorm:"column:sort_order" json:"sort"`
	Remark    string         `gorm:"size:512" json:"remark"`
}

const (
	StorageTypeLocal = 1
	StorageTypeOSS   = 2
)

// SysStorage 存储配置
type SysStorage struct {
	ID          int64          `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `gorm:"size:100" json:"name"`
	Code        string         `gorm:"size:64;uniqueIndex" json:"code"`
	Type        int            `gorm:"default:1" json:"type"`
	AccessKey   string         `gorm:"size:255" json:"access_key"`
	SecretKey   string         `gorm:"size:512" json:"-"`
	Endpoint    string         `gorm:"size:255" json:"endpoint"`
	BucketName  string         `gorm:"size:255" json:"bucket_name"`
	BaseURL     string         `gorm:"size:512" json:"base_url"`
	Domain      string         `gorm:"size:512" json:"domain"`
	Description string         `gorm:"size:512" json:"description"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	Sort        int            `json:"sort"`
	Status      string         `gorm:"size:1;default:0" json:"status"`
}

// SysFile 文件记录
type SysFile struct {
	ID           int64          `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	StorageID    int64          `gorm:"index" json:"storage_id"`
	Name         string         `gorm:"size:255" json:"name"`
	OriginalName string         `gorm:"size:255" json:"original_name"`
	Path         string         `gorm:"size:512;index" json:"path"`
	ParentPath   string         `gorm:"size:512;index" json:"parent_path"`
	URL          string         `gorm:"size:1024" json:"url"`
	Size         int64          `json:"size"`
	Extension    string         `gorm:"size:32" json:"extension"`
	ContentType  string         `gorm:"size:128" json:"content_type"`
	Type         int            `gorm:"default:1" json:"type"` // 0 文件夹 1 文件
}

// PhoneCallSession 真实电话外呼会话
type PhoneCallSession struct {
	ID          int64     `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	RobotID     int64     `gorm:"index" json:"robot_id"`
	Callee      string    `gorm:"size:32" json:"callee"`
	Status      string    `gorm:"size:32" json:"status"` // dialing/connected/ended/failed
	Provider    string    `gorm:"size:32" json:"provider"`
	ProviderRef string    `gorm:"size:128" json:"provider_ref"`
	ErrorMsg    string    `gorm:"size:512" json:"error_msg"`
	Transcript  string    `gorm:"type:text" json:"transcript"`
}
