package alert

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/alert_webhook"
	"github.com/1Panel-dev/1Panel/agent/utils/webhook_sender"
)

var customWebhookNow = time.Now

func CreateCustomWebhookAlertLog(
	alertType string,
	info dto.AlertDTO,
	create dto.AlertLogCreate,
	project string,
	params []dto.Param,
	config model.AlertConfig,
	transport *http.Transport,
	agentInfo *dto.AgentInfo,
) error {
	alertInfo := info
	alertInfo.Type = alertType
	create.Type = GetCronJobType(alertType)
	create.AlertRule = ProcessAlertRule(info)
	create.AlertDetail = ProcessAlertDetail(alertInfo, project, params, constant.Custom)
	return deliverCustomWebhook(create, config, transport, agentInfo, customWebhookNow())
}

func CreateTaskScanCustomWebhookAlertLog(
	info dto.AlertDTO,
	alertType string,
	create dto.AlertLogCreate,
	pushAlert dto.PushAlert,
	config model.AlertConfig,
	transport *http.Transport,
	agentInfo *dto.AgentInfo,
) error {
	params := CreateAlertParams(GetCronJobTypeName(pushAlert.Param))
	alertInfo := info
	alertInfo.Type = alertType
	create.Type = GetCronJobType(alertType)
	create.AlertRule = ProcessAlertRule(info)
	create.AlertDetail = ProcessAlertDetail(alertInfo, pushAlert.TaskName, params, constant.Custom)
	return deliverCustomWebhook(create, config, transport, agentInfo, customWebhookNow())
}

func deliverCustomWebhook(
	create dto.AlertLogCreate,
	config model.AlertConfig,
	transport *http.Transport,
	agentInfo *dto.AgentInfo,
	occurredAt time.Time,
) error {
	var alertLog model.AlertLog
	templateData, err := customWebhookTemplateData(create.AlertDetail, agentInfo, occurredAt)
	if err != nil {
		return saveCustomWebhookDeliveryError(create, &alertLog, err)
	}
	request, err := buildCustomWebhookRequest(config, templateData, transport)
	if err != nil {
		return saveCustomWebhookDeliveryError(create, &alertLog, err)
	}
	if _, err := webhook_sender.Execute(context.Background(), request); err != nil {
		return saveCustomWebhookDeliveryError(create, &alertLog, err)
	}
	create.Status = constant.AlertSuccess
	return SaveAlertLog(create, &alertLog)
}

func saveCustomWebhookDeliveryError(create dto.AlertLogCreate, alertLog *model.AlertLog, deliveryErr error) error {
	create.Status = constant.AlertError
	create.Message = deliveryErr.Error()
	if err := SaveAlertLog(create, alertLog); err != nil {
		global.LOG.Errorf("save custom webhook delivery error log failed: %v", err)
	}
	return deliveryErr
}

func customWebhookTemplateData(rawDetail string, agentInfo *dto.AgentInfo, occurredAt time.Time) (webhook_sender.TemplateData, error) {
	var detail dto.AlertDetail
	if err := json.Unmarshal([]byte(rawDetail), &detail); err != nil {
		return webhook_sender.TemplateData{}, errors.New("resolve custom webhook alert detail failed")
	}
	businessType := detail.SubType
	if businessType == "" {
		businessType = detail.Type
	}
	if businessType == "" {
		return webhook_sender.TemplateData{}, errors.New("resolve custom webhook alert detail failed")
	}
	content := GetSendContent(businessType, detail.Params, agentInfo)
	if content == "" {
		content = i18n.GetMsgWithMap("CommonAlert", map[string]interface{}{"msg": detail.Title})
	}
	message := strings.TrimSpace(webhook_sender.NormalizeToText(content))
	if message == "" {
		message = detail.Title
	}
	return webhook_sender.TemplateData{
		Title:     detail.Title,
		Message:   message,
		Type:      businessType,
		NodeName:  customWebhookNodeName(agentInfo),
		Timestamp: occurredAt,
	}, nil
}

func customWebhookNodeName(agentInfo *dto.AgentInfo) string {
	if agentInfo != nil && strings.TrimSpace(agentInfo.NodeName) != "" {
		return strings.TrimSpace(agentInfo.NodeName)
	}
	return strings.TrimSpace(getFallbackHostname())
}

func buildCustomWebhookRequest(
	config model.AlertConfig,
	data webhook_sender.TemplateData,
	transport *http.Transport,
) (webhook_sender.Request, error) {
	resolved, err := alert_webhook.Resolve(config)
	if err != nil {
		return webhook_sender.Request{}, errors.New("resolve custom webhook config failed")
	}
	return buildResolvedCustomWebhookRequest(resolved, data, transport)
}

func TestCustomWebhook(
	resolved dto.AlertCustomWebhookResolvedConfig,
	transport *http.Transport,
	agentInfo *dto.AgentInfo,
) (dto.AlertConfigTestResult, error) {
	request, err := buildResolvedCustomWebhookRequest(resolved, webhook_sender.TemplateData{
		Title:     "1Panel Webhook Test",
		Message:   "This is a test notification from 1Panel.",
		Type:      "test",
		NodeName:  customWebhookNodeName(agentInfo),
		Timestamp: customWebhookNow(),
	}, transport)
	if err != nil {
		return dto.AlertConfigTestResult{}, err
	}
	request.CaptureResponse = true
	result, executeErr := webhook_sender.Execute(context.Background(), request)
	durationMillis := result.Duration.Milliseconds()
	if result.Duration > 0 && durationMillis == 0 {
		durationMillis = 1
	}
	testResult := dto.AlertConfigTestResult{
		Success:    executeErr == nil,
		StatusCode: result.StatusCode,
		Duration:   durationMillis,
		Response:   result.Response,
	}
	if executeErr != nil {
		testResult.Message = executeErr.Error()
	}
	return testResult, nil
}

func buildResolvedCustomWebhookRequest(
	resolved dto.AlertCustomWebhookResolvedConfig,
	data webhook_sender.TemplateData,
	transport *http.Transport,
) (webhook_sender.Request, error) {
	preset, err := webhook_sender.ResolvePreset(resolved.Preset)
	if err != nil {
		return webhook_sender.Request{}, errors.New("render custom webhook request failed")
	}
	format := webhook_sender.BodyFormat(resolved.Body.Type)
	fields := make([]webhook_sender.FormField, 0, len(resolved.Body.Fields))
	for _, field := range resolved.Body.Fields {
		fields = append(fields, webhook_sender.FormField{Key: field.Key, Value: field.Value})
	}
	body, err := webhook_sender.RenderBody(webhook_sender.RenderRequest{
		Format:   format,
		Template: resolved.Body.Template,
		Fields:   fields,
		Data:     data,
	})
	if err != nil {
		return webhook_sender.Request{}, errors.New("render custom webhook request failed")
	}
	headers := make(map[string]string, len(resolved.Headers))
	for _, header := range resolved.Headers {
		headers[header.Key] = header.Value
	}
	return webhook_sender.Request{
		URL:       resolved.URL,
		Preset:    preset,
		Format:    format,
		Body:      body,
		Headers:   headers,
		Transport: transport,
	}, nil
}
