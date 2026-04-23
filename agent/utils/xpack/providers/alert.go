package providers

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

type AlertProvider interface {
	GetNodeErrorAlert() (uint, error)
	GetLicenseErrorAlert() (uint, error)

	CreateTaskScanSMSAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string) error
	CreateSMSAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, method string) error
	CreateTaskScanWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string, transport *http.Transport, agentInfo *dto.AgentInfo) error
	CreateWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, method string, transport *http.Transport, agentInfo *dto.AgentInfo) error
}
