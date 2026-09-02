package providers

import (
	"errors"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
)

var ErrCustomWebhookUnsupported = errors.New("custom webhook sender is unavailable")

type AlertProvider interface {
	GetNodeErrorAlert() (uint, error)
	GetLicenseErrorAlert() (uint, error)

	CreateTaskScanSMSAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, method string) error
	CreateSMSAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, method string) error
	CreateTaskScanWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo) error
	CreateWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo) error
}

type CustomWebhookTester interface {
	TestCustomWebhook(config dto.AlertCustomWebhookResolvedConfig) (dto.AlertConfigTestResult, error)
}

type DeliveryResult struct {
	Queued bool
	LogID  uint
}

type CustomWebhookDeliveryProvider interface {
	CreateCustomWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo, task dto.AlertTaskMetadata) (DeliveryResult, error)
	CreateTaskScanCustomWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo, task dto.AlertTaskMetadata) (DeliveryResult, error)
}
