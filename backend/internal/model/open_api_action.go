package model

// OpenAPIAction 开放平台接口定义（代码注册自动同步 + 管理端可编辑文档）。
type OpenAPIAction struct {
	Base
	Action         string `gorm:"size:128;uniqueIndex" json:"action"`
	Title          string `gorm:"size:128" json:"title"`
	Category       string `gorm:"size:64;index" json:"category"`
	Description    string `gorm:"size:512" json:"description"`
	Encrypted      bool   `gorm:"default:false" json:"encrypted"`
	Billable       bool   `gorm:"default:true" json:"billable"`
	Status         string `gorm:"size:1;default:0" json:"status"` // 0 启用 1 禁用
	RequestSchema  string `gorm:"type:text" json:"request_schema"`
	ResponseSchema string `gorm:"type:text" json:"response_schema"`
	RequestFields  string `gorm:"type:text" json:"request_fields"`
	ResponseFields string `gorm:"type:text" json:"response_fields"`
	DocMarkdown    string `gorm:"type:longtext" json:"doc_markdown"`
	Sort           int    `gorm:"default:0" json:"sort"`
	Source         string `gorm:"size:16;default:code" json:"source"` // code | manual
}
