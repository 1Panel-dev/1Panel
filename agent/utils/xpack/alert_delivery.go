package xpack

import (
	"fmt"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack/providers"
)

func DeliverCustomWebhookAlertLog(
	alertType string,
	info dto.AlertDTO,
	create dto.AlertLogCreate,
	project string,
	params []dto.Param,
	config model.AlertConfig,
	transport *http.Transport,
	agentInfo *dto.AgentInfo,
	task dto.AlertTaskMetadata,
) (providers.DeliveryResult, error) {
	if provider, ok := AlertProvider.(providers.CustomWebhookDeliveryProvider); ok {
		result, err := provider.CreateCustomWebhookAlertLog(alertType, info, create, project, params, config, transport, agentInfo, task)
		return validateCustomWebhookDeliveryResult(result, err)
	}
	err := AlertProvider.CreateWebhookAlertLog(alertType, info, create, project, params, config, transport, agentInfo)
	return providers.DeliveryResult{}, err
}

func DeliverTaskScanCustomWebhookAlertLog(
	alert dto.AlertDTO,
	alertType string,
	create dto.AlertLogCreate,
	pushAlert dto.PushAlert,
	config model.AlertConfig,
	transport *http.Transport,
	agentInfo *dto.AgentInfo,
	task dto.AlertTaskMetadata,
) (providers.DeliveryResult, error) {
	if provider, ok := AlertProvider.(providers.CustomWebhookDeliveryProvider); ok {
		result, err := provider.CreateTaskScanCustomWebhookAlertLog(alert, alertType, create, pushAlert, config, transport, agentInfo, task)
		return validateCustomWebhookDeliveryResult(result, err)
	}
	err := AlertProvider.CreateTaskScanWebhookAlertLog(alert, alertType, create, pushAlert, config, transport, agentInfo)
	return providers.DeliveryResult{}, err
}

func validateCustomWebhookDeliveryResult(result providers.DeliveryResult, err error) (providers.DeliveryResult, error) {
	if err != nil {
		return result, err
	}
	if result.Queued && result.LogID == 0 {
		return providers.DeliveryResult{}, fmt.Errorf("custom webhook provider queued a delivery without a log ID")
	}
	return result, nil
}
