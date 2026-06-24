package crypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"errors"
)

// TDESCipher 3DES/ECB/PKCS5Padding encryptor/decryptor.
type TDESCipher struct {
	block     cipher.Block
	blockSize int
}

func NewTDESCipher(key []byte) (*TDESCipher, error) {
	k := normalize3DESKey(key)
	block, err := des.NewTripleDESCipher(k)
	if err != nil {
		return nil, err
	}
	return &TDESCipher{block: block, blockSize: block.BlockSize()}, nil
}

func (c *TDESCipher) Encrypt(plain []byte) (string, error) {
	padded := pkcs5Pad(plain, c.blockSize)
	out := make([]byte, len(padded))
	for i := 0; i < len(padded); i += c.blockSize {
		c.block.Encrypt(out[i:i+c.blockSize], padded[i:i+c.blockSize])
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

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
