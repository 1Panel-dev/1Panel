package http

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"net/http"
	"strings"
	"time"
)

func GetHttpRes(url string) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: 10 * time.Second,
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, buserr.WithMap("ErrCreateHttpClient", map[string]interface{}{"err": err.Error()}, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if err == context.DeadlineExceeded {
			return nil, buserr.WithMap("ErrHttpReqTimeOut", map[string]interface{}{"err": err.Error()}, err)
		} else {
			if strings.Contains(err.Error(), "no such host") {
				return nil, buserr.New("ErrNoSuchHost")
			}
			return nil, buserr.WithMap("ErrHttpReqFailed", map[string]interface{}{"err": err.Error()}, err)
		}
	}
	if resp.StatusCode == 404 {
		return nil, buserr.New("ErrHttpReqNotFound")
	}

	return resp, nil
}
