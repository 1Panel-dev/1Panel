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
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
)

type AgentResponseError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *AgentResponseError) Error() string { return e.Message }

type HTTPStatusError struct {
	StatusCode int
	Message    string
}

func (e *HTTPStatusError) Error() string { return e.Message }

func NewLocalClient(reqUrl, reqMethod string, body io.Reader, ctx *gin.Context) (interface{}, error) {
	client := NewReusableClient()
	defer client.CloseIdleConnections()
	return client.Request(reqUrl, reqMethod, body, ctx)
}

func NewLocalClientWithContext(requestContext context.Context, reqURL, reqMethod string, body io.Reader, ctx *gin.Context, timeout time.Duration) (interface{}, error) {
	client := newReusableClientWithTimeout("/etc/1panel/agent.sock", timeout)
	defer client.CloseIdleConnections()
	return client.RequestWithContext(requestContext, reqURL, reqMethod, body, ctx)
}

type ReusableClient struct {
	client   *http.Client
	sockPath string
}

func NewReusableClient() *ReusableClient {
	return newReusableClient("/etc/1panel/agent.sock")
}

func newReusableClient(sockPath string) *ReusableClient {
	return newReusableClientWithTimeout(sockPath, 0)
}

func newReusableClientWithTimeout(sockPath string, timeout time.Duration) *ReusableClient {
	dialer := &net.Dialer{Timeout: timeout}
	transport := &http.Transport{
		MaxIdleConns:        12,
		MaxIdleConnsPerHost: 6,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", sockPath)
		},
	}
	return &ReusableClient{client: &http.Client{Transport: transport, Timeout: timeout}, sockPath: sockPath}
}

func (c *ReusableClient) CloseIdleConnections() {
	if c == nil || c.client == nil {
		return
	}
	c.client.CloseIdleConnections()
}

func (c *ReusableClient) Request(reqUrl, reqMethod string, body io.Reader, ctx *gin.Context) (interface{}, error) {
	return c.RequestWithContext(context.Background(), reqUrl, reqMethod, body, ctx)
}

func (c *ReusableClient) OpenWithContext(requestContext context.Context, reqUrl, reqMethod string, body io.Reader, ctx *gin.Context) (*http.Response, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("local agent client is not initialized")
	}
	if _, err := os.Stat(c.sockPath); err != nil {
		return nil, fmt.Errorf("no such agent.sock find in localhost, err: %v", err)
	}
	parsedURL, err := url.Parse("http://unix")
	if err != nil {
		return nil, fmt.Errorf("handle url Parse failed, err: %v \n", err)
	}
	rURL := &url.URL{
		Scheme: "http",
		Path:   reqUrl,
		Host:   parsedURL.Host,
	}

	req, err := http.NewRequestWithContext(requestContext, reqMethod, rURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("creating request failed, err: %v", err)
	}
	if ctx != nil {
		for key, values := range ctx.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client do request failed, err: %v", err)
	}
	return resp, nil
}

func (c *ReusableClient) RequestWithContext(requestContext context.Context, reqUrl, reqMethod string, body io.Reader, ctx *gin.Context) (interface{}, error) {
	resp, err := c.OpenWithContext(requestContext, reqUrl, reqMethod, body, ctx)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body from request failed, err: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		failure := &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("do request failed, err: %v", resp.Status),
		}
		var respJSON dto.Response
		if err := json.Unmarshal(bodyByte, &respJSON); err == nil && respJSON.Message != "" {
			failure.Message = fmt.Sprintf("do request failed, status=%v, message=%s", resp.Status, respJSON.Message)
		} else if msg := strings.TrimSpace(string(bodyByte)); msg != "" {
			failure.Message = fmt.Sprintf("do request failed, status=%v, body=%s", resp.Status, msg)
		}
		return nil, failure
	}

	var respJson dto.Response
	if err := json.Unmarshal(bodyByte, &respJson); err != nil {
		return nil, fmt.Errorf("json umarshal resp data failed, err: %v", err)
	}
	if respJson.Code != http.StatusOK {
		message := respJson.Message
		if global.I18n != nil {
			message = strings.ReplaceAll(message, i18n.Get("ErrInternalServerKey"), "")
		}
		return nil, &AgentResponseError{Code: respJson.Code, Message: message, Data: respJson.Data}
	}

	return respJson.Data, nil
}
