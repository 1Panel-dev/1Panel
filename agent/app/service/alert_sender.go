package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	alertUtil "github.com/1Panel-dev/1Panel/agent/utils/alert"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"strings"
)

type AlertSender struct {
	alert     dto.AlertDTO
	quotaType string
}

func NewAlertSender(alert dto.AlertDTO, quotaType string) *AlertSender {
	return &AlertSender{
		alert:     alert,
		quotaType: quotaType,
	}
}

func (s *AlertSender) Send(quota string, params []dto.Param) {
	methods := strings.Split(s.alert.Method, ",")
	for _, method := range methods {
		method = strings.TrimSpace(method)
		switch method {
		case constant.SMS:
			s.sendSMS(quota, params)
		case constant.Email:
			s.sendEmail(quota, params)
		case constant.WeCom, constant.DingTalk, constant.FeiShu:
			s.sendWebhook(quota, params, method)
		}
	}
}

func (s *AlertSender) ResourceSend(quota string, params []dto.Param) {
	methods := strings.Split(s.alert.Method, ",")
	for _, method := range methods {
		method = strings.TrimSpace(method)
		switch method {
		case constant.SMS:
			s.sendResourceSMS(quota, params)
		case constant.Email:
			s.sendResourceEmail(quota, params)
		case constant.WeCom, constant.DingTalk, constant.FeiShu:
			s.sendResourceWebhook(quota, params, method)
		}
	}
}

func (s *AlertSender) sendSMS(quota string, params []dto.Param) {
	if !alertUtil.CheckSMSSendLimit(constant.SMS) {
		return
	}

	totalCount, isValid := s.canSendAlert(constant.SMS)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   totalCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
	}

	err := xpack.CreateSMSAlertLog(s.alert.Type, s.alert, create, quota, params, constant.SMS)
	if err != nil {
		global.LOG.Errorf("%s alert sms push failed: %v", s.alert.Type, err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, constant.SMS)
}

func (s *AlertSender) sendEmail(quota string, params []dto.Param) {
	totalCount, isValid := s.canSendAlert(constant.Email)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:      constant.AlertSuccess,
		Count:       totalCount + 1,
		AlertId:     s.alert.ID,
		Type:        s.alert.Type,
		AlertRule:   alertUtil.ProcessAlertRule(s.alert),
		AlertDetail: alertUtil.ProcessAlertDetail(s.alert, quota, params, constant.Email),
	}

	transport := xpack.LoadRequestTransport()
	agentInfo, _ := xpack.GetAgentInfo()
	err := alertUtil.CreateEmailAlertLog(create, s.alert, params, transport, agentInfo)
	if err != nil {
		global.LOG.Errorf("%s alert email push failed: %v", s.alert.Type, err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, constant.Email)
}

func (s *AlertSender) sendWebhook(quota string, params []dto.Param, method string) {
	totalCount, isValid := s.canSendAlert(method)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   totalCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
	}
	transport := xpack.LoadRequestTransport()
	agentInfo, _ := xpack.GetAgentInfo()
	err := xpack.CreateWebhookAlertLog(s.alert.Type, s.alert, create, quota, params, method, transport, agentInfo)
	if err != nil {
		global.LOG.Errorf("%s alert %s webhook push failed: %v", s.alert.Type, method, err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, method)
}

func (s *AlertSender) sendResourceSMS(quota string, params []dto.Param) {
	if !alertUtil.CheckSMSSendLimit(constant.SMS) {
		return
	}

	todayCount, isValid := s.canResourceSendAlert(constant.SMS)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   todayCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
	}

	if err := xpack.CreateSMSAlertLog(s.alert.Type, s.alert, create, quota, params, constant.SMS); err != nil {
		global.LOG.Errorf("failed to send SMS alert: %v", err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, constant.SMS)
}

func (s *AlertSender) sendResourceEmail(quota string, params []dto.Param) {
	todayCount, isValid := s.canResourceSendAlert(constant.Email)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:      constant.AlertSuccess,
		Count:       todayCount + 1,
		AlertId:     s.alert.ID,
		Type:        s.alert.Type,
		AlertRule:   alertUtil.ProcessAlertRule(s.alert),
		AlertDetail: alertUtil.ProcessAlertDetail(s.alert, quota, params, constant.Email),
	}

	transport := xpack.LoadRequestTransport()
	agentInfo, _ := xpack.GetAgentInfo()
	if err := alertUtil.CreateEmailAlertLog(create, s.alert, params, transport, agentInfo); err != nil {
		global.LOG.Errorf("failed to send Email alert: %v", err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, constant.Email)
}

func (s *AlertSender) sendResourceWebhook(quota string, params []dto.Param, method string) {
	todayCount, isValid := s.canResourceSendAlert(method)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   todayCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
	}
	transport := xpack.LoadRequestTransport()
	agentInfo, _ := xpack.GetAgentInfo()
	if err := xpack.CreateWebhookAlertLog(s.alert.Type, s.alert, create, quota, params, method, transport, agentInfo); err != nil {
		global.LOG.Errorf("%s alert %s webhook push failed: %v", s.alert.Type, method, err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, method)
}

func (s *AlertSender) canSendAlert(method string) (uint, bool) {
	todayCount, totalCount, err := alertRepo.LoadTaskCount(s.alert.Type, s.quotaType, method)
	if err != nil {
		global.LOG.Errorf("error getting task count: %v", err)
		return totalCount, false
	}

	if todayCount >= 1 || s.alert.SendCount <= totalCount {
		return totalCount, false
	}
	return totalCount, true
}

func (s *AlertSender) canResourceSendAlert(method string) (uint, bool) {
	todayCount, _, err := alertRepo.LoadTaskCount(s.alert.Type, s.quotaType, method)
	if err != nil {
		global.LOG.Errorf("error getting task count: %v", err)
		return todayCount, false
	}
	if s.alert.SendCount <= todayCount {
		return todayCount, false
	}
	return todayCount, true
}
