//go:build !xpack

package xpack

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
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

func StartClam(startClam model.Clam, isUpdate bool) (int, error) {
	return 0, buserr.New(constant.ErrXpackNotFound)
}

func CreateAlert(createAlert dto.CreateOrUpdateAlert) error {
	return nil
}

func UpdateAlert(updateAlert dto.CreateOrUpdateAlert) error {
	return nil
}

func DeleteAlert(alertBase dto.AlertBase) error {
	return nil
}

func GetAlert(alertBase dto.AlertBase) uint {
	return 0
}

func PushAlert(pushAlert dto.PushAlert) error {
	return nil
}
