package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        int64          `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	Base
	Username    string `gorm:"size:64;uniqueIndex" json:"username"`
	Password    string `gorm:"size:128" json:"-"`
	Nickname    string `gorm:"size:64" json:"nickname"`
	Phone       string `gorm:"size:32" json:"phone,omitempty"`
	Email       string `gorm:"size:128" json:"email,omitempty"`
	Status      string `gorm:"size:1;default:0" json:"status"`
	Roles       string `gorm:"size:256" json:"-"`
	Permissions string `gorm:"type:text" json:"-"`
}

func (u User) RoleList() []string {
	if u.Roles == "" {
		return []string{"admin"}
	}
	return splitComma(u.Roles)
}

func (u User) PermList() []string {
	if u.Permissions == "" {
		return defaultPerms()
	}
	return splitComma(u.Permissions)
}

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func defaultPerms() []string {
	return []string{
		"module_system:user:query", "module_system:user:edit",
		"module_system:role:query", "module_system:role:edit",
		"module_system:menu:query", "module_system:menu:edit",
		"module_system:dict:query", "module_system:dict:edit",
		"openapp:query", "openapp:edit",
		"openstat:query", "openaction:query", "openaction:edit",
	}
}
