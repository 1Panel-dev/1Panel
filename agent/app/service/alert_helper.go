package service

import (
	"encoding/json"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	alertUtil "github.com/1Panel-dev/1Panel/agent/utils/alert"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	versionUtil "github.com/1Panel-dev/1Panel/agent/utils/version"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ResourceAlertInterval = 30
	CheckIntervalSec      = 3
	LoadCheckIntervalMin  = 5
)

type AlertTaskHelper struct {
	DiskIO chan []disk.IOCountersStat
	NetIO  chan []net.IOCountersStat
}

type IAlertTaskHelper interface {
	StopTask()
	StartTask()
	ResetTask()
	InitTask(alertType string)
}

var cpuLoad1, cpuLoad5, cpuLoad15 []float64
var memoryLoad1, memoryLoad5, memoryLoad15 []float64

var baseTypes = map[string]bool{"ssl": true, "siteEndTime": true, "panelPwdEndTime": true, "panelUpdate": true}
var resourceTypes = map[string]bool{"cpu": true, "memory": true, "disk": true, "load": true, "panelLogin": true, "sshLogin": true, "nodeException": true, "licenseException": true}

func NewIAlertTaskHelper() IAlertTaskHelper {
	return &AlertTaskHelper{
		DiskIO: make(chan []disk.IOCountersStat, 1),
		NetIO:  make(chan []net.IOCountersStat, 1),
	}
}
func (m *AlertTaskHelper) StartTask() {
	baseAlert, resourceAlert := m.getClassifiedAlerts()
	if len(baseAlert) == 0 && len(resourceAlert) == 0 {
		return
	}
	handleBaseAlerts(baseAlert)
	handleResourceAlerts(resourceAlert)
}

func (m *AlertTaskHelper) StopTask() {
	stopBaseJob()
	stopResourceJob()
}

func (m *AlertTaskHelper) ResetTask() {
	m.StopTask()
	m.StartTask()
}

func (m *AlertTaskHelper) InitTask(alertType string) {
	resetAlertState(alertType)
	if baseTypes[alertType] {
		stopBaseJob()
	} else if resourceTypes[alertType] {
		stopResourceJob()
	}
	m.StartTask()
}

func resetAlertState(alertType string) {
	switch alertType {
	case "cpu":
		cpuLoad1 = []float64{}
		cpuLoad5 = []float64{}
		cpuLoad15 = []float64{}
	case "memory":
		memoryLoad1 = []float64{}
		memoryLoad5 = []float64{}
		memoryLoad15 = []float64{}
	}
}

func (m *AlertTaskHelper) getClassifiedAlerts() (baseAlerts, resourceAlerts []dto.AlertDTO) {
	alertList, _ := NewIAlertService().GetAlerts()
	for _, alert := range alertList {
		if baseTypes[alert.Type] {
			baseAlerts = append(baseAlerts, alert)
		} else if resourceTypes[alert.Type] {
			resourceAlerts = append(resourceAlerts, alert)
		}
	}
	return
}

func handleBaseAlerts(baseAlerts []dto.AlertDTO) {
	if len(baseAlerts) == 0 {
		stopResourceJob()
		return
	}
	if global.AlertBaseJobID == 0 {
		baseTask(baseAlerts)
		jobID, err := global.Cron.AddFunc("*/30 * * * *", func() {
			baseTask(baseAlerts)
		})
		if err != nil {
			global.LOG.Errorf("alert base job start failed: %v", err)
			return
		}
		global.AlertBaseJobID = jobID
		global.LOG.Info("start alert base job")
	}
}

func handleResourceAlerts(resourceAlerts []dto.AlertDTO) {
	if len(resourceAlerts) == 0 {
		stopResourceJob()
		return
	}
	if global.AlertResourceJobID == 0 {
		jobID, err := global.Cron.AddFunc("*/1 * * * *", func() {
			resourceTask(resourceAlerts)
		})
		if err != nil {
			global.LOG.Errorf("alert resource job start failed: %v", err)
			return
		}
		global.AlertResourceJobID = jobID
		global.LOG.Info("start alert resource job")
	}
}

func stopBaseJob() {
	if global.AlertBaseJobID != 0 {
		global.Cron.Remove(global.AlertBaseJobID)
		global.AlertBaseJobID = 0
		global.LOG.Info("stop alert base job")
	}
}

func stopResourceJob() {
	if global.AlertResourceJobID != 0 {
		global.Cron.Remove(global.AlertResourceJobID)
		global.AlertResourceJobID = 0
		global.LOG.Info("stop alert resource job")
	}
}

func baseTask(baseAlert []dto.AlertDTO) {
	for _, alert := range baseAlert {
		if !alertUtil.CheckSendTimeRange(alert.Type) {
			continue
		}
		switch alert.Type {
		case "ssl":
			loadSSLInfo(alert)
		case "siteEndTime":
			loadWebsiteInfo(alert)
		case "panelPwdEndTime":
			if global.IsMaster {
				loadPanelPwd(alert)
			}
		case "panelUpdate":
			if global.IsMaster {
				loadPanelUpdate(alert)
			}
		}
	}
}

func resourceTask(resourceAlert []dto.AlertDTO) {
	minute := time.Now().Minute()
	for _, alert := range resourceAlert {
		if !alertUtil.CheckSendTimeRange(alert.Type) {
			continue
		}
		execute := minute%LoadCheckIntervalMin == 0
		switch alert.Type {
		case "cpu":
			loadCPUUsage(alert)
		case "memory":
			loadMemUsage(alert)
		case "load":
			loadLoadInfo(alert)
		case "disk":
			loadDiskUsage(alert)
		case "panelLogin":
			loadPanelLogin(alert)
		case "sshLogin":
			loadSSHLogin(alert)
		case "nodeException":
			if execute && global.IsMaster {
				loadNodeException(alert)
			}
		case "licenseException":
			if execute && global.IsMaster {
				loadLicenseException(alert)
			}
		}
	}
}

func loadSSLInfo(alert dto.AlertDTO) {
	opts := getRepoOptionsByProject(alert.Project)
	sslList, _ := repo.NewISSLRepo().List(opts...)
	if len(sslList) == 0 {
		return
	}
	daysDiffMap, projectMap := calculateSSLExpiryDays(sslList, alert.Cycle)
	projectJSON := serializeAndSortProjects(projectMap)
	if projectJSON == "" || len(daysDiffMap) == 0 {
		return
	}
	sender := NewAlertSender(alert, projectJSON)
	for daysDiff, domains := range daysDiffMap {
		domainStr := strings.Join(domains, ",")
		params := createAlertBaseParams(strconv.Itoa(len(domains)), strconv.Itoa(daysDiff))
		sender.Send(domainStr, params)
	}
}

func loadWebsiteInfo(alert dto.AlertDTO) {
	opts := getRepoOptionsByProject(alert.Project)
	websiteList, _ := websiteRepo.List(opts...)
	if len(websiteList) == 0 {
		return
	}

	daysDiffMap, projectMap := calculateWebsiteExpiryDays(websiteList, alert.Cycle)
	projectJSON := serializeAndSortProjects(projectMap)
	if projectJSON == "" || len(daysDiffMap) == 0 {
		return
	}
	sender := NewAlertSender(alert, projectJSON)
	for daysDiff, domains := range daysDiffMap {
		domainStr := strings.Join(domains, ",")
		params := createAlertBaseParams(strconv.Itoa(len(domains)), strconv.Itoa(daysDiff))
		sender.Send(domainStr, params)
	}
}

func loadPanelPwd(alert dto.AlertDTO) {
	// only master alert
	expDays, err := getSettingValue("ExpirationDays")
	if err != nil || expDays == "0" {
		global.LOG.Info("panel password expiration setting not enabled, skip")
		return
	}

	expTimeStr, err := getSettingValue("ExpirationTime")
	if err != nil {
		return
	}
	expTime, _ := time.Parse(constant.DateTimeLayout, expTimeStr)
	daysDiff := calculateDaysDifference(expTime)
	if daysDiff >= 0 && int(alert.Cycle) >= daysDiff {
		params := createAlertPwdParams(strconv.Itoa(daysDiff))
		sender := NewAlertSender(alert, expTimeStr)
		sender.Send(strconv.Itoa(daysDiff), params)
	}
}

func loadPanelUpdate(alert dto.AlertDTO) {
	// only master alert
	info, err := versionUtil.GetUpgradeVersionInfo()
	if err != nil {
		global.LOG.Errorf("error getting version info: %s", err)
		return
	}

	version := getValidVersion(info)
	if version == "" {
		return
	}

	sender := NewAlertSender(alert, version)
	sender.Send(version, []dto.Param{})
}

// 获取 CPU 使用率数据并发送到通道
func loadCPUUsage(alert dto.AlertDTO) {
	percent, err := cpu.Percent(time.Duration(CheckIntervalSec)*time.Second, false)
	if err != nil {
		global.LOG.Errorf("error getting cpu usage, err: %v", err)
		return
	}

	if len(percent) > 0 {
		var usageLoad *[]float64
		var threshold int

		switch alert.Cycle {
		case 1:
			usageLoad = &cpuLoad1
			threshold = 1
		case 5:
			usageLoad = &cpuLoad5
			threshold = 5
		case 15:
			usageLoad = &cpuLoad15
			threshold = 15
		}
		shouldSendResourceAlert(alert, percent[0], usageLoad, threshold)
	}
}

// 获取内存使用情况数据并发送到通道
func loadMemUsage(alert dto.AlertDTO) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		global.LOG.Errorf("error getting memory usage, err: %v", err)
		return
	}

	percent := memStat.UsedPercent
	var memoryLoad *[]float64
	var threshold int

	switch alert.Cycle {
	case 1:
		memoryLoad = &memoryLoad1
		threshold = 1
	case 5:
		memoryLoad = &memoryLoad5
		threshold = 5
	case 15:
		memoryLoad = &memoryLoad15
		threshold = 15
	}
	shouldSendResourceAlert(alert, percent, memoryLoad, threshold)
}

// 获取系统负载数据并发送到通道
func loadLoadInfo(alert dto.AlertDTO) {
	avgStat, err := load.Avg()
	if err != nil {
		global.LOG.Errorf("error getting load usage, err: %v", err)
		return
	}
	var loadValue float64
	CPUTotal, _ := cpu.Counts(true)
	switch alert.Cycle {
	case 1:
		loadValue = avgStat.Load1 / (float64(CPUTotal*2) * 0.75) * 100
	case 5:
		loadValue = avgStat.Load5 / (float64(CPUTotal*2) * 0.75) * 100
	case 15:
		loadValue = avgStat.Load15 / (float64(CPUTotal*2) * 0.75) * 100
	default:
		return
	}
	if loadValue < float64(alert.Count) {
		return
	}
	newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
	if err != nil {
		global.LOG.Errorf("task log record not found, err: %v", err)
	}
	if isAlertDue(newDate) {
		sendResourceAlert(alert, loadValue)
	}
}

func loadDiskUsage(alert dto.AlertDTO) {
	newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
	if err != nil {
		global.LOG.Errorf("record not found, err: %v", err)
		return
	}
	if isAlertDue(newDate) {
		if strings.Contains(alert.Project, "all") {
			err = processAllDisks(alert)
		} else {
			err = processSingleDisk(alert)
		}
	}
}

func loadPanelLogin(alert dto.AlertDTO) {
	count, isAlert, err := alertUtil.CountRecentFailedLoginLogs(alert.Cycle, alert.Count)
	alertType := alert.Type
	quota := strconv.Itoa(count)
	quotaType := strconv.Itoa(int(alert.Cycle))
	if err != nil {
		global.LOG.Errorf("Failed to count recent failed login logs: %v", err)
	}
	if isAlert {
		alertType = "panelLogin"
		quota = strconv.Itoa(count)
		quotaType = "panelLogin"
		params := []dto.Param{
			{
				Index: "1",
				Key:   "cycle",
				Value: "",
			},
			{
				Index: "2",
				Key:   "project",
				Value: "",
			},
		}
		sendAlerts(alert, alertType, quota, quotaType, params)
	}

	whitelist := strings.Split(strings.TrimSpace(alert.AdvancedParams), "\n")
	records, err := alertUtil.FindRecentSuccessLoginsNotInWhitelist(30, whitelist)
	if err != nil {
		global.LOG.Errorf("Failed to check recent failed ip login logs: %v", err)
	}
	if len(records) > 0 {
		quota = strings.Join(func() []string {
			var ips []string
			for _, r := range records {
				ips = append(ips, r.IP)
			}
			return ips
		}(), "\n")
		alertType = "panelIpLogin"
		quotaType = "panelIpLogin"
		params := []dto.Param{
			{
				Index: "1",
				Key:   "cycle",
				Value: "",
			},
			{
				Index: "2",
				Key:   "project",
				Value: " IP ",
			},
		}
		sendAlerts(alert, alertType, quota, quotaType, params)
	}
}

func loadSSHLogin(alert dto.AlertDTO) {
	count, isAlert, err := alertUtil.CountRecentFailedSSHLog(alert.Cycle, alert.Count)
	alertType := alert.Type
	quota := strconv.Itoa(count)
	quotaType := strconv.Itoa(int(alert.Cycle))
	if err != nil {
		global.LOG.Errorf("Failed to count recent failed ssh login logs: %v", err)
	}
	if isAlert {
		alertType = "sshLogin"
		quota = strconv.Itoa(count)
		quotaType = "sshLogin"
		params := []dto.Param{
			{
				Index: "1",
				Key:   "cycle",
				Value: " SSH ",
			},
			{
				Index: "2",
				Key:   "project",
				Value: "",
			},
		}
		sendAlerts(alert, alertType, quota, quotaType, params)
	}
	whitelist := strings.Split(strings.TrimSpace(alert.AdvancedParams), "\n")
	records, err := alertUtil.FindRecentSuccessLoginNotInWhitelist(30, whitelist)
	if err != nil {
		global.LOG.Errorf("Failed to check recent failed ip ssh login logs: %v", err)
	}
	if len(records) > 0 {
		quota = strings.Join(records, "\n")
		alertType = "sshIpLogin"
		quotaType = "sshIpLogin"
		params := []dto.Param{
			{
				Index: "1",
				Key:   "cycle",
				Value: " SSH ",
			},
			{
				Index: "2",
				Key:   "project",
				Value: " IP ",
			},
		}
		sendAlerts(alert, alertType, quota, quotaType, params)
	}
}

func loadNodeException(alert dto.AlertDTO) {
	// only master alert
	failCount, err := xpack.GetNodeErrorAlert()
	if err != nil {
		global.LOG.Errorf("error getting node, err: %s", err)
		return
	}
	if failCount > 0 {
		quotaType := "node-error"
		params := []dto.Param{
			{
				Index: "1",
				Key:   "cycle",
				Value: strconv.Itoa(int(failCount)),
			},
		}
		newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
		if err != nil {
			global.LOG.Errorf("record not found, err: %v", err)
			return
		}
		if isAlertDue(newDate) {
			sender := NewAlertSender(alert, quotaType)
			sender.ResourceSend(strconv.Itoa(int(failCount)), params)
		}
	}

}

func loadLicenseException(alert dto.AlertDTO) {
	// only master alert
	failCount, err := xpack.GetLicenseErrorAlert()
	if err != nil {
		global.LOG.Errorf("error getting license, err: %s", err)
		return
	}
	if failCount > 0 {
		quotaType := "license-error"
		params := []dto.Param{
			{
				Index: "1",
				Key:   "cycle",
				Value: strconv.Itoa(int(failCount)),
			},
		}
		newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
		if err != nil {
			global.LOG.Errorf("record not found, err: %v", err)
			return
		}
		if isAlertDue(newDate) {
			sender := NewAlertSender(alert, quotaType)
			sender.ResourceSend(strconv.Itoa(int(failCount)), params)
		}
	}
}

func sendAlerts(alert dto.AlertDTO, alertType, quota, quotaType string, params []dto.Param) {
	methods := strings.Split(alert.Method, ",")
	newDate, err := alertRepo.GetTaskLog(alertType, alert.ID)
	if err != nil {
		global.LOG.Errorf("task log record not found, err: %v", err)
	}
	if newDate.IsZero() || calculateMinutesDifference(newDate) > ResourceAlertInterval {
		for _, m := range methods {
			m = strings.TrimSpace(m)
			switch m {
			case constant.SMS:
				if !alertUtil.CheckSMSSendLimit(constant.SMS) {
					continue
				}
				todayCount, isValid := canSendAlertToday(alertType, quotaType, alert.SendCount, constant.SMS)
				if !isValid {
					continue
				}
				create := dto.AlertLogCreate{
					Type:    alertType,
					AlertId: alert.ID,
					Count:   todayCount + 1,
				}
				_ = xpack.CreateSMSAlertLog(alertType, alert, create, quotaType, params, constant.SMS)
				alertUtil.CreateNewAlertTask(quota, alertType, quotaType, constant.SMS)
				global.LOG.Infof("%s alert sms push successful", alertType)

			case constant.Email:
				todayCount, isValid := canSendAlertToday(alertType, quotaType, alert.SendCount, constant.Email)
				if !isValid {
					continue
				}
				create := dto.AlertLogCreate{
					Type:    alertType,
					AlertId: alert.ID,
					Count:   todayCount + 1,
				}
				alertInfo := alert
				alertInfo.Type = alertType
				create.AlertRule = alertUtil.ProcessAlertRule(alert)
				create.AlertDetail = alertUtil.ProcessAlertDetail(alertInfo, quotaType, params, constant.Email)
				transport := xpack.LoadRequestTransport()
				_ = alertUtil.CreateEmailAlertLog(create, alertInfo, params, transport)
				alertUtil.CreateNewAlertTask(quota, alertType, quotaType, constant.Email)
				global.LOG.Infof("%s alert email push successful", alertType)
			}
		}
	}
}

// ------------------------------
func getRepoOptionsByProject(project string) []repo.DBOption {
	var opts []repo.DBOption
	if project != "all" {
		itemID, _ := strconv.Atoi(project)
		opts = append(opts, repo.WithByID(uint(itemID)))
	}
	return opts
}

func serializeAndSortProjects(projectMap map[uint][]time.Time) string {
	if len(projectMap) == 0 {
		return ""
	}
	keys := make([]int, 0, len(projectMap))
	for k := range projectMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	projectJSON, err := json.Marshal(projectMap)
	if err != nil {
		global.LOG.Errorf("Failed to serialize projectMap: %v", err)
		return ""
	}

	return string(projectJSON)
}

func calculateSSLExpiryDays(sslList []model.WebsiteSSL, cycle uint) (map[int][]string, map[uint][]time.Time) {
	currentDate := time.Now()
	daysDiffMap := make(map[int][]string)
	projectMap := make(map[uint][]time.Time)

	for _, ssl := range sslList {
		daysDiff := int(ssl.ExpireDate.Sub(currentDate).Hours() / 24)
		if daysDiff > 0 && int(cycle) >= daysDiff {
			daysDiffMap[daysDiff] = append(daysDiffMap[daysDiff], ssl.PrimaryDomain)
			projectMap[ssl.ID] = append(projectMap[ssl.ID], ssl.ExpireDate)
		}
	}
	return daysDiffMap, projectMap
}

func calculateWebsiteExpiryDays(websites []model.Website, cycle uint) (map[int][]string, map[uint][]time.Time) {
	currentDate := time.Now()
	daysDiffMap := make(map[int][]string)
	projectMap := make(map[uint][]time.Time)

	for _, website := range websites {
		daysDiff := int(website.ExpireDate.Sub(currentDate).Hours() / 24)
		if daysDiff > 0 && int(cycle) >= daysDiff {
			daysDiffMap[daysDiff] = append(daysDiffMap[daysDiff], website.PrimaryDomain)
			projectMap[website.ID] = append(projectMap[website.ID], website.ExpireDate)
		}
	}
	return daysDiffMap, projectMap
}

func getSettingValue(key string) (string, error) {
	var setting model.Setting
	if err := global.CoreDB.Model(&model.Setting{}).Where("key = ?", key).First(&setting).Error; err != nil {
		global.LOG.Errorf("load %s from db setting failed: %v", key, err)
		return "", err
	}
	return setting.Value, nil
}

func getValidVersion(info *dto.UpgradeInfo) string {
	if info.NewVersion != "" {
		return info.NewVersion
	} else if info.TestVersion != "" {
		return info.TestVersion
	} else if info.LatestVersion != "" {
		return info.LatestVersion
	}
	return ""
}

func shouldSendResourceAlert(alert dto.AlertDTO, currentUsage float64, usageLoad *[]float64, threshold int) {
	newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
	if err != nil {
		global.LOG.Errorf("record not found, err: %v", err)
	}
	if isAlertDue(newDate) {
		*usageLoad = append(*usageLoad, currentUsage)
		if len(*usageLoad) > threshold {
			*usageLoad = (*usageLoad)[1:]
		}
		if len(*usageLoad) == threshold {
			avgUsage := average(*usageLoad)
			if avgUsage >= float64(alert.Count) {
				sendResourceAlert(alert, avgUsage)
			}
		}
	}
}

func isAlertDue(lastAlertTime time.Time) bool {
	if lastAlertTime.IsZero() {
		return true
	}
	return calculateMinutesDifference(lastAlertTime) > ResourceAlertInterval
}

func sendResourceAlert(alert dto.AlertDTO, value float64) {
	valueStr := common.FormatPercent(value)
	module := getModuleName(alert.Type)
	params := createAlertAvgParams(strconv.Itoa(int(alert.Cycle)), module, valueStr)
	sender := NewAlertSender(alert, strconv.Itoa(int(alert.Cycle)))
	sender.ResourceSend(valueStr, params)
}

func getModuleName(alertType string) string {
	var module string
	switch alertType {
	case "cpu":
		module = " CPU "
	case "memory":
		module = "内存"
	case "load":
		module = "负载"
	default:
	}
	return module
}

func canSendAlertToday(alertType, quotaType string, sendCount uint, method string) (uint, bool) {
	todayCount, _, err := alertRepo.LoadTaskCount(alertType, quotaType, method)
	if err != nil {
		global.LOG.Errorf("error getting task info, err: %v", err)
		return todayCount, false
	}
	if todayCount >= sendCount {
		return todayCount, false
	}

	return todayCount, true
}

func average(arr []float64) float64 {
	total := 0.0
	for _, v := range arr {
		total += v
	}
	return total / float64(len(arr))
}

func createAlertBaseParams(project, cycle string) []dto.Param {
	return []dto.Param{
		{
			Index: "1",
			Key:   "project",
			Value: project,
		},
		{
			Index: "2",
			Key:   "cycle",
			Value: cycle,
		},
	}
}

func createAlertPwdParams(cycle string) []dto.Param {
	return []dto.Param{
		{
			Index: "1",
			Key:   "cycle",
			Value: cycle,
		},
	}
}

func createAlertAvgParams(cycle, module, count string) []dto.Param {
	return []dto.Param{
		{
			Index: "1",
			Key:   "cycle",
			Value: cycle,
		},
		{
			Index: "2",
			Key:   "module",
			Value: module,
		},
		{
			Index: "3",
			Key:   "count",
			Value: count,
		},
	}
}

func createAlertDiskParams(project, count string) []dto.Param {
	return []dto.Param{
		{
			Index: "1",
			Key:   "project",
			Value: project,
		},
		{
			Index: "2",
			Key:   "count",
			Value: count,
		},
	}
}

func processAllDisks(alert dto.AlertDTO) error {
	diskList, err := NewIAlertService().GetDisks()
	if err != nil {
		global.LOG.Errorf("error getting disk list, err: %v", err)
		return err
	}
	for _, item := range diskList {
		if success, err := checkAndCreateDiskAlert(alert, item.Path); err == nil && success {
			global.LOG.Infof("disk alert pushed successfully for %s", item.Path)
		}
	}
	return nil
}

func processSingleDisk(alert dto.AlertDTO) error {
	success, err := checkAndCreateDiskAlert(alert, alert.Project)
	if err != nil {
		return err
	}
	if success {
		global.LOG.Infof("disk alert pushed successfully for %s", alert.Project)
	}
	return nil
}

func checkAndCreateDiskAlert(alert dto.AlertDTO, path string) (bool, error) {
	usageStat, err := disk.Usage(path)
	if err != nil {
		global.LOG.Errorf("error getting disk usage for %s, err: %v", path, err)
		return false, err
	}

	usedTotal, usedStr := calculateUsedTotal(alert.Cycle, usageStat)
	commonTotal := float64(alert.Count)
	if alert.Cycle == 1 {
		commonTotal *= 1024 * 1024 * 1024
	}
	if usedTotal < commonTotal {
		return false, nil
	}
	global.LOG.Infof("disk「 %s 」usage: %s", path, usedStr)
	params := createAlertDiskParams(path, usedStr)
	sender := NewAlertSender(alert, alert.Project)
	sender.ResourceSend(path, params)
	return true, nil
}

func calculateUsedTotal(cycle uint, usageStat *disk.UsageStat) (float64, string) {
	if cycle == 1 {
		return float64(usageStat.Used), common.FormatBytes(usageStat.Used)
	}
	return usageStat.UsedPercent, common.FormatPercent(usageStat.UsedPercent)
}

func calculateDaysDifference(expirationTime time.Time) int {
	currentDate := time.Now()
	formattedTime := currentDate.Format(constant.DateTimeLayout)
	parsedTime, _ := time.Parse(constant.DateTimeLayout, formattedTime)
	timeGap := expirationTime.Sub(parsedTime).Milliseconds()
	if timeGap < 0 {
		return -1
	}
	daysDifference := int(math.Floor(float64(timeGap) / (3600 * 1000 * 24)))
	return daysDifference
}

func calculateMinutesDifference(newDate time.Time) int {
	now := time.Now()
	if newDate.After(now) {
		return -1
	}
	minutesDifference := int(now.Sub(newDate).Minutes())
	return minutesDifference
}
