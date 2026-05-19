package common

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
)

const LocalTokenHeader = "X-Panel-Local-Token"

func EnsureLocalToken() (string, error) {
	secret, err := ensureLocalSecret()
	if err != nil {
		return "", err
	}
	return generateLocalToken(secret), nil
}

func ValidateLocalToken(token string) bool {
	if token == "" {
		return false
	}
	expected, err := EnsureLocalToken()
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}

func ensureLocalSecret() (string, error) {
	secretPath := localSecretPath()
	data, err := os.ReadFile(secretPath)
	if err == nil && len(data) != 0 {
		return string(data), nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(secretPath), 0700); err != nil {
		return "", err
	}
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}
	secret := hex.EncodeToString(secretBytes)
	if err := os.WriteFile(secretPath, []byte(secret), 0600); err != nil {
		return "", err
	}
	return secret, nil
}

func localSecretPath() string {
	return filepath.Join(global.CONF.Base.InstallDir, "1panel/tmp/.secret")
}

func generateLocalToken(secret string) string {
	today := time.Now().Format("2006-01-02")
	h := md5.New()
	h.Write([]byte(secret + "-" + today))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
