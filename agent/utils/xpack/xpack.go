//go:build xpack

package xpack

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	edition "github.com/1Panel-dev/1Panel/agent/xpack/edition"
	"github.com/gin-gonic/gin"
)

func RemoveTamper(website string) {
	edition.RemoveTamper(website)
}

func StartClam(startClam *model.Clam, isUpdate bool) (int, error) {
	return edition.StartClam(startClam, isUpdate)
}

func LoadNodeInfo(isBase bool) (model.NodeInfo, error) {
	return edition.LoadNodeInfo(isBase)
}

func GetImagePrefix() string {
	return edition.GetImagePrefix()
}

func IsUseCustomApp() bool {
	return edition.IsUseCustomApp()
}

func IsXpack() bool {
	return edition.IsXpack()
}

func CreateTaskScanSMSAlertLog(info dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string) error {
	return edition.CreateTaskScanSMSAlertLog(info, alertType, create, pushAlert, method)
}

func CreateSMSAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, method string) error {
	return edition.CreateSMSAlertLog(alertType, info, create, project, params, method)
}

func CreateTaskScanWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	return edition.CreateTaskScanWebhookAlertLog(alert, alertType, create, pushAlert, method, transport, agentInfo)
}

func CreateWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, method string, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	return edition.CreateWebhookAlertLog(alertType, info, create, project, params, method, transport, agentInfo)
}

func GetLicenseErrorAlert() (uint, error) {
	return edition.GetLicenseErrorAlert()
}

func GetNodeErrorAlert() (uint, error) {
	return edition.GetNodeErrorAlert()
}

func LoadRequestTransport() *http.Transport { return edition.LoadRequestTransport() }

func ValidateCertificate(c *gin.Context) bool {
	return edition.ValidateCertificate(c)
}

func PushSSLToNode(websiteSSL *model.WebsiteSSL) error {
	return edition.PushSSLToNode(websiteSSL)
}

func GetAgentInfo() (*dto.AgentInfo, error) {
	return edition.GetAgentInfo()
}
