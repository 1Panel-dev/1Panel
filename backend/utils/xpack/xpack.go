//go:build !xpack

package xpack

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
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
