package crypto

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"sort"
	"strings"
)

const SignMethodA2MD5 = "a2_md5"

// BuildA2MD5Sign MD5 签名：参数按 key 升序，拼接 key+value，末尾追加 signSecret，MD5 小写 hex。
func BuildA2MD5Sign(params map[string]string, signSecret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	sb.WriteString(signSecret)
	hash := md5.Sum([]byte(sb.String()))
	return fmt.Sprintf("%x", hash)
}

// VerifyA2MD5Sign 验签（大小写不敏感）。
func VerifyA2MD5Sign(params map[string]string, signSecret, sign string) bool {
	if sign == "" {
		return false
	}
	expected := BuildA2MD5Sign(params, signSecret)
	return strings.EqualFold(expected, sign)
}

// RandomDigits 生成指定位数随机数字字符串。
func RandomDigits(n int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, n)
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = digits[int(raw[i])%10]
	}
	return string(b), nil
}
