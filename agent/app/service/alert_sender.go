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

	_ = xpack.CreateSMSAlertLog(s.alert.Type, s.alert, create, quota, params, constant.SMS)
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, constant.SMS)
	global.LOG.Infof("%s alert sms push successful", s.alert.Type)
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
	_ = alertUtil.CreateEmailAlertLog(create, s.alert, params, transport)
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, constant.Email)
	global.LOG.Infof("%s alert email push successful", s.alert.Type)
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
	global.LOG.Infof("%s alert sms push successful", s.alert.Type)
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
	if err := alertUtil.CreateEmailAlertLog(create, s.alert, params, transport); err != nil {
		global.LOG.Errorf("failed to send Email alert: %v", err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, constant.Email)
	global.LOG.Infof("%s alert email push successful", s.alert.Type)
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
