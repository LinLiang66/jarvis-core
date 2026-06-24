package crypto

import (
	"bytes"
	"crypto/des"
	"encoding/base64"
	"errors"
	"sync"
)

// TDESCipher 可复用的 3DES/ECB/PKCS5Padding 加解密器，同一密钥仅初始化一次 block。
type TDESCipher struct {
	block    cipherBlock
	blockSize int
}

type cipherBlock interface {
	BlockSize() int
	Encrypt(dst, src []byte)
	Decrypt(dst, src []byte)
}

var (
	cipherPool sync.Map // string(key) -> *TDESCipher，供跨包可选的全局 dedup（主要缓存在 SessionStore）
)

// NewTDESCipher 根据会话密钥创建加解密器（密钥归一化 + block 初始化仅执行一次）。
func NewTDESCipher(key []byte) (*TDESCipher, error) {
	k := normalize3DESKey(key)
	cacheKey := string(k)
	if v, ok := cipherPool.Load(cacheKey); ok {
		return v.(*TDESCipher), nil
	}
	block, err := des.NewTripleDESCipher(k)
	if err != nil {
		return nil, err
	}
	c := &TDESCipher{block: block, blockSize: block.BlockSize()}
	actual, _ := cipherPool.LoadOrStore(cacheKey, c)
	return actual.(*TDESCipher), nil
}

// Encrypt 加密并返回 base64 密文。
func (c *TDESCipher) Encrypt(plain []byte) (string, error) {
	padded := pkcs5Pad(plain, c.blockSize)
	out := make([]byte, len(padded))
	for i := 0; i < len(padded); i += c.blockSize {
		c.block.Encrypt(out[i:i+c.blockSize], padded[i:i+c.blockSize])
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt 解密 base64 密文。
func (c *TDESCipher) Decrypt(cipherB64 string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return nil, err
	}
	if len(raw)%c.blockSize != 0 {
		return nil, errors.New("invalid cipher block size")
	}
	out := make([]byte, len(raw))
	for i := 0; i < len(raw); i += c.blockSize {
		c.block.Decrypt(out[i:i+c.blockSize], raw[i:i+c.blockSize])
	}
	return pkcs5Unpad(out)
}

// TripleDESEncrypt 一次性加密（兼容旧调用；推荐使用 TDESCipher）。
func TripleDESEncrypt(key, plain []byte) (string, error) {
	c, err := NewTDESCipher(key)
	if err != nil {
		return "", err
	}
	return c.Encrypt(plain)
}

// TripleDESDecrypt 一次性解密（兼容旧调用；推荐使用 TDESCipher）。
func TripleDESDecrypt(key []byte, cipherB64 string) ([]byte, error) {
	c, err := NewTDESCipher(key)
	if err != nil {
		return nil, err
	}
	return c.Decrypt(cipherB64)
}

func normalize3DESKey(key []byte) []byte {
	switch len(key) {
	case 16:
		out := make([]byte, 24)
		copy(out, key)
		copy(out[16:], key[:8])
		return out
	case 24:
		return key
	default:
		if len(key) > 24 {
			return key[:24]
		}
		out := make([]byte, 24)
		copy(out, key)
		return out
	}
}

func pkcs5Pad(data []byte, blockSize int) []byte {
	n := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(n)}, n)
	return append(data, padding...)
}

func pkcs5Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	n := int(data[len(data)-1])
	if n <= 0 || n > len(data) {
		return nil, errors.New("invalid padding")
	}
	for i := 0; i < n; i++ {
		if data[len(data)-1-i] != byte(n) {
			return nil, errors.New("invalid padding bytes")
		}
	}
	return data[:len(data)-n], nil
}
