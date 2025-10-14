package alert

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/email"
	"github.com/jinzhu/copier"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

var cronJobAlertTypes = []string{"shell", "app", "website", "database", "directory", "log", "snapshot", "curl", "cutWebsiteLog", "clean", "ntp"}

func CreateTaskScanEmailAlertLog(alert dto.AlertDTO, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string, transport *http.Transport) error {
	params := CreateAlertParams(GetCronJobTypeName(pushAlert.Param))
	alertDetail := ProcessAlertDetail(alert, pushAlert.TaskName, params, method)
	alertRule := ProcessAlertRule(alert)
	create.AlertRule = alertRule
	create.AlertDetail = alertDetail
	return CreateEmailAlertLog(create, alert, params, transport)
}

func CreateEmailAlertLog(create dto.AlertLogCreate, alert dto.AlertDTO, params []dto.Param, transport *http.Transport) error {
	var alertLog model.AlertLog
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.CommonConfig))
	if err != nil {
		return err
	}
	var cfg dto.AlertCommonConfig
	err = json.Unmarshal([]byte(config.Config), &cfg)
	if err != nil {
		return err
	}
	create.Method = constant.Email
	// 获取远端推送信息
	if !global.IsMaster && cfg.IsOffline == constant.StatusEnable {
		create.Status = constant.AlertPushing
		return SaveAlertLog(create, &alertLog)
	} else {
		emailConfig, err := alertRepo.GetConfig(alertRepo.WithByType(constant.EmailConfig))
		if err != nil {
			return err
		}
		var emailInfo dto.AlertEmailConfig
		err = json.Unmarshal([]byte(emailConfig.Config), &emailInfo)
		if err != nil {
			return err
		}
		username := emailInfo.UserName
		if username == "" {
			username = emailInfo.Sender
		}
		smtpConfig := email.SMTPConfig{
			Host:       emailInfo.Host,
			Port:       emailInfo.Port,
			Username:   username,
			Password:   emailInfo.Password,
			From:       fmt.Sprintf("%s <%s>", emailInfo.DisplayName, emailInfo.Sender),
			Encryption: emailInfo.Encryption,
			Recipient:  emailInfo.Recipient,
		}
		content := i18n.GetMsgWithMap("CommonAlert", map[string]interface{}{"msg": alert.Title})
		if GetEmailContent(alert.Type, params) != "" {
			content = GetEmailContent(alert.Type, params)
		}
		msg := email.EmailMessage{
			Subject: i18n.GetMsgByKey("PanelAlertTitle"),
			Body:    content,
			IsHTML:  true,
		}

		if err = email.SendMail(smtpConfig, msg, transport); err != nil {
			create.Message = err.Error()
			create.Status = constant.AlertError
			return SaveAlertLog(create, &alertLog)
		}
		create.Status = constant.AlertSuccess
		return SaveAlertLog(create, &alertLog)
	}
}

func SaveAlertLog(create dto.AlertLogCreate, alertLog *model.AlertLog) error {
	alertRepo := repo.NewIAlertRepo()
	if err := copier.Copy(&alertLog, &create); err != nil {
		return buserr.WithErr("ErrStructTransform", err)
	}

	if err := alertRepo.CreateLog(alertLog); err != nil {
		global.LOG.Errorf("Error creating alert logs, err: %v", err)
		return err
	}

	return nil
}

func CreateNewAlertTask(quota, alertType, quotaType, method string) {
	alertRepo := repo.NewIAlertRepo()
	taskBase := model.AlertTask{
		Type:      alertType,
		Quota:     quota,
		QuotaType: quotaType,
		Method:    method,
	}
	err := alertRepo.CreateAlertTask(&taskBase)
	if err != nil {
		global.LOG.Errorf("error creating alert tasks, err: %v", err)
	}
}

func ProcessAlertDetail(alert dto.AlertDTO, project string, params []dto.Param, method string) string {
	alertDetail := dto.AlertDetail{
		Type:    GetCronJobType(alert.Type),
		SubType: alert.Type,
		Title:   alert.Title,
		Method:  method,
		Project: project,
		Params:  params,
	}
	marshal, err := json.Marshal(alertDetail)
	if err != nil {
		global.LOG.Errorf("error processing alert detail, err: %v", err)
		return ""
	}
	return string(marshal)
}

func ProcessAlertRule(alert dto.AlertDTO) string {
	marshal, err := json.Marshal(alert)
	if err != nil {
		global.LOG.Errorf("error processing alert rule, err: %v", err)
		return ""
	}
	return string(marshal)
}

func GetCronJobType(alertType string) string {
	for _, at := range cronJobAlertTypes {
		if at == alertType {
			return "cronJob"
		}
	}
	return alertType
}

func GetCronJobTypeName(cronJobType string) string {
	module := cronJobType
	switch cronJobType {
	case "shell":
		module = "Shell 脚本"
	case "app":
		module = "备份应用"
	case "website":
		module = "备份网站"
	case "database":
		module = "备份数据库"
	case "log":
		module = "备份日志"
	case "directory":
		module = "备份目录"
	case "curl":
		module = "访问 URL"
	case "cutWebsiteLog":
		module = "切割网站日志"
	case "clean":
		module = "缓存清理"
	case "snapshot":
		module = "系统快照"
	case "ntp":
		module = "同步服务器时间"
	default:
	}
	return module
}

func CreateAlertParams(param string) []dto.Param {
	return []dto.Param{
		{
			Index: "1",
			Key:   "param",
			Value: param,
		},
	}
}

var checkTaskMutex sync.Mutex

func CheckSMSSendLimit(method string) bool {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.SMSConfig))
	if err != nil {
		return false
	}
	var cfg dto.AlertSmsConfig
	err = json.Unmarshal([]byte(config.Config), &cfg)
	if err != nil {
		return false
	}
	limitCount := cfg.AlertDailyNum
	checkTaskMutex.Lock()
	defer checkTaskMutex.Unlock()
	todayCount, err := alertRepo.GetLicensePushCount(method)
	if err != nil {
		global.LOG.Errorf("error getting license push count info, err: %v", err)
		return false
	}
	if todayCount >= limitCount {
		return false
	}

	return true
}

type Settings struct {
	NoticeAlert   Category `json:"noticeAlert"`
	ResourceAlert Category `json:"resourceAlert"`
}

type Category struct {
	SendTimeRange string   `json:"sendTimeRange"`
	Type          []string `json:"type"`
}

// CheckSendTimeRange 是否在时间范围内
func CheckSendTimeRange(alertType string) bool {
	alertRepo := repo.NewIAlertRepo()
	config, err := alertRepo.GetConfig(alertRepo.WithByType(constant.CommonConfig))
	if err != nil {
		return false
	}
	var cfg dto.AlertCommonConfig
	err = json.Unmarshal([]byte(config.Config), &cfg)
	if err != nil {
		return false
	}

	var timeRange string
	if contains(cfg.AlertSendTimeRange.NoticeAlert.Type, alertType) {
		timeRange = cfg.AlertSendTimeRange.NoticeAlert.SendTimeRange
	} else if contains(cfg.AlertSendTimeRange.ResourceAlert.Type, alertType) {
		timeRange = cfg.AlertSendTimeRange.ResourceAlert.SendTimeRange
	} else {
		global.LOG.Warnf("Alert type not found in sendTimeRange: %s", alertType)
		return false
	}

	if !isWithinTimeRange(timeRange) {
		return false
	}
	return true
}

func contains(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

func isWithinTimeRange(savedTimeString string) bool {
	now := time.Now()
	timeParts := strings.Split(savedTimeString, " - ")
	if len(timeParts) != 2 {
		global.LOG.Info("Time range string format error, should be: 'HH:MM:SS - HH:MM:SS'")
		return false
	}
	startTime, err1 := time.Parse("15:04:05", strings.TrimSpace(timeParts[0]))
	endTime, err2 := time.Parse("15:04:05", strings.TrimSpace(timeParts[1]))
	if err1 != nil || err2 != nil {
		global.LOG.Infof("Invalid time format in range: %s, errors: %v, %v", savedTimeString, err1, err2)
		return false
	}

	skipTime := time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, now.Location())
	endSkipTime := time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, now.Location())

	if endSkipTime.Before(skipTime) {
		return now.After(skipTime) || now.Before(endSkipTime)
	}
	return now.After(skipTime) && now.Before(endSkipTime)
}

func GetEmailContent(alertType string, params []dto.Param) string {
	switch GetCronJobType(alertType) {
	case "ssl":
		return i18n.GetMsgWithMap("SSLAlert", map[string]interface{}{"num": getValueByIndex(params, "1"), "day": getValueByIndex(params, "2")})
	case "siteEndTime":
		return i18n.GetMsgWithMap("WebSiteAlert", map[string]interface{}{"num": getValueByIndex(params, "1"), "day": getValueByIndex(params, "2")})
	case "panelPwdEndTime":
		return i18n.GetMsgWithMap("PanelPwdExpirationAlert", map[string]interface{}{"day": getValueByIndex(params, "1")})
	case "licenseTime":
		return i18n.GetMsgWithMap("LicenseExpirationAlert", map[string]interface{}{"day": getValueByIndex(params, "1")})
	case "panelUpdate":
		return i18n.GetMsgByKey("PanelVersionAlert")
	case "cpu":
		return i18n.GetMsgWithMap("ResourceAlert", map[string]interface{}{"time": getValueByIndex(params, "1"), "name": getValueByIndex(params, "2"), "used": getValueByIndex(params, "3")})
	case "memory":
		return i18n.GetMsgWithMap("ResourceAlert", map[string]interface{}{"time": getValueByIndex(params, "1"), "name": getValueByIndex(params, "2"), "used": getValueByIndex(params, "3")})
	case "load":
		return i18n.GetMsgWithMap("ResourceAlert", map[string]interface{}{"time": getValueByIndex(params, "1"), "name": getValueByIndex(params, "2"), "used": getValueByIndex(params, "3")})
	case "disk":
		return i18n.GetMsgWithMap("DiskUsedAlert", map[string]interface{}{"name": getValueByIndex(params, "1"), "used": getValueByIndex(params, "2")})
	case "cronJob":
		return i18n.GetMsgWithMap("CronJobFailedAlert", map[string]interface{}{"name": getValueByIndex(params, "1")})
	case "clams":
		return i18n.GetMsgWithMap("ClamAlert", map[string]interface{}{"num": getValueByIndex(params, "1")})
	case "panelLogin":
		return i18n.GetMsgWithMap("SSHAndPanelLoginAlert", map[string]interface{}{"name": getValueByIndex(params, "1"), "ip": getValueByIndex(params, "2")})
	case "sshLogin":
		return i18n.GetMsgWithMap("SSHAndPanelLoginAlert", map[string]interface{}{"name": getValueByIndex(params, "1"), "ip": getValueByIndex(params, "2")})
	case "panelIpLogin":
		return i18n.GetMsgWithMap("SSHAndPanelLoginAlert", map[string]interface{}{"name": getValueByIndex(params, "1"), "ip": getValueByIndex(params, "2")})
	case "sshIpLogin":
		return i18n.GetMsgWithMap("SSHAndPanelLoginAlert", map[string]interface{}{"name": getValueByIndex(params, "1"), "ip": getValueByIndex(params, "2")})
	case "nodeException":
		return i18n.GetMsgWithMap("NodeExceptionAlert", map[string]interface{}{"num": getValueByIndex(params, "1")})
	case "licenseException":
		return i18n.GetMsgWithMap("LicenseExceptionAlert", map[string]interface{}{"num": getValueByIndex(params, "1")})
	default:
		return ""
	}
}

func getValueByIndex(params []dto.Param, index string) string {
	for _, p := range params {
		if p.Index == index {
			return p.Value
		}
	}
	return ""
}

func CountRecentFailedLoginLogs(minutes uint, failCount uint) (int, bool, error) {
	now := time.Now()
	startTime := now.Add(-time.Duration(minutes) * time.Minute)
	db := global.CoreDB.Model(&model.LoginLog{})
	var count int64
	err := db.Where("created_at >= ? AND status = ?", startTime, constant.StatusFailed).
		Count(&count).Error
	if err != nil {
		return 0, false, err
	}
	return int(count), int(count) >= int(failCount), nil
}

func FindRecentSuccessLoginsNotInWhitelist(minutes int, whitelist []string) ([]model.LoginLog, error) {
	now := time.Now()
	startTime := now.Add(-time.Duration(minutes) * time.Minute)

	whitelistMap := make(map[string]struct{})
	for _, ip := range whitelist {
		whitelistMap[ip] = struct{}{}
	}

	var logs []model.LoginLog
	err := global.CoreDB.Model(&model.LoginLog{}).
		Where("created_at >= ? AND status = ?", startTime, constant.StatusSuccess).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	var abnormalLogs []model.LoginLog
	for _, log := range logs {
		if _, ok := whitelistMap[log.IP]; !ok {
			abnormalLogs = append(abnormalLogs, log)
		}
	}
	return abnormalLogs, nil
}

func CountRecentFailedSSHLog(minutes uint, maxAllowed uint) (int, bool, error) {
	lines, err := grepSSHLog([]string{"Failed password", "Invalid user", "authentication failure"})
	if err != nil {
		return 0, false, err
	}

	thresholdTime := time.Now().Add(-time.Duration(minutes) * time.Minute)
	count := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		t, err := parseLogTime(line)
		if err != nil {
			continue
		}
		if t.After(thresholdTime) {
			count++
		}
	}
	return count, count >= int(maxAllowed), nil
}

func FindRecentSuccessLoginNotInWhitelist(minutes int, whitelist []string) ([]string, error) {
	lines, err := grepSSHLog([]string{"Accepted password", "Accepted publickey"})
	if err != nil {
		return nil, err
	}

	thresholdTime := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var abnormalLogins []string

	whitelistMap := make(map[string]struct{}, len(whitelist))
	for _, ip := range whitelist {
		whitelistMap[ip] = struct{}{}
	}

	ipRegex := regexp.MustCompile(`from\s+([0-9.]+)\s+port\s+(\d+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		t, err := parseLogTime(line)
		if err != nil || t.Before(thresholdTime) {
			continue
		}

		match := ipRegex.FindStringSubmatch(line)
		if len(match) >= 2 {
			ip := match[1]
			if _, ok := whitelistMap[ip]; !ok {
				abnormalLogins = append(abnormalLogins, fmt.Sprintf("%s-%s", ip, t.Format("2006-01-02 15:04:05")))
			}
		}
	}

	return abnormalLogins, nil
}

func findGrepPath() (string, error) {
	path, err := exec.LookPath("grep")
	if err != nil {
		return "", fmt.Errorf("grep not found in PATH: %w", err)
	}
	return path, nil
}

func grepSSHLog(keywords []string) ([]string, error) {
	logFiles := []string{"/var/log/secure", "/var/log/auth.log"}
	var results []string
	seen := make(map[string]struct{})

	grepPath, err := findGrepPath()
	if err != nil {
		return nil, fmt.Errorf("find grep failed: %w", err)
	}

	for _, logFile := range logFiles {
		if _, err := os.Stat(logFile); err != nil {
			continue
		}
		for _, keyword := range keywords {
			cmd := exec.Command(grepPath, "-a", keyword, logFile)
			output, err := cmd.Output()
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					if exitErr.ExitCode() == 1 {
						continue
					}
				}
				return nil, fmt.Errorf("read log file fail [%s]: %w", logFile, err)
			}

			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					if _, exists := seen[line]; !exists {
						results = append(results, line)
						seen[line] = struct{}{}
					}
				}
			}
		}
	}

	return results, nil
}

func parseLogTime(line string) (time.Time, error) {
	if len(line) < 15 {
		return time.Time{}, nil
	}
	timeStr := line[:15]
	parsedTime, err := time.ParseInLocation("Jan 2 15:04:05", timeStr, time.Local)
	if err != nil {
		return time.Time{}, nil
	}
	return parsedTime.AddDate(time.Now().Year(), 0, 0), nil
}
