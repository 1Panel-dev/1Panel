package req_helper

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/xpack"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
)

func HandleGet(url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 300,
	}
	client.Transport = loadRequestTransport()
	defer client.CloseIdleConnections()

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
		_ = resp.Body.Close()
		return nil, buserr.New("ErrHttpReqNotFound")
	}

	return resp, nil
}

func HandleRequest(url, method string, timeout int) (int, []byte, error) {
	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	client := http.Client{Timeout: time.Duration(timeout) * time.Second, Transport: transport}
	return HandleRequestWithClient(&client, url, method, timeout)
}

func HandleRequestWithClient(client *http.Client, url, method string, timeout int) (int, []byte, error) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("handle request failed, error message: %v", r)
			return
		}
	}()

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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}

type RequestResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

func HandleRequestWithHeaders(client *http.Client, url, method string, timeout int, reqHeaders map[string]string) (*RequestResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("handle request failed, error message: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range reqHeaders {
		request.Header.Set(k, v)
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &RequestResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Header:     resp.Header,
	}, nil
}

func RequestFile(url, method string, timeout int) (io.ReadCloser, context.CancelFunc, error) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("handle request failed, error message: %v", r)
			return
		}
	}()
	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	client := http.Client{Timeout: time.Duration(timeout) * time.Second, Transport: transport}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, cancel, err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return nil, cancel, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, cancel, errors.New(resp.Status)
	}
	return resp.Body, cancel, nil
}

func loadRequestTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
}

var ErrProxyRequired = errors.New("runtime environment requires a configured proxy")

// LoadRuntimeProxyURL reads and validates local configuration without creating an HTTP client.
// A nil URL preserves the default network behavior in the global environment.
func LoadRuntimeProxyURL() (*url.URL, error) {
	var rows []model.Setting
	if err := global.DB.Where("key IN ?", []string{"Edition", "ProxyType", "ProxyUrl", "ProxyPort", "ProxyUser", "ProxyPasswd"}).Find(&rows).Error; err != nil {
		return nil, err
	}
	settings := make(map[string]string, len(rows))
	for _, row := range rows {
		settings[row.Key] = row.Value
	}
	if settings["Edition"] == "intl" {
		return nil, nil
	}
	kind, host, port := settings["ProxyType"], strings.TrimSpace(settings["ProxyUrl"]), settings["ProxyPort"]
	if kind == "" || kind == "close" || host == "" || port == "" {
		return nil, ErrProxyRequired
	}
	if kind != "http" && kind != "https" && kind != "socks5" {
		return nil, ErrProxyRequired
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return nil, ErrProxyRequired
	}
	if !strings.Contains(host, "://") {
		host = kind + "://" + host
	}
	address, err := url.Parse(host)
	if err != nil || address.Scheme != kind || address.Hostname() == "" || address.User != nil || address.RawQuery != "" || address.Fragment != "" || (address.Path != "" && address.Path != "/") {
		return nil, ErrProxyRequired
	}
	address.Host = net.JoinHostPort(address.Hostname(), port)
	address.Path = ""
	if settings["ProxyUser"] != "" {
		password, err := encrypt.StringDecrypt(settings["ProxyPasswd"])
		if err != nil {
			return nil, ErrProxyRequired
		}
		address.User = url.UserPassword(settings["ProxyUser"], password)
	}
	return address, nil
}

// NewRuntimeHTTPClient builds a client for requests using the validated runtime proxy configuration.
func NewRuntimeHTTPClient() (*http.Client, error) {
	address, err := LoadRuntimeProxyURL()
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	if address == nil {
		return client, nil
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Explicit proxy selection prevents NO_PROXY from bypassing the panel proxy.
	// net/http also supports SOCKS5 URLs and proxy authentication.
	transport.Proxy = http.ProxyURL(address)
	client.Transport = transport
	return client, nil
}
