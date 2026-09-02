package helper

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	alertUtil "github.com/1Panel-dev/1Panel/agent/utils/alert"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack/providers"
)

type alertHelper struct{}

var (
	_ providers.CustomWebhookTester           = (*alertHelper)(nil)
	_ providers.CustomWebhookDeliveryProvider = (*alertHelper)(nil)
)

var loadCommunityCustomWebhookContext = func() (*http.Transport, *dto.AgentInfo) {
	multiNode := &multiNodeHelper{}
	agentInfo, _ := multiNode.GetAgentInfo()
	return multiNode.LoadRequestTransport(), agentInfo
}

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
	if config.Type == constant.Custom {
		return alertUtil.CreateTaskScanCustomWebhookAlertLog(alert, alertType, create, pushAlert, config, transport, agentInfo)
	}
	return nil
}

func (a *alertHelper) CreateWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	if config.Type == constant.Custom {
		return alertUtil.CreateCustomWebhookAlertLog(alertType, info, create, project, params, config, transport, agentInfo)
	}
	return nil
}

func (a *alertHelper) CreateCustomWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo, _ dto.AlertTaskMetadata) (providers.DeliveryResult, error) {
	err := alertUtil.CreateCustomWebhookAlertLog(alertType, info, create, project, params, config, transport, agentInfo)
	return providers.DeliveryResult{}, err
}

func (a *alertHelper) CreateTaskScanCustomWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo, _ dto.AlertTaskMetadata) (providers.DeliveryResult, error) {
	err := alertUtil.CreateTaskScanCustomWebhookAlertLog(alert, alertType, create, pushAlert, config, transport, agentInfo)
	return providers.DeliveryResult{}, err
}

func (a *alertHelper) TestCustomWebhook(config dto.AlertCustomWebhookResolvedConfig) (dto.AlertConfigTestResult, error) {
	transport, agentInfo := loadCommunityCustomWebhookContext()
	return alertUtil.TestCustomWebhook(config, transport, agentInfo)
}

func (a *alertHelper) GetLicenseErrorAlert() (uint, error) {
	return 0, nil
}

func (a *alertHelper) GetNodeErrorAlert() (uint, error) {
	return 0, nil
}
