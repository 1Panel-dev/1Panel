package http

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/xpack"
)

func HandleGet(url, method string) (int, []byte, error) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf(" A panic occurred during handle request, error message: %v", r)
			return
		}
	}()

	client := http.Client{Timeout: 10 * time.Second}
	ok, transport := xpack.LoadRequestTransport()
	if ok {
		client.Transport = transport
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, body, nil
}
