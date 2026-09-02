package alert_push

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
	"github.com/jinzhu/copier"
)

func PushAlert(pushAlert dto.PushAlert) error {
	if !alertUtil.CheckSendTimeRange(alertUtil.GetCronJobType(pushAlert.AlertType)) {
		return nil
	}

	alertRepo := repo.NewIAlertRepo()
	alertInfo, err := alertRepo.Get(alertRepo.WithByType(pushAlert.AlertType), alertRepo.WithByProject(strconv.Itoa(int(pushAlert.EntryID))), repo.WithByStatus(constant.AlertEnable))
	if err != nil {
		return err
	}
	var alert dto.AlertDTO
	_ = copier.Copy(&alert, &alertInfo)

	methods := strings.Split(alert.Method, ",")
	for _, m := range methods {
		m = strings.TrimSpace(m)
		if configId, err := strconv.ParseUint(m, 10, 64); err == nil {
			pushByConfigId(alertRepo, alert, pushAlert, uint(configId))
		} else {
			pushByLegacyMethod(alertRepo, alert, pushAlert, m)
		}
	}
	return nil
}

func pushByConfigId(alertRepo repo.IAlertRepo, alert dto.AlertDTO, pushAlert dto.PushAlert, configId uint) {
	config, err := alertRepo.GetConfigById(configId)
	if err != nil {
		global.LOG.Errorf("alert config not found for id %d: %v", configId, err)
		return
	}
	sendAlert(alertRepo, alert, pushAlert, config)
}

func pushByLegacyMethod(alertRepo repo.IAlertRepo, alert dto.AlertDTO, pushAlert dto.PushAlert, method string) {
	typeMap := map[string]string{
		"mail":          constant.Email,
		constant.Bark:   constant.Bark,
		constant.SMS:    constant.SMS,
		constant.Custom: constant.Custom,
	}
	configType := method
	if mapped, ok := typeMap[method]; ok {
		configType = mapped
	}
	config, err := alertRepo.GetConfig(alertRepo.WithByType(configType))
	if err != nil {
		return
	}
	sendAlert(alertRepo, alert, pushAlert, config)
}

func sendAlert(alertRepo repo.IAlertRepo, alert dto.AlertDTO, pushAlert dto.PushAlert, config model.AlertConfig) {
	if !alertUtil.IsAlertConfigEnabled(config) {
		return
	}
	methodStr := strconv.Itoa(int(config.ID))
	switch config.Type {
	case constant.SMS:
		if !alertUtil.CheckSMSSendLimit(config, methodStr) {
			return
		}
		todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)
		if err != nil || alert.SendCount <= todayCount {
			return
		}
		create := dto.AlertLogCreate{
			Type:    alertUtil.GetCronJobType(alert.Type),
			AlertId: alert.ID,
			Count:   todayCount + 1,
			Method:  methodStr,
		}
		err = xpack.AlertProvider.CreateTaskScanSMSAlertLog(alert, alert.Type, create, pushAlert, config, methodStr)
		if err != nil {
			global.LOG.Errorf("%s alert sms push failed: %v", alert.Type, err)
			return
		}
		alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)

	case constant.Email:
		todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)
		if err != nil || alert.SendCount <= todayCount {
			return
		}
		create := dto.AlertLogCreate{
			Type:    alertUtil.GetCronJobType(alert.Type),
			AlertId: alert.ID,
			Count:   todayCount + 1,
			Method:  methodStr,
		}
		transport := xpack.MultiNodeProvider.LoadRequestTransport()
		agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
		err = alertUtil.CreateTaskScanEmailAlertLog(alert, create, pushAlert, constant.Email, transport, agentInfo, config)
		if err != nil {
			global.LOG.Errorf("%s alert email push failed: %v", alert.Type, err)
			return
		}
		alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)

	case constant.Bark:
		todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)
		if err != nil || alert.SendCount <= todayCount {
			return
		}
		create := dto.AlertLogCreate{
			Type:    alertUtil.GetCronJobType(alert.Type),
			AlertId: alert.ID,
			Count:   todayCount + 1,
			Method:  methodStr,
		}
		transport := xpack.MultiNodeProvider.LoadRequestTransport()
		agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
		params := alertUtil.CreateAlertParams(alertUtil.GetCronJobTypeName(pushAlert.Param))
		alertDetail := alertUtil.ProcessAlertDetail(alert, pushAlert.TaskName, params, constant.Bark)
		alertRule := alertUtil.ProcessAlertRule(alert)
		create.AlertRule = alertRule
		create.AlertDetail = alertDetail
		err = alertUtil.CreateBarkAlertLog(create, alert, params, transport, agentInfo, config)
		if err != nil {
			global.LOG.Errorf("%s alert bark push failed: %v", alert.Type, err)
			return
		}
		alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)

	case constant.WeCom, constant.DingTalk, constant.FeiShu, constant.Custom:
		todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)
		if err != nil || alert.SendCount <= todayCount {
			return
		}
		create := dto.AlertLogCreate{
			Type:    alertUtil.GetCronJobType(alert.Type),
			AlertId: alert.ID,
			Count:   todayCount + 1,
			Method:  methodStr,
		}
		transport := xpack.MultiNodeProvider.LoadRequestTransport()
		agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
		queued := false
		if config.Type == constant.Custom {
			task := dto.AlertTaskMetadata{
				AlertID:   alert.ID,
				Type:      alertUtil.GetCronJobType(alert.Type),
				Quota:     strconv.Itoa(int(pushAlert.EntryID)),
				QuotaType: strconv.Itoa(int(pushAlert.EntryID)),
				Method:    methodStr,
			}
			result, deliveryErr := xpack.DeliverTaskScanCustomWebhookAlertLog(alert, alert.Type, create, pushAlert, config, transport, agentInfo, task)
			queued, err = result.Queued, deliveryErr
			if err == nil && result.Queued {
				_, err = alertUtil.RecordQueuedAlertTask(result.LogID, task)
			}
		} else {
			err = xpack.AlertProvider.CreateTaskScanWebhookAlertLog(alert, alert.Type, create, pushAlert, config, transport, agentInfo)
		}
		if err != nil {
			global.LOG.Errorf("%s alert %s webhook push failed: %v", alert.Type, methodStr, err)
			return
		}
		if !queued {
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), methodStr)
		}
	}
}
