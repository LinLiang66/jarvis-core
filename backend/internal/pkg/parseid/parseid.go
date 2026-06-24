package parseid

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Path 解析路径或任意字符串为 int64
func Path(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("id 不能为空")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("无效 id: %s", s)
	}
	return n, nil
}

// PathKey 解析并返回规范化的十进制字符串（供 GORM、URL 使用，与 JSON 响应一致）
func PathKey(s string) (string, error) {
	n, err := Path(s)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(n, 10), nil
}

// GinKey 读取 gin 路径参数并规范化为 string 主键
func GinKey(c *gin.Context, name string) (string, error) {
	return PathKey(c.Param(name))
}

// NormalizeStrings 校验前端传入的 string id 列表（去重、去空白、须为正整数）
func NormalizeStrings(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, errors.New("ids 不能为空")
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		key, err := PathKey(raw)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	if len(out) == 0 {
		return nil, errors.New("ids 不能为空")
	}
	return out, nil
}

// ToInt64Slice 将前端 string id 列表转为 int64（允许空列表）
func ToInt64Slice(ids []string) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	keys, err := NormalizeStrings(ids)
	if err != nil {
		return nil, err
	}
	out := make([]int64, len(keys))
	for i, s := range keys {
		out[i], _ = Path(s)
	}
	return out, nil
}

// BindDeleteIDs 绑定 POST /delete 请求体中的 ids（前端为 string 数组）
func BindDeleteIDs(c *gin.Context) ([]string, error) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		return nil, errors.New("ids 不能为空")
	}
	return NormalizeStrings(body.IDs)
}
