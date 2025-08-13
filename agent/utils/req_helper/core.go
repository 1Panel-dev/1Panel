package req_helper

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"net/http"
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
	return nil
}
