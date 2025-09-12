package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/alert_push"
	"github.com/1Panel-dev/1Panel/agent/utils/clam"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/systemctl"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/jinzhu/copier"
	"github.com/robfig/cron/v3"
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
	HandleOnce(id uint) error

	LoadFile(req dto.ClamFileReq) (string, error)
	UpdateFile(req dto.UpdateByNameAndFile) error

	SearchRecords(req dto.ClamLogSearch) (int64, interface{}, error)
	CleanRecord(id uint) error
}

func NewIClamService() IClamService {
	return &ClamService{}
}

func (c *ClamService) LoadBaseInfo() (dto.ClamBaseInfo, error) {
	var baseInfo dto.ClamBaseInfo
	baseInfo.Version = "-"
	baseInfo.FreshVersion = "-"
	exist1, _ := systemctl.IsExist(constant.ClamServiceNameCentOs)
	if exist1 {
		c.serviceName = constant.ClamServiceNameCentOs
		baseInfo.IsExist = true
		baseInfo.IsActive, _ = systemctl.IsActive(constant.ClamServiceNameCentOs)
	}
	exist2, _ := systemctl.IsExist(constant.ClamServiceNameUbuntu)
	if exist2 {
		c.serviceName = constant.ClamServiceNameUbuntu
		baseInfo.IsExist = true
		baseInfo.IsActive, _ = systemctl.IsActive(constant.ClamServiceNameUbuntu)
	}
	freshExist, _ := systemctl.IsExist(constant.FreshClamService)
	if freshExist {
		baseInfo.FreshIsExist = true
		baseInfo.FreshIsActive, _ = systemctl.IsActive(constant.FreshClamService)
	}
	if !cmd.Which("clamdscan") {
		baseInfo.IsActive = false
	}

	if baseInfo.IsActive {
		version, err := cmd.RunDefaultWithStdoutBashC("clamdscan --version")
		if err == nil {
			if strings.Contains(version, "/") {
				baseInfo.Version = strings.TrimPrefix(strings.Split(version, "/")[0], "ClamAV ")
			} else {
				baseInfo.Version = strings.TrimPrefix(version, "ClamAV ")
			}
		}
	} else {
		_ = clam.CheckClamIsActive(false, clamRepo)
	}
	if baseInfo.FreshIsActive {
		version, err := cmd.RunDefaultWithStdoutBashC("freshclam --version")
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
		stdout, err := cmd.RunDefaultWithStdoutBashCf("systemctl %s %s", operate, c.serviceName)
		if err != nil {
			return fmt.Errorf("%s the %s failed, err: %s", operate, c.serviceName, stdout)
		}
		return nil
	case "fresh-start", "fresh-restart", "fresh-stop":
		stdout, err := cmd.RunDefaultWithStdoutBashCf("systemctl %s %s", strings.TrimPrefix(operate, "fresh-"), constant.FreshClamService)
		if err != nil {
			return fmt.Errorf("%s the %s failed, err: %s", operate, c.serviceName, stdout)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (c *ClamService) SearchWithPage(req dto.SearchClamWithPage) (int64, interface{}, error) {
	total, clams, err := clamRepo.Page(req.Page, req.PageSize, repo.WithByLikeName(req.Info), repo.WithOrderRuleBy(req.OrderBy, req.Order))
	if err != nil {
		return 0, nil, err
	}
	var datas []dto.ClamInfo
	for _, clam := range clams {
		var item dto.ClamInfo
		if err := copier.Copy(&item, &clam); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		datas = append(datas, item)
	}
	for i := 0; i < len(datas); i++ {
		record, _ := clamRepo.RecordFirst(datas[i].ID)
		if record.ID != 0 {
			datas[i].LastRecordStatus = record.Status
			datas[i].LastRecordTime = record.StartTime.Format(constant.DateTimeLayout)
		} else {
			datas[i].LastRecordTime = "-"
		}
		alertBase := dto.AlertBase{
			AlertType: "clams",
			EntryID:   datas[i].ID,
		}
		alertInfo, _ := alertRepo.Get(alertRepo.WithByType(alertBase.AlertType), alertRepo.WithByProject(strconv.Itoa(int(alertBase.EntryID))), repo.WithByStatus(constant.AlertEnable))
		datas[i].AlertMethod = alertInfo.Method
		if alertInfo.SendCount != 0 {
			datas[i].AlertCount = alertInfo.SendCount
		} else {
			datas[i].AlertCount = 0
		}
	}
	return total, datas, err
}

func (c *ClamService) Create(req dto.ClamCreate) error {
	clam, _ := clamRepo.Get(repo.WithByName(req.Name))
	if clam.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	if cmd.CheckIllegal(req.Path) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := copier.Copy(&clam, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	if clam.InfectedStrategy == "none" || clam.InfectedStrategy == "remove" {
		clam.InfectedDir = ""
	}
	if len(req.Spec) != 0 {
		entryID, err := xpack.StartClam(&clam, false)
		if err != nil {
			return err
		}
		clam.EntryID = entryID
		clam.Status = constant.StatusEnable
	}
	if err := clamRepo.Create(&clam); err != nil {
		return err
	}
	if req.AlertCount != 0 && req.AlertTitle != "" && req.AlertMethod != "" {
		createAlert := dto.AlertCreate{
			Title:     req.AlertTitle,
			SendCount: req.AlertCount,
			Method:    req.AlertMethod,
			Type:      "clams",
			Project:   strconv.Itoa(int(clam.ID)),
			Status:    constant.AlertEnable,
		}
		err := NewIAlertService().CreateAlert(createAlert)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ClamService) Update(req dto.ClamUpdate) error {
	if cmd.CheckIllegal(req.Path) {
		return buserr.New("ErrCmdIllegal")
	}
	clam, _ := clamRepo.Get(repo.WithByName(req.Name))
	if clam.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if req.InfectedStrategy == "none" || req.InfectedStrategy == "remove" {
		req.InfectedDir = ""
	}
	var clamItem model.Clam
	if err := copier.Copy(&clamItem, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
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
		newEntryID, err := xpack.StartClam(&clamItem, true)
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
	updateAlert := dto.AlertCreate{
		Title:     req.AlertTitle,
		SendCount: req.AlertCount,
		Method:    req.AlertMethod,
		Type:      "clams",
		Project:   strconv.Itoa(int(clam.ID)),
	}
	err := NewIAlertService().ExternalUpdateAlert(updateAlert)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClamService) UpdateStatus(id uint, status string) error {
	clam, _ := clamRepo.Get(repo.WithByID(id))
	if clam.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	var (
		entryID int
		err     error
	)
	if status == constant.StatusEnable {
		entryID, err = xpack.StartClam(&clam, true)
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
		clam, _ := clamRepo.Get(repo.WithByID(id))
		if clam.ID == 0 {
			continue
		}
		if len(clam.Spec) != 0 {
			global.Cron.Remove(cron.EntryID(clam.EntryID))
		}
		_ = c.CleanRecord(clam.ID)
		if req.RemoveInfected {
			_ = os.RemoveAll(path.Join(clam.InfectedDir, "1panel-infected", clam.Name))
		}
		if err := clamRepo.Delete(repo.WithByID(id)); err != nil {
			return err
		}
		err := alertRepo.Delete(alertRepo.WithByProject(strconv.Itoa(int(clam.ID))), alertRepo.WithByType("clams"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ClamService) HandleOnce(id uint) error {
	if active := clam.CheckClamIsActive(true, clamRepo); !active {
		return buserr.New("ErrClamdscanNotFound")
	}
	clamItem, _ := clamRepo.Get(repo.WithByID(id))
	if clamItem.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	record := clamRepo.StartRecords(clamItem.ID)
	taskItem, err := task.NewTaskWithOps("clam-"+clamItem.Name, task.TaskScan, task.TaskScopeClam, record.TaskID, clamItem.ID)
	if err != nil {
		return fmt.Errorf("new task for exec shell failed, err: %v", err)
	}
	clam.AddScanTask(taskItem, clamItem, record.StartTime.Format(constant.DateTimeSlimLayout))
	go func() {
		err := taskItem.Execute()
		taskRepo := repo.NewITaskRepo()
		taskItem, _ := taskRepo.GetFirst(taskRepo.WithByID(record.TaskID))
		if len(taskItem.ID) == 0 {
			record.TaskID = ""
		}
		if err != nil {
			clamRepo.EndRecords(record, constant.StatusFailed, err.Error())
			return
		}
		handleAlert(record.InfectedFiles, clamItem.Name, clamItem.ID)
		clam.AnalysisFromLog(taskItem.LogFile, &record)
		clamRepo.EndRecords(record, constant.StatusDone, "")
	}()
	return nil
}

func (c *ClamService) SearchRecords(req dto.ClamLogSearch) (int64, interface{}, error) {
	clam, _ := clamRepo.Get(repo.WithByID(req.ClamID))
	if clam.ID == 0 {
		return 0, nil, buserr.New("ErrRecordNotFound")
	}

	total, records, err := clamRepo.PageRecords(req.Page, req.PageSize, clamRepo.WithByClamID(req.ClamID), repo.WithByStatus(req.Status), repo.WithByCreatedAt(req.StartTime, req.EndTime))
	if err != nil {
		return 0, nil, err
	}
	var datas []dto.ClamRecord
	for _, record := range records {
		var item dto.ClamRecord
		if err := copier.Copy(&item, &record); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		datas = append(datas, item)
	}
	return int64(total), datas, nil
}

func (c *ClamService) CleanRecord(id uint) error {
	record, _ := clamRepo.ListRecord()
	for _, item := range record {
		if len(item.TaskID) != 0 {
			continue
		}
		taskItem, _ := taskRepo.GetFirst(taskRepo.WithByID(item.TaskID))
		if len(taskItem.LogFile) != 0 {
			_ = os.Remove(taskItem.LogFile)
		}
	}
	return clamRepo.DeleteRecord(clamRepo.WithByClamID(id))
}

func (c *ClamService) LoadFile(req dto.ClamFileReq) (string, error) {
	filePath := ""
	switch req.Name {
	case "clamd":
		if c.serviceName == constant.ClamServiceNameUbuntu {
			filePath = "/etc/clamav/clamd.conf"
		} else {
			filePath = "/etc/clamd.d/scan.conf"
		}
	case "clamd-log":
		filePath = c.loadLogPath("clamd-log")
		if len(filePath) != 0 {
			break
		}
		if c.serviceName == constant.ClamServiceNameUbuntu {
			filePath = "/var/log/clamav/clamav.log"
		} else {
			filePath = "/var/log/clamd.scan"
		}
	case "freshclam":
		if c.serviceName == constant.ClamServiceNameUbuntu {
			filePath = "/etc/clamav/freshclam.conf"
		} else {
			filePath = "/etc/freshclam.conf"
		}
	case "freshclam-log":
		filePath = c.loadLogPath("freshclam-log")
		if len(filePath) != 0 {
			break
		}
		if c.serviceName == constant.ClamServiceNameUbuntu {
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
		if c.serviceName == constant.ClamServiceNameUbuntu {
			service = constant.ClamServiceNameUbuntu
			filePath = "/etc/clamav/clamd.conf"
		} else {
			service = constant.ClamServiceNameCentOs
			filePath = "/etc/clamd.d/scan.conf"
		}
	case "freshclam":
		if c.serviceName == constant.ClamServiceNameUbuntu {
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

func (c *ClamService) loadLogPath(name string) string {
	confPath := ""
	if name == "clamd-log" {
		if c.serviceName == constant.ClamServiceNameUbuntu {
			confPath = "/etc/clamav/clamd.conf"
		} else {
			confPath = "/etc/clamd.d/scan.conf"
		}
	} else {
		if c.serviceName == constant.ClamServiceNameUbuntu {
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

func handleAlert(infectedFiles, clamName string, clamId uint) {
	itemInfected, _ := strconv.Atoi(strings.TrimSpace(infectedFiles))
	if itemInfected < 0 {
		return
	}
	pushAlert := dto.PushAlert{
		TaskName:  clamName,
		AlertType: "clams",
		EntryID:   clamId,
		Param:     strconv.Itoa(itemInfected),
	}
	_ = alert_push.PushAlert(pushAlert)
}
