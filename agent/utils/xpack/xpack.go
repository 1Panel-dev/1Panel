//go:build !xpack

package xpack

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"gorm.io/gorm"
)

func RemoveTamper(website string) {}

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

func LoadGpuInfo() []interface{} {
	return nil
}

func StartClam(startClam model.Clam, isUpdate bool) (int, error) {
	return 0, buserr.New(constant.ErrXpackNotFound)
}

func InitNodeData(tx *gorm.DB) (bool, string, error) { return true, "127.0.0.1", nil }

func RequestToMaster(reqUrl, reqMethod string, reqBody io.Reader) (interface{}, error) {
	return nil, nil
}

func GetImagePrefix() string {
	return ""
}
