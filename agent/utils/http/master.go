package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

func RequestToMaster(reqUrl, reqMethod string, reqBody io.Reader) (interface{}, error) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	parsedURL, err := url.Parse(global.CONF.System.MasterRequestAddr)
	if err != nil {
		return nil, fmt.Errorf("handle url Parse failed, err: %v \n", err)
	}
	rURL := &url.URL{
		Path: reqUrl,
		Host: parsedURL.Host,
	}
	req, err := http.NewRequest(reqMethod, rURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("handle request failed, err: %v \n", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(constant.JWTHeaderName, global.CONF.System.MasterRequestToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client do request failed, err: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("do request failed, err: %v", resp.Status)
	}
	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body from request failed, err: %v", err)
	}
	var respJson dto.Response
	if err := json.Unmarshal(bodyByte, &respJson); err != nil {
		return nil, fmt.Errorf("json unmarshal resp data failed, err: %v", err)
	}
	if respJson.Code != http.StatusOK {
		return nil, fmt.Errorf("do request success but handle failed, err: %v", respJson.Message)
	}
	return respJson.Data, nil
}
