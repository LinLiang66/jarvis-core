package storage

import (
	"net/url"
	"strings"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/model"
)

// ResolveFileURL 按存储配置补全文件访问地址。
func ResolveFileURL(cfg *config.Config, st *model.SysStorage, f *model.SysFile) string {
	if st == nil || f == nil {
		return ""
	}
	raw := strings.TrimSpace(f.URL)
	if raw == "" {
		raw = strings.TrimLeft(strings.TrimSpace(f.Path), "/")
	}
	if isHTTPURL(raw) {
		return BuildPublicURL(st, raw)
	}
	objectKey := strings.TrimLeft(strings.TrimSpace(f.Path), "/")
	if objectKey == "" {
		objectKey = objectPathFromURL(raw, st)
	}
	engine := NewEngine(cfg, st)
	if st.Type == model.StorageTypeLocal {
		return BuildPublicURL(st, JoinURL(engine.localDomain(), objectKey))
	}
	if strings.TrimSpace(st.Domain) != "" {
		return BuildPublicURL(st, JoinURL(st.Domain, objectKey))
	}
	return BuildPublicURL(st, engine.buildOSSRawURL(st.BucketName, objectKey))
}

func isHTTPURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// NormalizeFileURL 上传完成后规范化写入数据库的 URL。
func NormalizeFileURL(st *model.SysStorage, publicURL, objectKey string) string {
	publicURL = strings.TrimSpace(publicURL)
	if publicURL != "" {
		return BuildPublicURL(st, publicURL)
	}
	objectKey = strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	if strings.TrimSpace(st.Domain) != "" {
		u, err := url.Parse(strings.TrimRight(st.Domain, "/") + "/" + objectKey)
		if err == nil {
			return u.String()
		}
		return JoinURL(st.Domain, objectKey)
	}
	return objectKey
}
