package proxy_local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/i18n"
)

func NewLocalClient(reqUrl, reqMethod string, body io.Reader) (interface{}, error) {
	sockPath := "/etc/1panel/agent.sock"
	if _, err := os.Stat(sockPath); err != nil {
		return nil, fmt.Errorf("no such agent.sock find in localhost, err: %v", err)
	}
	dialUnix := func() (conn net.Conn, err error) {
		return net.Dial("unix", sockPath)
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialUnix()
		},
	}
	client := &http.Client{
		Transport: transport,
	}
	defer client.CloseIdleConnections()
	parsedURL, err := url.Parse("http://unix")
	if err != nil {
		return nil, fmt.Errorf("handle url Parse failed, err: %v \n", err)
	}
	rURL := &url.URL{
		Scheme: "http",
		Path:   reqUrl,
		Host:   parsedURL.Host,
	}

	req, err := http.NewRequest(reqMethod, rURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("creating request failed, err: %v", err)
	}

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
		return nil, fmt.Errorf("json umarshal resp data failed, err: %v", err)
	}
	if respJson.Code != http.StatusOK {
		return nil, errors.New(strings.ReplaceAll(respJson.Message, i18n.Get("ErrInternalServerKey"), ""))
	}

	return respJson.Data, nil
}
