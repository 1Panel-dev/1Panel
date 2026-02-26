//go:build !xpack

package xpack

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/gin-gonic/gin"
)

func RemoveTamper(website string) {}

func StartClam(startClam *model.Clam, isUpdate bool) (int, error) {
	return 0, buserr.New("ErrXpackNotFound")
}

func LoadNodeInfo(isBase bool) (model.NodeInfo, error) {
	var info model.NodeInfo
	info.BaseDir = common.LoadParams("BASE_DIR")
	info.Version = common.LoadParams("ORIGINAL_VERSION")
	info.Edition = common.LoadParamsWithoutPanic("PANEL_EDITION")
	info.Scope = "master"
	global.IsMaster = true
	return info, nil
}

func GetImagePrefix() string {
	return ""
}

func IsUseCustomApp() bool {
	return false
}

func CreateTaskScanSMSAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string) error {
	return nil
}

func CreateSMSAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, method string) error {
	return nil
}

func CreateTaskScanWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	return nil
}

func CreateWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, method string, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	return nil
}

func GetLicenseErrorAlert() (uint, error) {
	return 0, nil
}

func GetNodeErrorAlert() (uint, error) {
	return 0, nil
}

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

func ValidateCertificate(c *gin.Context) bool {
	return true
}

func PushSSLToNode(websiteSSL *model.WebsiteSSL) error {
	return nil
}

func GetAgentInfo() (*dto.AgentInfo, error) {
	return nil, nil
}
