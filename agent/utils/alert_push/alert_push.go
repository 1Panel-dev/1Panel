package alert_push

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	alertUtil "github.com/1Panel-dev/1Panel/agent/utils/alert"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/jinzhu/copier"
	"strconv"
	"strings"
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
		switch m {
		case constant.SMS:
			if !alertUtil.CheckSMSSendLimit(constant.SMS) {
				continue
			}
			todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), constant.SMS)
			if err != nil || alert.SendCount <= todayCount {
				continue
			}
			var create = dto.AlertLogCreate{
				Type:    alertUtil.GetCronJobType(alert.Type),
				AlertId: alert.ID,
				Count:   todayCount + 1,
			}
			err = xpack.AlertProvider.CreateTaskScanSMSAlertLog(alert, alert.Type, create, pushAlert, constant.SMS)
			if err != nil {
				global.LOG.Errorf("%s alert sms push failed: %v", alert.Type, err)
				continue
			}
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), constant.SMS)
		case constant.Email:
			todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), constant.Email)
			if err != nil || alert.SendCount <= todayCount {
				continue
			}
			var create = dto.AlertLogCreate{
				Type:    alertUtil.GetCronJobType(alert.Type),
				AlertId: alert.ID,
				Count:   todayCount + 1,
			}
			transport := xpack.MultiNodeProvider.LoadRequestTransport()
			agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
			err = alertUtil.CreateTaskScanEmailAlertLog(alert, create, pushAlert, constant.Email, transport, agentInfo)
			if err != nil {
				global.LOG.Errorf("%s alert email push failed: %v", alert.Type, err)
				continue
			}
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), constant.Email)
		case constant.Bark:
			todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), constant.Bark)
			if err != nil || alert.SendCount <= todayCount {
				continue
			}
			var create = dto.AlertLogCreate{
				Type:    alertUtil.GetCronJobType(alert.Type),
				AlertId: alert.ID,
				Count:   todayCount + 1,
			}
			transport := xpack.MultiNodeProvider.LoadRequestTransport()
			agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
			params := alertUtil.CreateAlertParams(alertUtil.GetCronJobTypeName(pushAlert.Param))
			alertDetail := alertUtil.ProcessAlertDetail(alert, pushAlert.TaskName, params, constant.Bark)
			alertRule := alertUtil.ProcessAlertRule(alert)
			create.AlertRule = alertRule
			create.AlertDetail = alertDetail
			err = alertUtil.CreateBarkAlertLog(create, alert, params, transport, agentInfo)
			if err != nil {
				global.LOG.Errorf("%s alert bark push failed: %v", alert.Type, err)
				continue
			}
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), constant.Bark)
		case constant.WeCom, constant.DingTalk, constant.FeiShu:
			todayCount, _, err := alertRepo.LoadTaskCount(alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), m)
			if err != nil || alert.SendCount <= todayCount {
				continue
			}
			var create = dto.AlertLogCreate{
				Type:    alertUtil.GetCronJobType(alert.Type),
				AlertId: alert.ID,
				Count:   todayCount + 1,
			}
			transport := xpack.MultiNodeProvider.LoadRequestTransport()
			agentInfo, _ := xpack.MultiNodeProvider.GetAgentInfo()
			err = xpack.AlertProvider.CreateTaskScanWebhookAlertLog(alert, alert.Type, create, pushAlert, m, transport, agentInfo)
			if err != nil {
				global.LOG.Errorf("%s alert %s webhook push failed: %v", alert.Type, m, err)
				continue
			}
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), m)
		default:
		}
	}
	return nil
}
