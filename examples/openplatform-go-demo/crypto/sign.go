package crypto

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"sort"
	"strings"
)

const SignMethodA2MD5 = "a2_md5"

// BuildA2MD5Sign sorts keys, concatenates key+value, appends signSecret, MD5 hex lowercase.
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

// RandomDigits returns a random numeric string of length n.
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
