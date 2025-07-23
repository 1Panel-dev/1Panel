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

	// 根据发送方式推送不同的日志记录
	methods := strings.Split(alert.Method, ",")
	for _, m := range methods {
		m = strings.TrimSpace(m)
		switch m {
		case constant.SMS:
			if !alertUtil.CheckTaskFrequency(constant.SMS) {
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
			_ = xpack.CreateTaskScanSMSAlertLog(alert, create, pushAlert, constant.SMS)
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
			transport := xpack.LoadRequestTransport()
			err = alertUtil.CreateTaskScanEmailAlertLog(alert, create, pushAlert, constant.Email, transport)
			if err != nil {
				return err
			}
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(pushAlert.EntryID)), alertUtil.GetCronJobType(alert.Type), strconv.Itoa(int(pushAlert.EntryID)), constant.Email)
		default:
		}
	}
	global.LOG.Infof("%s alert push successful", alert.Type)
	return nil
}
