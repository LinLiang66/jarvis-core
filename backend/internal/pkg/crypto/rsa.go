package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

// GenerateRSAKeyPair 生成 RSA 2048 密钥对，返回 PKCS#8 私钥 PEM 与 PKIX 公钥 PEM。
func GenerateRSAKeyPair() (privatePEM, publicPEM string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", "", err
	}
	privBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: privDER}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}
	return string(pem.EncodeToMemory(privBlock)), string(pem.EncodeToMemory(pubBlock)), nil
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid private key PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not RSA private key")
	}
	return priv, nil
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return key, nil
}

// EncryptByPrivateKey 与 Java RSAUtil.encryptByPrivateKey 兼容（RSA/ECB/PKCS1Padding + 私钥加密）。
func EncryptByPrivateKey(privatePEM, plain string) (string, error) {
	priv, err := parsePrivateKey(privatePEM)
	if err != nil {
		return "", err
	}
	data, err := rsa.SignPKCS1v15(rand.Reader, priv, 0, []byte(plain))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DecryptByPublicKey 服务端用应用公钥解密客户端私钥加密的数据。
func DecryptByPublicKey(publicPEM, cipherB64 string) (string, error) {
	pub, err := parsePublicKey(publicPEM)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}
	plain, err := publicDecryptPKCS1v15(pub, raw)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// EncryptByPublicKey 服务端用应用公钥加密（客户端用私钥解密）。
func EncryptByPublicKey(publicPEM, plain string) (string, error) {
	pub, err := parsePublicKey(publicPEM)
	if err != nil {
		return "", err
	}
	data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plain))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DecryptByPrivateKey 与 Java RSAUtil.decryptByPrivateKey 兼容。
func DecryptByPrivateKey(privatePEM, cipherB64 string) (string, error) {
	priv, err := parsePrivateKey(privatePEM)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, priv, raw)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// StripPEMHeaders 去除 PEM 头尾，返回 base64(DER) 便于客户端存储 appSecret。
func StripPEMHeaders(pemStr string) string {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return pemStr
	}
	return base64.StdEncoding.EncodeToString(block.Bytes)
}

// ToPEMPrivateKey 将 strip 后的 base64 还原为 PEM 私钥。
func ToPEMPrivateKey(stripped string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(stripped)
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: raw}
	return string(pem.EncodeToMemory(block)), nil
}

// publicDecryptPKCS1v15 公钥解密（私钥 PKCS1v15 "加密" 的逆运算）。
func publicDecryptPKCS1v15(pub *rsa.PublicKey, ciphertext []byte) ([]byte, error) {
	k := (pub.N.BitLen() + 7) / 8
	if len(ciphertext) != k {
		return nil, fmt.Errorf("cipher length %d != key size %d", len(ciphertext), k)
	}
	c := new(big.Int).SetBytes(ciphertext)
	m := new(big.Int).Exp(c, big.NewInt(int64(pub.E)), pub.N)
	em := leftPad(m.Bytes(), k)
	return unpadPKCS1v15(em)
}

func unpadPKCS1v15(em []byte) ([]byte, error) {
	if len(em) < 11 {
		return nil, errors.New("message too short")
	}
	if em[0] != 0x00 || em[1] != 0x01 {
		return nil, errors.New("invalid padding header")
	}
	i := 2
	for ; i < len(em); i++ {
		if em[i] == 0x00 {
			break
		}
		if em[i] != 0xff {
			return nil, errors.New("invalid padding")
		}
	}
	if i >= len(em)-1 {
		return nil, errors.New("invalid padding structure")
	}
	return em[i+1:], nil
}

func leftPad(b []byte, size int) []byte {
	if len(b) >= size {
		return b[len(b)-size:]
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}
