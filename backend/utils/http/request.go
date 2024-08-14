package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/xpack"
)

func HandleGet(url, method string, timeout int) (int, []byte, error) {
	transport := xpack.LoadRequestTransport()
	return HandleGetWithTransport(url, method, transport, timeout)
}

func HandleGetWithTransport(url, method string, transport *http.Transport, timeout int) (int, []byte, error) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("handle request failed, error message: %v", r)
			return
		}
	}()

	client := http.Client{Timeout: time.Duration(timeout) * time.Second, Transport: transport}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, body, nil
}
