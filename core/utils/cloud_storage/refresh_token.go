package cloud_storage

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/go-resty/resty/v2"
)

func loadParamFromVars(key string, vars map[string]interface{}) string {
	if _, ok := vars[key]; !ok {
		if key != "bucket" && key != "port" && key != "authMode" && key != "passPhrase" {
			global.LOG.Errorf("load param %s from vars failed, err: not exist!", key)
		}
		return ""
	}

	return fmt.Sprintf("%v", vars[key])
}

type aliTokenResp struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

func RefreshALIToken(varMap map[string]interface{}) (string, error) {
	refresh_token := loadParamFromVars("refresh_token", varMap)
	if len(refresh_token) == 0 {
		return "", errors.New("no such refresh token find in db")
	}
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	data := map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": refresh_token,
	}

	url := "https://api.aliyundrive.com/token/refresh"
	resp, err := client.R().
		SetBody(data).
		Post(url)

	if err != nil {
		return "", fmt.Errorf("load account token failed, err: %v", err)
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("load account token failed, code: %v", resp.StatusCode())
	}
	var respItem aliTokenResp
	if err := json.Unmarshal(resp.Body(), &respItem); err != nil {
		return "", err
	}
	return respItem.RefreshToken, nil
}

func RefreshToken(grantType string, tokenType string, varMap map[string]interface{}) (string, error) {
	data := url.Values{}
	isCN := loadParamFromVars("isCN", varMap)
	data.Set("client_id", loadParamFromVars("client_id", varMap))
	data.Set("client_secret", loadParamFromVars("client_secret", varMap))
	if grantType == "refresh_token" {
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", loadParamFromVars("refresh_token", varMap))
	} else {
		data.Set("grant_type", "authorization_code")
		data.Set("code", loadParamFromVars("code", varMap))
	}
	data.Set("redirect_uri", loadParamFromVars("redirect_uri", varMap))
	client := &http.Client{}
	defer client.CloseIdleConnections()
	url := "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	if isCN == "true" {
		url = "https://login.chinacloudapi.cn/common/oauth2/v2.0/token"
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("new http post client for access token failed, err: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request for access token failed, err: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read data from response body failed, err: %v", err)
	}

	tokenMap := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &tokenMap); err != nil {
		return "", fmt.Errorf("unmarshal data from response body failed, err: %v", err)
	}
	if tokenType == "accessToken" {
		accessToken, ok := tokenMap["access_token"].(string)
		if !ok {
			return "", errors.New("no such access token in response")
		}
		tokenMap = nil
		return accessToken, nil
	}
	refreshToken, ok := tokenMap["refresh_token"].(string)
	if !ok {
		return "", errors.New("no such access token in response")
	}
	tokenMap = nil
	return refreshToken, nil
}
