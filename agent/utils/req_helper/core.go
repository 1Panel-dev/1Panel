package req_helper

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
)

func PostLocalCore(url string) error {
	settingRepo := repo.NewISettingRepo()
	port, err := settingRepo.GetValueByKey("ServerPort")
	if err != nil {
		return err
	}
	sslStatus, err := settingRepo.GetValueByKey("SSL")
	if err != nil {
		return err
	}
	var prefix string
	if sslStatus == "Disable" {
		prefix = "http://"
	} else {
		prefix = "https://"
	}
	reloadURL := fmt.Sprintf("%s://127.0.0.1:%s/api/v2%s", prefix, port, url)
	req, err := http.NewRequest("POST", reloadURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	client := &http.Client{}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
