package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// LoadPrivateKeyFromDerBase64 loads AppSecret (PKCS#8 DER Base64) as RSA private key PEM.
func LoadPrivateKeyFromDerBase64(appSecretBase64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(appSecretBase64))
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: raw}
	return string(pem.EncodeToMemory(block)), nil
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

// EncryptByPrivateKey is compatible with Java RsaUtil.encryptByPrivateKey.
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

// DecryptByPrivateKey decrypts serverPart encrypted with the server public key.
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
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}
