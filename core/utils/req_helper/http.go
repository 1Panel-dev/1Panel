package req_helper

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
)

func HandleRequest(url, method string, timeout int) (int, []byte, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
	return handleRequestWithTransport(url, method, transport, timeout)
}

func HandleRequestWithProxy(url, method string, timeout int) (int, []byte, error) {
	transport := xpack.LoadRequestTransport()
	return handleRequestWithTransport(url, method, transport, timeout)
}

func handleRequestWithTransport(url, method string, transport *http.Transport, timeout int) (int, []byte, error) {
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, body, nil
}

func HandleGet(url string) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
	return handleGetWithTransport(url, transport)
}

func HandleGetWithProxy(url string) (*http.Response, error) {
	transport := xpack.LoadRequestTransport()
	return handleGetWithTransport(url, transport)
}

func handleGetWithTransport(url string, transport *http.Transport) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 300,
	}
	client.Transport = transport

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, buserr.WithMap("ErrCreateHttpClient", map[string]interface{}{"err": err.Error()}, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
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
