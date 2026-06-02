package helper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	alertUtil "github.com/1Panel-dev/1Panel/agent/utils/alert"
	"github.com/1Panel-dev/1Panel/agent/utils/bark"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack/providers"
)

type alertHelper struct{}

func NewIAlertProvider() providers.AlertProvider {
	return &alertHelper{}
}

func (a *alertHelper) CreateTaskScanSMSAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, method string) error {
	params := alertUtil.CreateAlertParams(alertUtil.GetCronJobTypeName(pushAlert.Param))
	create.AlertRule = alertUtil.ProcessAlertRule(alert)
	create.AlertDetail = alertUtil.ProcessAlertDetail(alert, pushAlert.TaskName, params, method)
	if create.Status == "" {
		create.Status = constant.AlertSuccess
	}
	return alertUtil.SaveAlertLog(create, &model.AlertLog{})
}

func (a *alertHelper) CreateSMSAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, method string) error {
	create.AlertRule = alertUtil.ProcessAlertRule(info)
	if create.AlertDetail == "" {
		create.AlertDetail = alertUtil.ProcessAlertDetail(info, project, params, method)
	}
	if create.Status == "" {
		create.Status = constant.AlertSuccess
	}
	return alertUtil.SaveAlertLog(create, &model.AlertLog{})
}

func (a *alertHelper) CreateTaskScanWebhookAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	params := alertUtil.CreateAlertParams(alertUtil.GetCronJobTypeName(pushAlert.Param))
	create.AlertRule = alertUtil.ProcessAlertRule(alert)
	create.AlertDetail = alertUtil.ProcessAlertDetail(alert, pushAlert.TaskName, params, config.Type)
	return a.CreateWebhookAlertLog(alertType, alert, create, pushAlert.TaskName, params, config, transport, agentInfo)
}

func (a *alertHelper) CreateWebhookAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, config model.AlertConfig, transport *http.Transport, agentInfo *dto.AgentInfo) error {
	var webhookInfo dto.AlertWebhookConfig
	if err := json.Unmarshal([]byte(config.Config), &webhookInfo); err != nil {
		create.Message = err.Error()
		create.Status = constant.AlertError
		return alertUtil.SaveAlertLog(create, &model.AlertLog{})
	}
	if webhookInfo.Url == "" {
		create.Message = "webhook url is required"
		create.Status = constant.AlertError
		return alertUtil.SaveAlertLog(create, &model.AlertLog{})
	}
	create.AlertRule = alertUtil.ProcessAlertRule(info)
	if create.AlertDetail == "" {
		create.AlertDetail = alertUtil.ProcessAlertDetail(info, project, params, config.Type)
	}
	content := alertUtil.GetSendContent(info.Type, params, agentInfo)
	if content == "" {
		content = fmt.Sprintf("%s: %s", info.Type, info.Title)
	}
	if err := bark.SendMessage(webhookInfo.Url, "1Panel Alert", content, transport); err != nil {
		create.Message = err.Error()
		create.Status = constant.AlertError
		return alertUtil.SaveAlertLog(create, &model.AlertLog{})
	}
	if create.Status == "" {
		create.Status = constant.AlertSuccess
	}
	return alertUtil.SaveAlertLog(create, &model.AlertLog{})
}

func (a *alertHelper) GetLicenseErrorAlert() (uint, error) {
	return 0, nil
}

func (a *alertHelper) GetNodeErrorAlert() (uint, error) {
	return 0, nil
}
