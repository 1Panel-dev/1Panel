//go:build !xpack

package xpack

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context, currentNode string) {}

func UpdateGroup(name string, group, newGroup uint) error { return nil }

func CheckBackupUsed(name string) error { return nil }

func RequestToAllAgent(reqUrl, reqMethod string, reqBody io.Reader) error { return nil }

func LoadRequestTransport() *http.Transport {
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
