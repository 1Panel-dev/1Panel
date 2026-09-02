package service

import (
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	alertUtil "github.com/1Panel-dev/1Panel/agent/utils/alert"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
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
	s.sendByConfigIds(s.alert.Method, quota, params, false)
}

func (s *AlertSender) ResourceSend(quota string, params []dto.Param) {
	s.sendByConfigIds(s.alert.Method, quota, params, true)
}

func (s *AlertSender) sendByConfigIds(methodStr string, quota string, params []dto.Param, isResource bool) {
	alertRepo := repo.NewIAlertRepo()
	configIds := strings.Split(methodStr, ",")
	for _, idStr := range configIds {
		idStr = strings.TrimSpace(idStr)
		configId, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			s.sendByLegacyMethod(idStr, quota, params, isResource)
			continue
		}
		config, err := alertRepo.GetConfigById(uint(configId))
		if err != nil {
			global.LOG.Errorf("alert config not found for id %d: %v", configId, err)
			continue
		}
		s.sendByConfig(config, quota, params, isResource)
	}
}

func (s *AlertSender) sendByConfig(config model.AlertConfig, quota string, params []dto.Param, isResource bool) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	switch config.Type {
	case constant.SMS:
		if isResource {
			s.sendResourceSMSWithConfig(config, quota, params)
		} else {
			s.sendSMSWithConfig(config, quota, params)
		}
	case constant.Email:
		if isResource {
			s.sendResourceEmailWithConfig(config, quota, params)
		} else {
			s.sendEmailWithConfig(config, quota, params)
		}
	case constant.Bark:
		if isResource {
			s.sendResourceBarkWithConfig(config, quota, params)
		} else {
			s.sendBarkWithConfig(config, quota, params)
		}
	case constant.WeCom, constant.DingTalk, constant.FeiShu, constant.Custom:
		if isResource {
			s.sendResourceWebhookWithConfig(config, quota, params)
		} else {
			s.sendWebhookWithConfig(config, quota, params)
		}
	}
}

func (s *AlertSender) sendByLegacyMethod(method string, quota string, params []dto.Param, isResource bool) {
	alertRepo := repo.NewIAlertRepo()
	typeMap := map[string]string{"mail": constant.Email, constant.Bark: constant.Bark, constant.SMS: constant.SMS, constant.Custom: constant.Custom}
	configType := method
	if mapped, ok := typeMap[method]; ok {
		configType = mapped
	}
	config, err := alertRepo.GetConfig(alertRepo.WithByType(configType))
	if err != nil {
		return
	}
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	s.sendByConfig(config, quota, params, isResource)
}

func (s *AlertSender) sendSMSWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	method := strconv.Itoa(int(config.ID))
	if !alertUtil.CheckSMSSendLimit(config, method) {
		return
	}

	totalCount, isValid := s.canSendAlert(method)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   totalCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
		Method:  method,
	}

	err := xpack.AlertProvider.CreateSMSAlertLog(s.alert.Type, s.alert, create, quota, params, config, method)
	if err != nil {
		global.LOG.Errorf("%s alert sms push failed: %v", s.alert.Type, err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, method)
}

func (s *AlertSender) sendSMS(quota string, params []dto.Param) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.SMS))
	if err != nil {
		return
	}
	s.sendSMSWithConfig(config, quota, params)
}

func (s *AlertSender) sendEmailWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	totalCount, isValid := s.canSendAlert(strconv.Itoa(int(config.ID)))
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
		Method:      strconv.Itoa(int(config.ID)),
	}

	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
	err := alertUtil.CreateEmailAlertLog(create, s.alert, params, transport, agentInfo, config)
	if err != nil {
		global.LOG.Errorf("%s alert email push failed: %v", s.alert.Type, err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, strconv.Itoa(int(config.ID)))
}

func (s *AlertSender) sendResourceEmailWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	todayCount, isValid := s.canResourceSendAlert(strconv.Itoa(int(config.ID)))
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
		Method:      strconv.Itoa(int(config.ID)),
	}

	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
	if err := alertUtil.CreateEmailAlertLog(create, s.alert, params, transport, agentInfo, config); err != nil {
		global.LOG.Errorf("failed to send Email alert: %v", err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, strconv.Itoa(int(config.ID)))
}

func (s *AlertSender) sendEmail(quota string, params []dto.Param) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.EmailConfig))
	if err != nil {
		return
	}
	s.sendEmailWithConfig(config, quota, params)
}

func (s *AlertSender) sendResourceEmail(quota string, params []dto.Param) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.EmailConfig))
	if err != nil {
		return
	}
	s.sendResourceEmailWithConfig(config, quota, params)
}

func (s *AlertSender) sendBarkWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	totalCount, isValid := s.canSendAlert(strconv.Itoa(int(config.ID)))
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:      constant.AlertSuccess,
		Count:       totalCount + 1,
		AlertId:     s.alert.ID,
		Type:        s.alert.Type,
		AlertRule:   alertUtil.ProcessAlertRule(s.alert),
		AlertDetail: alertUtil.ProcessAlertDetail(s.alert, quota, params, constant.Bark),
		Method:      strconv.Itoa(int(config.ID)),
	}

	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
	err := alertUtil.CreateBarkAlertLog(create, s.alert, params, transport, agentInfo, config)
	if err != nil {
		global.LOG.Errorf("%s alert bark push failed: %v", s.alert.Type, err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, strconv.Itoa(int(config.ID)))
}

func (s *AlertSender) sendResourceBarkWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	todayCount, isValid := s.canResourceSendAlert(strconv.Itoa(int(config.ID)))
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:      constant.AlertSuccess,
		Count:       todayCount + 1,
		AlertId:     s.alert.ID,
		Type:        s.alert.Type,
		AlertRule:   alertUtil.ProcessAlertRule(s.alert),
		AlertDetail: alertUtil.ProcessAlertDetail(s.alert, quota, params, constant.Bark),
		Method:      strconv.Itoa(int(config.ID)),
	}

	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
	if err := alertUtil.CreateBarkAlertLog(create, s.alert, params, transport, agentInfo, config); err != nil {
		global.LOG.Errorf("failed to send Bark alert: %v", err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, strconv.Itoa(int(config.ID)))
}

func (s *AlertSender) sendBark(quota string, params []dto.Param) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.Bark))
	if err != nil {
		return
	}
	s.sendBarkWithConfig(config, quota, params)
}

func (s *AlertSender) sendResourceBark(quota string, params []dto.Param) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.Bark))
	if err != nil {
		return
	}
	s.sendResourceBarkWithConfig(config, quota, params)
}

func (s *AlertSender) sendWebhookWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	totalCount, isValid := s.canSendAlert(strconv.Itoa(int(config.ID)))
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   totalCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
		Method:  strconv.Itoa(int(config.ID)),
	}
	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
	queued := false
	var err error
	if config.Type == constant.Custom {
		task := dto.AlertTaskMetadata{
			AlertID:   s.alert.ID,
			Type:      s.alert.Type,
			Quota:     quota,
			QuotaType: s.quotaType,
			Method:    strconv.Itoa(int(config.ID)),
		}
		result, deliveryErr := xpack.DeliverCustomWebhookAlertLog(s.alert.Type, s.alert, create, quota, params, config, transport, agentInfo, task)
		queued, err = result.Queued, deliveryErr
		if err == nil && result.Queued {
			_, err = alertUtil.RecordQueuedAlertTask(result.LogID, task)
		}
	} else {
		err = xpack.AlertProvider.CreateWebhookAlertLog(s.alert.Type, s.alert, create, quota, params, config, transport, agentInfo)
	}
	if err != nil {
		global.LOG.Errorf("%s alert %s webhook push failed: %v", s.alert.Type, config.Type, err)
		return
	}
	if !queued {
		alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, strconv.Itoa(int(config.ID)))
	}
}

func (s *AlertSender) sendResourceWebhookWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	todayCount, isValid := s.canResourceSendAlert(strconv.Itoa(int(config.ID)))
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   todayCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
		Method:  strconv.Itoa(int(config.ID)),
	}
	transport := xpack.MultiNodeProvider.LoadRequestTransport()
	agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
	queued := false
	var err error
	if config.Type == constant.Custom {
		task := dto.AlertTaskMetadata{
			AlertID:   s.alert.ID,
			Type:      s.alert.Type,
			Quota:     quota,
			QuotaType: s.quotaType,
			Method:    strconv.Itoa(int(config.ID)),
		}
		result, deliveryErr := xpack.DeliverCustomWebhookAlertLog(s.alert.Type, s.alert, create, quota, params, config, transport, agentInfo, task)
		queued, err = result.Queued, deliveryErr
		if err == nil && result.Queued {
			_, err = alertUtil.RecordQueuedAlertTask(result.LogID, task)
		}
	} else {
		err = xpack.AlertProvider.CreateWebhookAlertLog(s.alert.Type, s.alert, create, quota, params, config, transport, agentInfo)
	}
	if err != nil {
		global.LOG.Errorf("%s alert %s webhook push failed: %v", s.alert.Type, config.Type, err)
		return
	}
	if !queued {
		alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, strconv.Itoa(int(config.ID)))
	}
}

func (s *AlertSender) sendWebhook(quota string, params []dto.Param, method string) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(method))
	if err != nil {
		return
	}
	s.sendWebhookWithConfig(config, quota, params)
}

func (s *AlertSender) sendResourceWebhook(quota string, params []dto.Param, method string) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(method))
	if err != nil {
		return
	}
	s.sendResourceWebhookWithConfig(config, quota, params)
}

func (s *AlertSender) sendResourceSMSWithConfig(config model.AlertConfig, quota string, params []dto.Param) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	method := strconv.Itoa(int(config.ID))
	if !alertUtil.CheckSMSSendLimit(config, method) {
		return
	}

	todayCount, isValid := s.canResourceSendAlert(method)
	if !isValid {
		return
	}

	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   todayCount + 1,
		AlertId: s.alert.ID,
		Type:    s.alert.Type,
		Method:  method,
	}

	if err := xpack.AlertProvider.CreateSMSAlertLog(s.alert.Type, s.alert, create, quota, params, config, method); err != nil {
		global.LOG.Errorf("failed to send SMS alert: %v", err)
		return
	}
	alertUtil.CreateNewAlertTask(quota, s.alert.Type, s.quotaType, method)
}

func (s *AlertSender) sendResourceSMS(quota string, params []dto.Param) {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.SMSConfig))
	if err != nil {
		return
	}
	s.sendResourceSMSWithConfig(config, quota, params)
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
