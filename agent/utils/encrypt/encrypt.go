package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

func StringEncryptWithBase64(text string) (string, error) {
	accessKeyItem, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	encryptKeyItem, err := StringEncrypt(string(accessKeyItem))
	if err != nil {
		return "", err
	}
	return encryptKeyItem, nil
}

func StringEncryptWithKey(text, key string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	pass := []byte(text)
	xpass, err := aesEncryptWithSalt([]byte(key), pass)
	if err == nil {
		pass64 := base64.StdEncoding.EncodeToString(xpass)
		return pass64, err
	}
	return "", err
}

func StringEncrypt(text string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	if len(global.CONF.Base.EncryptKey) == 0 {
		var encryptSetting model.Setting
		if err := global.DB.Where("key = ?", "EncryptKey").First(&encryptSetting).Error; err != nil {
			return "", err
		}
		global.CONF.Base.EncryptKey = encryptSetting.Value
	}
	key := global.CONF.Base.EncryptKey
	return StringEncryptWithKey(text, key)
}

func StringDecryptWithBase64(text string) (string, error) {
	decryptItem, err := StringDecrypt(text)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(decryptItem)), nil
}

func StringDecryptWithKey(text, key string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	bytesPass, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	var tpass []byte
	tpass, err = aesDecryptWithSalt([]byte(key), bytesPass)
	if err == nil {
		result := string(tpass[:])
		return result, err
	}
	return "", err
}

func StringDecrypt(text string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	if len(global.CONF.Base.EncryptKey) == 0 {
		var encryptSetting model.Setting
		if err := global.DB.Where("key = ?", "EncryptKey").First(&encryptSetting).Error; err != nil {
			return "", err
		}
		global.CONF.Base.EncryptKey = encryptSetting.Value
	}
	key := global.CONF.Base.EncryptKey
	return StringDecryptWithKey(text, key)
}

func padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

func unPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func aesEncryptWithSalt(key, plaintext []byte) ([]byte, error) {
	plaintext = padding(plaintext, aes.BlockSize)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[0:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}
func aesDecryptWithSalt(key, ciphertext []byte) ([]byte, error) {
	var block cipher.Block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("iciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)
	ciphertext = unPadding(ciphertext)
	return ciphertext, nil
}
