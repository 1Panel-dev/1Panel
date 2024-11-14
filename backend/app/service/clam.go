package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
	"github.com/1Panel-dev/1Panel/backend/utils/xpack"
	"github.com/jinzhu/copier"
	"github.com/robfig/cron/v3"

	"github.com/pkg/errors"
)

const (
	clamServiceNameCentOs = "clamd@scan.service"
	clamServiceNameUbuntu = "clamav-daemon.service"
	freshClamService      = "clamav-freshclam.service"
	resultDir             = "clamav"
)

type ClamService struct {
	serviceName string
}

type IClamService interface {
	LoadBaseInfo() (dto.ClamBaseInfo, error)
	Operate(operate string) error
	SearchWithPage(search dto.SearchClamWithPage) (int64, interface{}, error)
	Create(req dto.ClamCreate) error
	Update(req dto.ClamUpdate) error
	UpdateStatus(id uint, status string) error
	Delete(req dto.ClamDelete) error
	HandleOnce(req dto.OperateByID) error
	LoadFile(req dto.ClamFileReq) (string, error)
	UpdateFile(req dto.UpdateByNameAndFile) error
	LoadRecords(req dto.ClamLogSearch) (int64, interface{}, error)
	CleanRecord(req dto.OperateByID) error

	LoadRecordLog(req dto.ClamLogReq) (string, error)
}

func NewIClamService() IClamService {
	return &ClamService{}
}

func (c *ClamService) LoadBaseInfo() (dto.ClamBaseInfo, error) {
	var baseInfo dto.ClamBaseInfo
	baseInfo.Version = "-"
	baseInfo.FreshVersion = "-"
	exist1, _ := systemctl.IsExist(clamServiceNameCentOs)
	if exist1 {
		c.serviceName = clamServiceNameCentOs
		baseInfo.IsExist = true
		baseInfo.IsActive, _ = systemctl.IsActive(clamServiceNameCentOs)
	}
	exist2, _ := systemctl.IsExist(clamServiceNameUbuntu)
	if exist2 {
		c.serviceName = clamServiceNameUbuntu
		baseInfo.IsExist = true
		baseInfo.IsActive, _ = systemctl.IsActive(clamServiceNameUbuntu)
	}
	freshExist, _ := systemctl.IsExist(freshClamService)
	if freshExist {
		baseInfo.FreshIsExist = true
		baseInfo.FreshIsActive, _ = systemctl.IsActive(freshClamService)
	}
	if !cmd.Which("clamdscan") {
		baseInfo.IsActive = false
	}

	if baseInfo.IsActive {
		version, err := cmd.Exec("clamdscan --version")
		if err == nil {
			if strings.Contains(version, "/") {
				baseInfo.Version = strings.TrimPrefix(strings.Split(version, "/")[0], "ClamAV ")
			} else {
				baseInfo.Version = strings.TrimPrefix(version, "ClamAV ")
			}
		}
	} else {
		_ = StopAllCronJob(false)
	}
	if baseInfo.FreshIsActive {
		version, err := cmd.Exec("freshclam --version")
		if err == nil {
			if strings.Contains(version, "/") {
				baseInfo.FreshVersion = strings.TrimPrefix(strings.Split(version, "/")[0], "ClamAV ")
			} else {
				baseInfo.FreshVersion = strings.TrimPrefix(version, "ClamAV ")
			}
		}
	}
	return baseInfo, nil
}

func (c *ClamService) Operate(operate string) error {
	switch operate {
	case "start", "restart", "stop":
		stdout, err := cmd.Execf("systemctl %s %s", operate, c.serviceName)
		if err != nil {
			return fmt.Errorf("%s the %s failed, err: %s", operate, c.serviceName, stdout)
		}
		return nil
	case "fresh-start", "fresh-restart", "fresh-stop":
		stdout, err := cmd.Execf("systemctl %s %s", strings.TrimPrefix(operate, "fresh-"), freshClamService)
		if err != nil {
			return fmt.Errorf("%s the %s failed, err: %s", operate, c.serviceName, stdout)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (c *ClamService) SearchWithPage(req dto.SearchClamWithPage) (int64, interface{}, error) {
	total, commands, err := clamRepo.Page(req.Page, req.PageSize, commonRepo.WithLikeName(req.Info), commonRepo.WithOrderRuleBy(req.OrderBy, req.Order))
	if err != nil {
		return 0, nil, err
	}
	var datas []dto.ClamInfo
	for _, command := range commands {
		var item dto.ClamInfo
		if err := copier.Copy(&item, &command); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		item.LastHandleDate = "-"
		datas = append(datas, item)
	}
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	for i := 0; i < len(datas); i++ {
		logPaths := loadFileByName(datas[i].Name)
		sort.Slice(logPaths, func(i, j int) bool {
			return logPaths[i] > logPaths[j]
		})
		if len(logPaths) != 0 {
			t1, err := time.ParseInLocation(constant.DateTimeSlimLayout, logPaths[0], nyc)
			if err != nil {
				continue
			}
			datas[i].LastHandleDate = t1.Format(constant.DateTimeLayout)
		}
		alertBase := dto.AlertBase{
			AlertType: "clams",
			EntryID:   datas[i].ID,
		}
		alertCount := xpack.GetAlert(alertBase)
		if alertCount != 0 {
			datas[i].AlertCount = alertCount
		} else {
			datas[i].AlertCount = 0
		}
	}
	return total, datas, err
}

func (c *ClamService) Create(req dto.ClamCreate) error {
	clam, _ := clamRepo.Get(commonRepo.WithByName(req.Name))
	if clam.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&clam, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if clam.InfectedStrategy == "none" || clam.InfectedStrategy == "remove" {
		clam.InfectedDir = ""
	}
	if len(req.Spec) != 0 {
		entryID, err := xpack.StartClam(clam, false)
		if err != nil {
			return err
		}
		clam.EntryID = entryID
		clam.Status = constant.StatusEnable
	}
	if err := clamRepo.Create(&clam); err != nil {
		return err
	}

	if req.AlertCount != 0 {
		createAlert := dto.CreateOrUpdateAlert{
			AlertTitle: req.AlertTitle,
			AlertCount: req.AlertCount,
			AlertType:  "clams",
			EntryID:    clam.ID,
		}
		err := xpack.CreateAlert(createAlert)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ClamService) Update(req dto.ClamUpdate) error {
	clam, _ := clamRepo.Get(commonRepo.WithByName(req.Name))
	if clam.ID == 0 {
		return constant.ErrRecordNotFound
	}
	if req.InfectedStrategy == "none" || req.InfectedStrategy == "remove" {
		req.InfectedDir = ""
	}
	var clamItem model.Clam
	if err := copier.Copy(&clamItem, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	clamItem.EntryID = clam.EntryID
	upMap := map[string]interface{}{}
	if len(clam.Spec) != 0 && clam.EntryID != 0 {
		global.Cron.Remove(cron.EntryID(clamItem.EntryID))
		upMap["entry_id"] = 0
	}
	if len(req.Spec) == 0 {
		upMap["status"] = ""
		upMap["entry_id"] = 0
	}
	if len(req.Spec) != 0 && clam.Status != constant.StatusDisable {
		newEntryID, err := xpack.StartClam(clamItem, true)
		if err != nil {
			return err
		}
		upMap["entry_id"] = newEntryID
	}
	if len(clam.Spec) == 0 && len(req.Spec) != 0 {
		upMap["status"] = constant.StatusEnable
	}

	upMap["name"] = req.Name
	upMap["path"] = req.Path
	upMap["infected_dir"] = req.InfectedDir
	upMap["infected_strategy"] = req.InfectedStrategy
	upMap["spec"] = req.Spec
	upMap["description"] = req.Description
	if err := clamRepo.Update(req.ID, upMap); err != nil {
		return err
	}
	updateAlert := dto.CreateOrUpdateAlert{
		AlertTitle: req.AlertTitle,
		AlertType:  "clams",
		AlertCount: req.AlertCount,
		EntryID:    clam.ID,
	}
	err := xpack.UpdateAlert(updateAlert)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClamService) UpdateStatus(id uint, status string) error {
	clam, _ := clamRepo.Get(commonRepo.WithByID(id))
	if clam.ID == 0 {
		return constant.ErrRecordNotFound
	}
	var (
		entryID int
		err     error
	)
	if status == constant.StatusEnable {
		entryID, err = xpack.StartClam(clam, true)
		if err != nil {
			return err
		}
	} else {
		global.Cron.Remove(cron.EntryID(clam.EntryID))
		global.LOG.Infof("stop cronjob entryID: %v", clam.EntryID)
	}

	return clamRepo.Update(clam.ID, map[string]interface{}{"status": status, "entry_id": entryID})
}

func (c *ClamService) Delete(req dto.ClamDelete) error {
	for _, id := range req.Ids {
		clam, _ := clamRepo.Get(commonRepo.WithByID(id))
		if clam.ID == 0 {
			continue
		}
		if req.RemoveRecord {
			_ = os.RemoveAll(path.Join(global.CONF.System.DataDir, resultDir, clam.Name))
		}
		if req.RemoveInfected {
			_ = os.RemoveAll(path.Join(clam.InfectedDir, "1panel-infected", clam.Name))
		}
		if err := clamRepo.Delete(commonRepo.WithByID(id)); err != nil {
			return err
		}
		alertBase := dto.AlertBase{
			AlertType: "clams",
			EntryID:   clam.ID,
		}
		err := xpack.DeleteAlert(alertBase)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ClamService) HandleOnce(req dto.OperateByID) error {
	if cleaned := StopAllCronJob(true); cleaned {
		return buserr.New("ErrClamdscanNotFound")
	}
	clam, _ := clamRepo.Get(commonRepo.WithByID(req.ID))
	if clam.ID == 0 {
		return constant.ErrRecordNotFound
	}
	if cmd.CheckIllegal(clam.Path) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	logFile := path.Join(global.CONF.System.DataDir, resultDir, clam.Name, timeNow)
	if _, err := os.Stat(path.Dir(logFile)); err != nil {
		_ = os.MkdirAll(path.Dir(logFile), os.ModePerm)
	}
	go func() {
		strategy := ""
		switch clam.InfectedStrategy {
		case "remove":
			strategy = "--remove"
		case "move":
			dir := path.Join(clam.InfectedDir, "1panel-infected", clam.Name, timeNow)
			strategy = "--move=" + dir
			if _, err := os.Stat(dir); err != nil {
				_ = os.MkdirAll(dir, os.ModePerm)
			}
		case "copy":
			dir := path.Join(clam.InfectedDir, "1panel-infected", clam.Name, timeNow)
			strategy = "--copy=" + dir
			if _, err := os.Stat(dir); err != nil {
				_ = os.MkdirAll(dir, os.ModePerm)
			}
		}
		global.LOG.Debugf("clamdscan --fdpass %s %s -l %s", strategy, clam.Path, logFile)
		stdout, err := cmd.Execf("clamdscan --fdpass %s %s -l %s", strategy, clam.Path, logFile)
		handleAlert(stdout, clam.Name, clam.ID)
		if err != nil {
			global.LOG.Errorf("clamdscan failed, stdout: %v, err: %v", stdout, err)
		}
	}()
	return nil
}

func (c *ClamService) LoadRecords(req dto.ClamLogSearch) (int64, interface{}, error) {
	clam, _ := clamRepo.Get(commonRepo.WithByID(req.ClamID))
	if clam.ID == 0 {
		return 0, nil, constant.ErrRecordNotFound
	}
	logPaths := loadFileByName(clam.Name)
	if len(logPaths) == 0 {
		return 0, nil, nil
	}

	var filterFiles []string
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	for _, item := range logPaths {
		t1, err := time.ParseInLocation(constant.DateTimeSlimLayout, item, nyc)
		if err != nil {
			continue
		}
		if t1.After(req.StartTime) && t1.Before(req.EndTime) {
			filterFiles = append(filterFiles, item)
		}
	}
	if len(filterFiles) == 0 {
		return 0, nil, nil
	}

	sort.Slice(filterFiles, func(i, j int) bool {
		return filterFiles[i] > filterFiles[j]
	})

	var records []string
	total, start, end := len(filterFiles), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]string, 0)
	} else {
		if end >= total {
			end = total
		}
		records = filterFiles[start:end]
	}

	var datas []dto.ClamLog
	for i := 0; i < len(records); i++ {
		item := loadResultFromLog(path.Join(global.CONF.System.DataDir, resultDir, clam.Name, records[i]))
		datas = append(datas, item)
	}
	return int64(total), datas, nil
}
func (c *ClamService) LoadRecordLog(req dto.ClamLogReq) (string, error) {
	logPath := path.Join(global.CONF.System.DataDir, resultDir, req.ClamName, req.RecordName)
	var tail string
	if req.Tail != "0" {
		tail = req.Tail
	} else {
		tail = "+1"
	}
	cmd := exec.Command("tail", "-n", tail, logPath)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tail -n %v failed, err: %v", req.Tail, err)
	}
	return string(stdout), nil
}

func (c *ClamService) CleanRecord(req dto.OperateByID) error {
	clam, _ := clamRepo.Get(commonRepo.WithByID(req.ID))
	if clam.ID == 0 {
		return constant.ErrRecordNotFound
	}
	pathItem := path.Join(global.CONF.System.DataDir, resultDir, clam.Name)
	_ = os.RemoveAll(pathItem)
	return nil
}

func (c *ClamService) LoadFile(req dto.ClamFileReq) (string, error) {
	filePath := ""
	switch req.Name {
	case "clamd":
		if c.serviceName == clamServiceNameUbuntu {
			filePath = "/etc/clamav/clamd.conf"
		} else {
			filePath = "/etc/clamd.d/scan.conf"
		}
	case "clamd-log":
		filePath = c.loadLogPath("clamd-log")
		if len(filePath) != 0 {
			break
		}
		if c.serviceName == clamServiceNameUbuntu {
			filePath = "/var/log/clamav/clamav.log"
		} else {
			filePath = "/var/log/clamd.scan"
		}
	case "freshclam":
		if c.serviceName == clamServiceNameUbuntu {
			filePath = "/etc/clamav/freshclam.conf"
		} else {
			filePath = "/etc/freshclam.conf"
		}
	case "freshclam-log":
		filePath = c.loadLogPath("freshclam-log")
		if len(filePath) != 0 {
			break
		}
		if c.serviceName == clamServiceNameUbuntu {
			filePath = "/var/log/clamav/freshclam.log"
		} else {
			filePath = "/var/log/freshclam.log"
		}
	default:
		return "", fmt.Errorf("not support such type")
	}
	if _, err := os.Stat(filePath); err != nil {
		return "", buserr.New("ErrHttpReqNotFound")
	}
	var tail string
	if req.Tail != "0" {
		tail = req.Tail
	} else {
		tail = "+1"
	}
	cmd := exec.Command("tail", "-n", tail, filePath)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tail -n %v failed, err: %v", req.Tail, err)
	}
	return string(stdout), nil
}

func (c *ClamService) UpdateFile(req dto.UpdateByNameAndFile) error {
	filePath := ""
	service := ""
	switch req.Name {
	case "clamd":
		if c.serviceName == clamServiceNameUbuntu {
			service = clamServiceNameUbuntu
			filePath = "/etc/clamav/clamd.conf"
		} else {
			service = clamServiceNameCentOs
			filePath = "/etc/clamd.d/scan.conf"
		}
	case "freshclam":
		if c.serviceName == clamServiceNameUbuntu {
			filePath = "/etc/clamav/freshclam.conf"
		} else {
			filePath = "/etc/freshclam.conf"
		}
		service = "clamav-freshclam.service"
	default:
		return fmt.Errorf("not support such type")
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	_ = systemctl.Restart(service)
	return nil
}

func StopAllCronJob(withCheck bool) bool {
	if withCheck {
		isActive := false
		exist1, _ := systemctl.IsExist(clamServiceNameCentOs)
		if exist1 {
			isActive, _ = systemctl.IsActive(clamServiceNameCentOs)
		}
		exist2, _ := systemctl.IsExist(clamServiceNameUbuntu)
		if exist2 {
			isActive, _ = systemctl.IsActive(clamServiceNameUbuntu)
		}
		if isActive {
			return false
		}
	}
	clams, _ := clamRepo.List(commonRepo.WithByStatus(constant.StatusEnable))
	for i := 0; i < len(clams); i++ {
		global.Cron.Remove(cron.EntryID(clams[i].EntryID))
		_ = clamRepo.Update(clams[i].ID, map[string]interface{}{"status": constant.StatusDisable, "entry_id": 0})
	}
	return true
}

func loadFileByName(name string) []string {
	var logPaths []string
	pathItem := path.Join(global.CONF.System.DataDir, resultDir, name)
	_ = filepath.Walk(pathItem, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || info.Name() == name {
			return nil
		}
		logPaths = append(logPaths, info.Name())
		return nil
	})
	return logPaths
}
func loadResultFromLog(pathItem string) dto.ClamLog {
	var data dto.ClamLog
	data.Name = path.Base(pathItem)
	data.Status = constant.StatusWaiting
	file, err := os.ReadFile(pathItem)
	if err != nil {
		return data
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if strings.Contains(line, "- SCAN SUMMARY -") {
			data.Status = constant.StatusDone
		}
		if data.Status != constant.StatusDone {
			continue
		}
		switch {
		case strings.HasPrefix(line, "Infected files:"):
			data.InfectedFiles = strings.TrimPrefix(line, "Infected files:")
		case strings.HasPrefix(line, "Total errors:"):
			data.TotalError = strings.TrimPrefix(line, "Total errors:")
		case strings.HasPrefix(line, "Time:"):
			if strings.Contains(line, "(") {
				data.ScanTime = strings.ReplaceAll(strings.Split(line, "(")[1], ")", "")
				continue
			}
			data.ScanTime = strings.TrimPrefix(line, "Time:")
		case strings.HasPrefix(line, "Start Date:"):
			data.ScanDate = strings.TrimPrefix(line, "Start Date:")
		}
	}
	return data
}
func (c *ClamService) loadLogPath(name string) string {
	confPath := ""
	if name == "clamd-log" {
		if c.serviceName == clamServiceNameUbuntu {
			confPath = "/etc/clamav/clamd.conf"
		} else {
			confPath = "/etc/clamd.d/scan.conf"
		}
	} else {
		if c.serviceName == clamServiceNameUbuntu {
			confPath = "/etc/clamav/freshclam.conf"
		} else {
			confPath = "/etc/freshclam.conf"
		}
	}
	if _, err := os.Stat(confPath); err != nil {
		return ""
	}
	content, err := os.ReadFile(confPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	if name == "clamd-log" {
		for _, line := range lines {
			if strings.HasPrefix(line, "LogFile ") {
				return strings.Trim(strings.ReplaceAll(line, "LogFile ", ""), " ")
			}
		}
	} else {
		for _, line := range lines {
			if strings.HasPrefix(line, "UpdateLogFile ") {
				return strings.Trim(strings.ReplaceAll(line, "UpdateLogFile ", ""), " ")
			}
		}
	}

	return ""
}

func handleAlert(stdout, clamName string, clamId uint) {
	if strings.Contains(stdout, "- SCAN SUMMARY -") {
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Infected files: ") {
				infectedFiles, _ := strconv.Atoi(strings.TrimPrefix(line, "Infected files: "))
				if infectedFiles > 0 {
					pushAlert := dto.PushAlert{
						TaskName:  clamName,
						AlertType: "clams",
						EntryID:   clamId,
						Param:     strconv.Itoa(infectedFiles),
					}
					err := xpack.PushAlert(pushAlert)
					if err != nil {
						global.LOG.Errorf("clamdscan push failed, err: %v", err)
					}
					break
				}
			}
		}
	}
}
