package service

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/dockerhub"
)

type ImageRemoteCheckStatus struct {
	ProxyRequired bool `json:"proxyRequired"`
}

// GetImageRemoteCheckStatus validates local settings without contacting Docker Hub.
// Both image searches and update checks use the same proxy policy.
func GetImageRemoteCheckStatus() (ImageRemoteCheckStatus, error) {
	client, err := dockerhub.NewHTTPClient()
	if err != nil {
		var business buserr.BusinessError
		if errors.As(err, &business) && business.Msg == "ErrImageCheckProxyRequired" {
			return ImageRemoteCheckStatus{ProxyRequired: true}, nil
		}
		return ImageRemoteCheckStatus{}, err
	}
	if client.Transport != nil {
		client.CloseIdleConnections()
	}
	return ImageRemoteCheckStatus{}, nil
}
