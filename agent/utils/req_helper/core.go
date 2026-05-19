package req_helper

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

const LocalTokenHeader = "X-Panel-Local-Token"

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

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	defer client.CloseIdleConnections()
	return postLocalCoreTo(fmt.Sprintf("%s://127.0.0.1:%s", prefix, serverPortSetting.Value), url, client)
}

func postLocalCoreTo(baseURL string, url string, client *http.Client) error {
	token, err := ensureLocalToken()
	if err != nil {
		return err
	}
	if client == nil {
		client = http.DefaultClient
	}

	reloadURL := strings.TrimRight(baseURL, "/") + "/api/v2" + url
	req, err := http.NewRequest(http.MethodPost, reloadURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	req.Header.Set(LocalTokenHeader, token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request local core failed: status %s", resp.Status)
	}

	var result dto.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Code != http.StatusOK {
		return fmt.Errorf("request local core failed: code %d, message: %s", result.Code, result.Message)
	}
	return nil
}

func ensureLocalToken() (string, error) {
	secret, err := ensureLocalSecret()
	if err != nil {
		return "", err
	}
	return generateLocalToken(secret), nil
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
	tmpDir := global.Dir.TmpDir
	if tmpDir == "" {
		tmpDir = filepath.Join(global.CONF.Base.InstallDir, "1panel/tmp")
	}
	return filepath.Join(tmpDir, ".secret")
}

func generateLocalToken(secret string) string {
	today := time.Now().Format("2006-01-02")
	h := md5.New()
	h.Write([]byte(secret + "-" + today))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
