package model

// OpenApp 开放平台应用凭证。
type OpenApp struct {
	Base
	AppID       string `gorm:"size:64;uniqueIndex" json:"app_id"`
	AppName     string `gorm:"size:128" json:"app_name"`
	SignSecret  string `gorm:"size:128" json:"-"`                    // yd_md5 签名密钥
	RSAPublicKey  string `gorm:"type:text" json:"rsa_public_key"`    // 应用 RSA 公钥 PEM
	RSAPrivateKey string `gorm:"type:text" json:"-"`                 // 仅创建时返回给管理员，之后可选清空
	Status     string `gorm:"size:1;default:0" json:"status"` // 0 正常 1 禁用
	TotalQuota int    `gorm:"default:0" json:"total_quota"`   // 可用配额余额，0 表示无可用次数
	Remark     string `gorm:"size:512" json:"remark,omitempty"`
	TotalCalls int64  `gorm:"default:0" json:"total_calls"` // 累计调用次数
}

// OpenAPICallLog 开放平台 API 调用明细。
type OpenAPICallLog struct {
	Base
	AppID    string `gorm:"size:64;index" json:"app_id"`
	Action   string `gorm:"size:128;index" json:"action"`
	Success  bool   `gorm:"index" json:"success"`
	Duration int64  `json:"duration_ms"` // 毫秒
	Message  string `gorm:"size:512" json:"message,omitempty"`
	ClientIP string `gorm:"size:64" json:"client_ip,omitempty"`
}

// OpenAPIDailyStat 按应用+动作+日期的调用统计。
type OpenAPIDailyStat struct {
	Base
	AppID        string `gorm:"size:64;uniqueIndex:idx_open_stat" json:"app_id"`
	Action       string `gorm:"size:128;uniqueIndex:idx_open_stat" json:"action"`
	StatDate     string `gorm:"size:10;uniqueIndex:idx_open_stat" json:"stat_date"` // YYYY-MM-DD
	TotalCount   int64  `gorm:"default:0" json:"total_count"`
	SuccessCount int64  `gorm:"default:0" json:"success_count"`
	FailCount    int64  `gorm:"default:0" json:"fail_count"`
}

// OpenAPIHourlySyncLog 小时桶同步幂等记录（防止集群重复入库）。
type OpenAPIHourlySyncLog struct {
	Base
	HourKey string `gorm:"size:10;uniqueIndex" json:"hour_key"` // YYYYMMDDHH
}
