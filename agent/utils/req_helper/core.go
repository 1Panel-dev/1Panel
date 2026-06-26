package req_helper

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func PostLocalCore(url string) error {
	var serverPortSetting model.Setting
	_ = global.CoreDB.Model(&model.Setting{}).Where("key = ?", "ServerPort").First(&serverPortSetting).Error
	var sslStatusSetting model.Setting
	_ = global.CoreDB.Model(&model.Setting{}).Where("key = ?", "SSL").First(&sslStatusSetting).Error

	var prefix string
	if sslStatusSetting.Value == "Disable" {
		prefix = "http"
	} else {
		prefix = "https"
	}

	reloadURL := fmt.Sprintf("%s://127.0.0.1:%s/api/v2%s", prefix, serverPortSetting.Value, url)
	req, err := http.NewRequest("POST", reloadURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	token, err := loadDailyLocalToken()
	if err != nil {
		return err
	}
	req.Header.Set("X-Panel-Local-Token", token)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("core request failed: status=%s, body=%s", resp.Status, strings.TrimSpace(string(bodyBytes)))
	}
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if len(bytes.TrimSpace(bodyBytes)) > 0 {
		if err := json.Unmarshal(bodyBytes, &result); err == nil && result.Code != 0 && result.Code != http.StatusOK {
			return fmt.Errorf("core request failed: code=%d, message=%s", result.Code, result.Message)
		}
	}
	return nil
}

func loadDailyLocalToken() (string, error) {
	secret, err := loadOrGenerateLocalSecret()
	if err != nil {
		return "", err
	}
	today := time.Now().Format("2006-01-02")
	h := md5.New()
	h.Write([]byte(secret + "-" + today))
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

func loadOrGenerateLocalSecret() (string, error) {
	secretFilePath := path.Join(global.Dir.TmpDir, ".secret")
	data, err := os.ReadFile(secretFilePath)
	if err == nil {
		secret := strings.TrimSpace(string(data))
		if secret != "" {
			return secret, nil
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	dir := filepath.Dir(secretFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	secret := hex.EncodeToString(key)
	if err := os.WriteFile(secretFilePath, []byte(secret), 0600); err != nil {
		return "", err
	}
	return secret, nil
}
