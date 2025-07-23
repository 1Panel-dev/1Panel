package service

import (
	"encoding/json"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	versionUtil "github.com/1Panel-dev/1Panel/agent/utils/version"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	alertUtil "github.com/1Panel-dev/1Panel/agent/utils/alert"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
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

const ResourceAlertInterval = 30

var baseTypes = map[string]bool{"ssl": true, "siteEndTime": true, "panelPwdEndTime": true, "panelUpdate": true}
var resourceTypes = map[string]bool{"cpu": true, "memory": true, "disk": true, "load": true}

func NewIAlertTaskHelper() IAlertTaskHelper {
	return &AlertTaskHelper{}
}
func (m *AlertTaskHelper) StartTask() {
	baseAlert, resourceAlert := handleTask()
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
	if alertType == "cpu" {
		cpuLoad1 = []float64{}
		cpuLoad5 = []float64{}
		cpuLoad15 = []float64{}
	}
	if alertType == "memory" {
		memoryLoad1 = []float64{}
		memoryLoad5 = []float64{}
		memoryLoad15 = []float64{}
	}
	if baseTypes[alertType] {
		stopBaseJob()
	} else if resourceTypes[alertType] {
		stopResourceJob()
	}
	m.StartTask()
}

func resourceTask(resourceAlert []dto.AlertDTO) {
	for _, alert := range resourceAlert {
		if !alertUtil.CheckSendTimeRange(alert.Type) {
			continue
		}
		switch alert.Type {
		case "cpu":
			loadCPUUsage(alert)
		case "memory":
			loadMemUsage(alert)
		case "load":
			loadLoadInfo(alert)
		case "disk":
			loadDiskUsage(alert)
		default:
		}
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
		default:
		}
	}
}

func handleTask() (baseAlert []dto.AlertDTO, resourceAlert []dto.AlertDTO) {
	alertList, _ := NewIAlertService().GetAlerts()
	baseAlert, resourceAlert = classifyAlerts(alertList)
	return baseAlert, resourceAlert
}

func classifyAlerts(alertList []dto.AlertDTO) (baseAlert, resourceAlert []dto.AlertDTO) {
	for _, alert := range alertList {
		if baseTypes[alert.Type] {
			baseAlert = append(baseAlert, alert)
		} else if resourceTypes[alert.Type] {
			resourceAlert = append(resourceAlert, alert)
		}
	}
	return
}

func handleBaseAlerts(baseAlert []dto.AlertDTO) {
	if len(baseAlert) > 0 {
		if global.AlertBaseJobID == 0 {
			baseTask(baseAlert)
			jobID, err := global.Cron.AddFunc("*/30 * * * *", func() {
				baseTask(baseAlert)
			})
			if err != nil {
				global.LOG.Errorf("alert base job start failed: %v", err)
				return
			}
			global.AlertBaseJobID = jobID
			global.LOG.Info("start alert base job")
		}
	} else {
		stopBaseJob()
	}
}

func handleResourceAlerts(resourceAlert []dto.AlertDTO) {
	if len(resourceAlert) > 0 {
		if global.AlertResourceJobID == 0 {
			jobID, err := global.Cron.AddFunc("*/1 * * * *", func() {
				resourceTask(resourceAlert)
			})
			if err != nil {
				global.LOG.Errorf("alert resource job start failed: %v", err)
				return
			}
			global.AlertResourceJobID = jobID
			global.LOG.Info("start alert resource job")
		}
	} else {
		stopResourceJob()
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

func loadSSLInfo(alert dto.AlertDTO) {
	var opts []repo.DBOption
	if alert.Project != "all" {
		itemID, _ := strconv.Atoi(alert.Project)
		opts = append(opts, repo.WithByID(uint(itemID)))
	}

	sslList, _ := repo.NewISSLRepo().List(opts...)
	currentDate := time.Now()
	daysDifferenceMap := make(map[int][]string)
	projectMap := make(map[uint][]time.Time)
	for _, ssl := range sslList {
		daysDifference := int(ssl.ExpireDate.Sub(currentDate).Hours() / 24)
		if daysDifference > 0 && int(alert.Cycle) >= daysDifference {
			daysDifferenceMap[daysDifference] = append(daysDifferenceMap[daysDifference], ssl.PrimaryDomain)
			projectMap[ssl.ID] = append(projectMap[ssl.ID], ssl.ExpireDate)
		}
	}
	projectJSON := serializeAndSortProjects(projectMap)
	if projectJSON == "" {
		return
	}
	todayCount, totalCount, err := alertRepo.LoadTaskCount(alert.Type, projectJSON)
	if err != nil || todayCount >= 1 || alert.SendCount <= totalCount {
		return
	}
	if len(daysDifferenceMap) > 0 {
		create := dto.AlertLogCreate{
			Status:  constant.AlertSuccess,
			Count:   totalCount + 1,
			AlertId: alert.ID,
			Type:    alert.Type,
		}
		methods := strings.Split(alert.Method, ",")
		for _, m := range methods {
			m = strings.TrimSpace(m)
			for daysDifference, domain := range daysDifferenceMap {
				primaryDomain := strings.Join(domain, ",")
				var params []dto.Param
				params = createAlertBaseParams(strconv.Itoa(len(primaryDomain)), strconv.Itoa(daysDifference))
				switch m {
				case constant.SMS:
					if !alertUtil.CheckTaskFrequency(constant.SMS) {
						continue
					}
					_ = xpack.CreateSMSAlertLog(alert, create, primaryDomain, params, constant.SMS)
					alertUtil.CreateNewAlertTask(alert.Project, alert.Type, projectJSON, constant.SMS)
				case constant.Email:
					alertDetail := alertUtil.ProcessAlertDetail(alert, primaryDomain, params, constant.Email)
					alertRule := alertUtil.ProcessAlertRule(alert)
					create.AlertRule = alertRule
					create.AlertDetail = alertDetail
					transport := xpack.LoadRequestTransport()
					_ = alertUtil.CreateEmailAlertLog(create, alert, params, transport)
					alertUtil.CreateNewAlertTask(alert.Project, alert.Type, projectJSON, constant.Email)
				default:
				}
			}
		}
		global.LOG.Info("SSL alert push successful")
	}
}

func loadWebsiteInfo(alert dto.AlertDTO) {
	var opts []repo.DBOption
	if alert.Project != "all" {
		itemID, _ := strconv.Atoi(alert.Project)
		opts = append(opts, repo.WithByID(uint(itemID)))
	}

	websiteList, _ := websiteRepo.List(opts...)
	currentDate := time.Now()
	daysDifferenceMap := make(map[int][]string)
	projectMap := make(map[uint][]time.Time)
	for _, website := range websiteList {
		daysDifference := int(website.ExpireDate.Sub(currentDate).Hours() / 24)
		if daysDifference > 0 && int(alert.Cycle) >= daysDifference {
			daysDifferenceMap[daysDifference] = append(daysDifferenceMap[daysDifference], website.PrimaryDomain)
			projectMap[website.ID] = append(projectMap[website.ID], website.ExpireDate)
		}
	}
	projectJSON := serializeAndSortProjects(projectMap)
	if projectJSON == "" {
		return
	}
	todayCount, totalCount, err := alertRepo.LoadTaskCount(alert.Type, projectJSON)
	if err != nil || todayCount >= 1 || alert.SendCount <= totalCount {
		return
	}
	if len(daysDifferenceMap) > 0 {
		create := dto.AlertLogCreate{
			Status:  constant.AlertSuccess,
			Count:   totalCount + 1,
			AlertId: alert.ID,
			Type:    alert.Type,
		}
		methods := strings.Split(alert.Method, ",")
		for _, m := range methods {
			m = strings.TrimSpace(m)
			for daysDifference, websites := range daysDifferenceMap {
				primaryDomain := strings.Join(websites, ",")
				var params []dto.Param
				params = createAlertBaseParams(strconv.Itoa(len(websites)), strconv.Itoa(daysDifference))
				switch m {
				case constant.SMS:
					if !alertUtil.CheckTaskFrequency(constant.SMS) {
						continue
					}
					_ = xpack.CreateSMSAlertLog(alert, create, primaryDomain, params, constant.SMS)
					alertUtil.CreateNewAlertTask(alert.Project, alert.Type, projectJSON, constant.SMS)
				case constant.Email:
					alertDetail := alertUtil.ProcessAlertDetail(alert, primaryDomain, params, constant.Email)
					alertRule := alertUtil.ProcessAlertRule(alert)
					create.AlertDetail = alertDetail
					create.AlertRule = alertRule
					transport := xpack.LoadRequestTransport()
					_ = alertUtil.CreateEmailAlertLog(create, alert, params, transport)
					alertUtil.CreateNewAlertTask(alert.Project, alert.Type, projectJSON, constant.Email)
				default:
				}
			}
		}
		global.LOG.Info("website expiration alert push successful")
	}
}

func loadPanelPwd(alert dto.AlertDTO) {
	// only master alert
	var expirationDays model.Setting
	if err := global.CoreDB.Model(&model.Setting{}).Where("key = ?", "ExpirationDays").First(&expirationDays).Error; err != nil {
		global.LOG.Errorf("load %s from db setting failed, err: %v", "ExpirationDays", err)
		return
	}
	if expirationDays.Value == "0" {
		global.LOG.Info("panel password expiration setting not enabled, skip")
		return
	}
	var expirationTime model.Setting
	if err := global.CoreDB.Model(&model.Setting{}).Where("key = ?", "ExpirationTime").First(&expirationTime).Error; err != nil {
		global.LOG.Errorf("load %s from db setting failed, err: %v", "ExpirationTime", err)
		return
	}
	todayCount, totalCount, err := alertRepo.LoadTaskCount(alert.Type, expirationTime.Value)
	if err != nil || todayCount >= 1 || alert.SendCount <= totalCount {
		return
	}
	create := dto.AlertLogCreate{
		Count:   totalCount + 1,
		AlertId: alert.ID,
		Type:    alert.Type,
	}

	var params []dto.Param
	defaultDate, _ := time.Parse(constant.DateTimeLayout, expirationTime.Value)
	daysDifference := calculateDaysDifference(defaultDate)
	if daysDifference >= 0 && int(alert.Cycle) >= daysDifference {
		params = createAlertPwdParams(strconv.Itoa(daysDifference))
		methods := strings.Split(alert.Method, ",")
		for _, m := range methods {
			m = strings.TrimSpace(m)
			switch m {
			case constant.SMS:
				if !alertUtil.CheckTaskFrequency(constant.SMS) {
					continue
				}
				_ = xpack.CreateSMSAlertLog(alert, create, strconv.Itoa(daysDifference), params, constant.SMS)
				alertUtil.CreateNewAlertTask(expirationTime.Value, alert.Type, expirationTime.Value, constant.SMS)
			case constant.Email:
				alertDetail := alertUtil.ProcessAlertDetail(alert, strconv.Itoa(daysDifference), params, constant.Email)
				alertRule := alertUtil.ProcessAlertRule(alert)
				create.AlertRule = alertRule
				create.AlertDetail = alertDetail
				transport := xpack.LoadRequestTransport()
				_ = alertUtil.CreateEmailAlertLog(create, alert, params, transport)
				alertUtil.CreateNewAlertTask(expirationTime.Value, alert.Type, expirationTime.Value, constant.Email)
			default:
			}
		}
		global.LOG.Info("panel password expiration alert push successful")
	}
}

func loadPanelUpdate(alert dto.AlertDTO) {
	// only master alert
	info, err := versionUtil.GetUpgradeVersionInfo()
	if err != nil {
		global.LOG.Errorf("error getting version, err: %s", err)
		return
	}

	// 获取版本信息
	var version string
	// 检查哪个版本字段不为空，并赋值
	if info.NewVersion != "" {
		version = info.NewVersion
	} else if info.TestVersion != "" {
		version = info.TestVersion
	} else if info.LatestVersion != "" {
		version = info.LatestVersion
	}
	if version == "" {
		return
	}

	todayCount, totalCount, err := alertRepo.LoadTaskCount(alert.Type, version)
	if err != nil || todayCount >= 1 || alert.SendCount <= totalCount {
		return
	}
	var create = dto.AlertLogCreate{
		Type:    alert.Type,
		AlertId: alert.ID,
		Count:   1,
	}
	var params []dto.Param
	methods := strings.Split(alert.Method, ",")
	for _, m := range methods {
		m = strings.TrimSpace(m)
		switch m {
		case constant.SMS:
			if !alertUtil.CheckTaskFrequency(constant.SMS) {
				continue
			}
			_ = xpack.CreateSMSAlertLog(alert, create, version, params, constant.SMS)
			alertUtil.CreateNewAlertTask(version, alert.Type, version, constant.SMS)
		case constant.Email:
			alertDetail := alertUtil.ProcessAlertDetail(alert, version, params, constant.Email)
			alertRule := alertUtil.ProcessAlertRule(alert)
			create.AlertRule = alertRule
			create.AlertDetail = alertDetail
			transport := xpack.LoadRequestTransport()
			_ = alertUtil.CreateEmailAlertLog(create, alert, params, transport)
			alertUtil.CreateNewAlertTask(version, alert.Type, version, constant.Email)
		default:
		}
	}
	global.LOG.Info("panel update alert push successful")
}

// 获取 CPU 使用率数据并发送到通道
func loadCPUUsage(alert dto.AlertDTO) {
	percent, err := cpu.Percent(3*time.Second, false)
	if err != nil {
		global.LOG.Errorf("error getting cpu usage, err: %v", err)
		return
	}

	if len(percent) > 0 {
		var cpuLoad *[]float64
		var threshold int

		switch alert.Cycle {
		case 1:
			cpuLoad = &cpuLoad1
			threshold = 1
		case 5:
			cpuLoad = &cpuLoad5
			threshold = 5
		case 15:
			cpuLoad = &cpuLoad15
			threshold = 15
		default:
			return
		}

		if checkAndSendAlert(alert, percent[0], cpuLoad, threshold) {
			global.LOG.Info("cpu alert push successful")
		}
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
	default:
		return
	}
	if checkAndSendAlert(alert, percent, memoryLoad, threshold) {
		global.LOG.Info("memory alert push successful")
	}
}

// 获取系统负载数据并发送到通道
func loadLoadInfo(alert dto.AlertDTO) {
	todayCount, isValid := checkTaskFrequency(alert.Type, strconv.Itoa(int(alert.Cycle)), alert.SendCount)
	if isValid {
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
		newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
		if err != nil {
			global.LOG.Errorf("task log record not found, err: %v", err)
		}
		if newDate.IsZero() || calculateMinutesDifference(newDate) > ResourceAlertInterval {
			if loadValue >= float64(alert.Count) {
				global.LOG.Infof("%d minute load: %f,detail: %v", alert.Cycle, loadValue, avgStat)
				createAndLogAlert(alert, loadValue, todayCount)
				global.LOG.Info("load alert task push successful")
			}
		}
	}

}

// 内存/cpu检查是否需要发送告警并处理相关逻辑
func checkAndSendAlert(alert dto.AlertDTO, currentUsage float64, usageLoad *[]float64, threshold int) bool {
	todayCount, isValid := checkTaskFrequency(alert.Type, strconv.Itoa(int(alert.Cycle)), alert.SendCount)
	if !isValid {
		return false
	}
	newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
	if err != nil {
		global.LOG.Errorf("record not found, err: %v", err)
		return false
	}

	*usageLoad = append(*usageLoad, currentUsage)

	if len(*usageLoad) > threshold {
		*usageLoad = (*usageLoad)[1:]
	}

	if newDate.IsZero() || calculateMinutesDifference(newDate) > ResourceAlertInterval {
		if len(*usageLoad) == threshold {
			avgUsage := average(*usageLoad)
			if avgUsage >= float64(alert.Count) {
				global.LOG.Infof("%d minute %s: %f , usage: %v", threshold, alert.Type, avgUsage, usageLoad)
				createAndLogAlert(alert, avgUsage, todayCount)
				return true
			}
		}
	}
	return false
}

// 检查是否超过今日发送次数限制
func checkTaskFrequency(alertType, quotaType string, sendCount uint) (uint, bool) {
	todayCount, _, err := alertRepo.LoadTaskCount(alertType, quotaType)
	if err != nil {
		global.LOG.Errorf("error getting task info, err: %v", err)
		return todayCount, false
	}
	if todayCount >= sendCount {
		return todayCount, false
	}

	return todayCount, true
}

// 创建告警日志和详情
func createAndLogAlert(alert dto.AlertDTO, avgUsage float64, todayCount uint) {
	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   todayCount + 1,
		AlertId: alert.ID,
		Type:    alert.Type,
	}
	avgUsagePercent := common.FormatPercent(avgUsage)
	params := createAlertAvgParams(strconv.Itoa(int(alert.Cycle)), getModule(alert.Type), avgUsagePercent)
	methods := strings.Split(alert.Method, ",")
	for _, m := range methods {
		m = strings.TrimSpace(m)
		switch m {
		case constant.SMS:
			if !alertUtil.CheckTaskFrequency(constant.SMS) {
				continue
			}
			_ = xpack.CreateSMSAlertLog(alert, create, avgUsagePercent, params, constant.SMS)
			alertUtil.CreateNewAlertTask(avgUsagePercent, alert.Type, strconv.Itoa(int(alert.Cycle)), constant.SMS)
		case constant.Email:
			alertDetail := alertUtil.ProcessAlertDetail(alert, avgUsagePercent, params, constant.Email)
			alertRule := alertUtil.ProcessAlertRule(alert)
			create.AlertRule = alertRule
			create.AlertDetail = alertDetail
			transport := xpack.LoadRequestTransport()
			_ = alertUtil.CreateEmailAlertLog(create, alert, params, transport)
			alertUtil.CreateNewAlertTask(avgUsagePercent, alert.Type, strconv.Itoa(int(alert.Cycle)), constant.Email)
		default:
		}
	}
}

func getModule(alertType string) string {
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

func loadDiskUsage(alert dto.AlertDTO) {
	todayCount, isValid := checkTaskFrequency(alert.Type, alert.Project, alert.SendCount)
	if isValid {
		newDate, err := alertRepo.GetTaskLog(alert.Type, alert.ID)
		if err != nil {
			global.LOG.Errorf("record not found, err: %v", err)
		}

		if newDate.IsZero() || calculateMinutesDifference(newDate) > ResourceAlertInterval {
			if strings.Contains(alert.Project, "all") {
				err = processAllDisks(alert, todayCount)
			} else {
				err = processSingleDisk(alert, todayCount)
			}
			if err != nil {
				global.LOG.Errorf("error processing disk usage, err: %v", err)
			}
		}
	}

}

func processAllDisks(alert dto.AlertDTO, todayCount uint) error {
	diskList, err := NewIAlertService().GetDisks()
	if err != nil {
		global.LOG.Errorf("error getting disk list, err: %v", err)
		return err
	}

	var flag bool
	for _, item := range diskList {
		if success, err := checkAndCreateDiskAlert(alert, item.Path, todayCount); err == nil && success {
			flag = true
		}
	}
	if flag {
		global.LOG.Info("all disk alert push successful")
	}
	return nil
}

func processSingleDisk(alert dto.AlertDTO, todayCount uint) error {
	success, err := checkAndCreateDiskAlert(alert, alert.Project, todayCount)
	if err != nil {
		return err
	}
	if success {
		global.LOG.Info("disk alert push successful")
	}
	return nil
}

func checkAndCreateDiskAlert(alert dto.AlertDTO, path string, todayCount uint) (bool, error) {
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
	create := dto.AlertLogCreate{
		Status:  constant.AlertSuccess,
		Count:   todayCount + 1,
		AlertId: alert.ID,
		Type:    alert.Type,
	}
	var params []dto.Param
	params = createAlertDiskParams(path, usedStr)
	methods := strings.Split(alert.Method, ",")
	for _, m := range methods {
		m = strings.TrimSpace(m)
		switch m {
		case constant.SMS:
			if !alertUtil.CheckTaskFrequency(constant.SMS) {
				continue
			}
			_ = xpack.CreateSMSAlertLog(alert, create, path, params, constant.SMS)
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(alert.Cycle)), alert.Type, alert.Project, constant.SMS)
		case constant.Email:
			alertDetail := alertUtil.ProcessAlertDetail(alert, path, params, constant.Email)
			alertRule := alertUtil.ProcessAlertRule(alert)
			create.AlertRule = alertRule
			create.AlertDetail = alertDetail
			transport := xpack.LoadRequestTransport()
			_ = alertUtil.CreateEmailAlertLog(create, alert, params, transport)
			alertUtil.CreateNewAlertTask(strconv.Itoa(int(alert.Cycle)), alert.Type, alert.Project, constant.Email)
		default:
		}
	}

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

func serializeAndSortProjects(projectMap map[uint][]time.Time) string {
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
