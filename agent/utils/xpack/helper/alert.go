package helper

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack/providers"
)

type alertHelper struct{}

func NewIAlertProvider() providers.AlertProvider {
	return &alertHelper{}
}

func (a *alertHelper) CreateTaskScanSMSAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, method string) error {
	return nil
}

func (a *alertHelper) CreateSMSAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, method string) error {
	return nil
}

func (a *alertHelper) CreateTaskScanWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	return nil
}

func (a *alertHelper) CreateWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	return nil
}

func (a *alertHelper) GetLicenseErrorAlert() (uint, error) {
	return 0, nil
}

func (a *alertHelper) GetNodeErrorAlert() (uint, error) {
	return 0, nil
}
