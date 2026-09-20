package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

func APIKeyEncryptionKey() (string, error) {
	if global.CONF.Base.EncryptKey != "" {
		return global.CONF.Base.EncryptKey, nil
	}
	var value model.Setting
	if global.DB == nil {
		return "", errors.New("API key encryption key unavailable")
	}
	if err := global.DB.Where("key = ?", "EncryptKey").First(&value).Error; err != nil {
		return "", err
	}
	if value.Value == "" {
		return "", errors.New("API key encryption key unavailable")
	}
	return value.Value, nil
}

func apiKeyAEAD(key string) (cipher.AEAD, error) {
	if key == "" {
		return nil, errors.New("API key encryption key unavailable")
	}
	derived := sha256.Sum256([]byte("1panel:api-key:v1:" + key))
	block, err := aes.NewCipher(derived[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func EncryptAPIKey(secret, id, key string) (string, error) {
	aead, err := apiKeyAEAD(key)
	if err != nil {
		return "", err
	}
	if secret == "" || id == "" {
		return "", errors.New("empty API key")
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(secret), []byte("api-key:v1:"+id))), nil
}

func DecryptAPIKey(value, id, key string) (string, error) {
	aead, err := apiKeyAEAD(key)
	if err != nil {
		return "", err
	}
	data, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil || len(data) < aead.NonceSize()+aead.Overhead() {
		return "", errors.New("invalid API key ciphertext")
	}
	plain, err := aead.Open(nil, data[:aead.NonceSize()], data[aead.NonceSize():], []byte("api-key:v1:"+id))
	if err != nil {
		return "", errors.New("invalid API key ciphertext")
	}
	return string(plain), nil
}
