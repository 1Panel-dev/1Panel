package service

import (
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/copier"
	"github.com/1Panel-dev/1Panel/agent/utils/email"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/shirou/gopsutil/v3/disk"
	"sort"
	"strings"
	"sync"
	"time"
)

type AlertService struct{}

type IAlertService interface {
	PageAlert(req dto.AlertSearch) (int64, []dto.AlertDTO, error)
	GetAlerts() ([]dto.AlertDTO, error)
	CreateAlert(create dto.AlertCreate) error
	UpdateAlert(req dto.AlertUpdate) error
	DeleteAlert(id uint) error
	GetAlert(id uint) (dto.AlertDTO, error)
	UpdateStatus(id uint, status string) error
	ExternalUpdateAlert(req dto.AlertCreate) error

	GetDisks() ([]dto.DiskDTO, error)
	PageAlertLogs(req dto.AlertLogSearch) (int64, []dto.AlertLogDTO, error)
	CleanAlertLogs() error
	GetClams() ([]dto.ClamDTO, error)
	GetCronJobs(req dto.CronJobReq) ([]dto.CronJobDTO, error)

	GetAlertConfig() ([]model.AlertConfig, error)
	UpdateAlertConfig(req dto.AlertConfigUpdate) error
	DeleteAlertConfig(id uint) error
	TestAlertConfig(req dto.AlertConfigTest) (bool, error)
}

func NewIAlertService() IAlertService {
	return &AlertService{}
}

func (a AlertService) PageAlert(search dto.AlertSearch) (int64, []dto.AlertDTO, error) {
	var (
		opts   []repo.DBOption
		result []dto.AlertDTO
	)
	if search.Status != "" {
		opts = append(opts, repo.WithByStatus(search.Status))
	}
	if search.Type != "" {
		opts = append(opts, alertRepo.WithByType(search.Type))
	}
	opts = append(opts, repo.WithOrderBy("created_at desc"))

	total, alerts, err := alertRepo.Page(search.Page, search.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	for _, item := range alerts {

		result = append(result, dto.AlertDTO{
			ID:        item.ID,
			Type:      item.Type,
			Cycle:     item.Cycle,
			Count:     item.Count,
			Method:    item.Method,
			Title:     item.Title,
			Project:   item.Project,
			Status:    item.Status,
			SendCount: item.SendCount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return total, result, err
}

func (a AlertService) GetAlerts() ([]dto.AlertDTO, error) {
	var (
		opts   []repo.DBOption
		result []dto.AlertDTO
	)
	opts = append(opts, repo.WithByStatus(constant.AlertEnable))
	alerts, err := alertRepo.List(opts...)
	if err != nil {
		return nil, err
	}
	for _, item := range alerts {

		result = append(result, dto.AlertDTO{
			ID:        item.ID,
			Type:      item.Type,
			Cycle:     item.Cycle,
			Count:     item.Count,
			Method:    item.Method,
			Title:     item.Title,
			Project:   item.Project,
			Status:    item.Status,
			SendCount: item.SendCount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return result, err
}

func (a AlertService) CreateAlert(create dto.AlertCreate) error {
	var alertID uint
	var alertInfo model.Alert
	if create.Project != "" {
		alertInfo, _ := alertRepo.Get(alertRepo.WithByType(create.Type), alertRepo.WithByProject(create.Project))
		alertID = alertInfo.ID
	} else {
		alertInfo, _ := alertRepo.Get(alertRepo.WithByType(create.Type))
		alertID = alertInfo.ID
	}

	if alertID != 0 {
		var upAlert dto.AlertUpdate
		if err := copier.Copy(&upAlert, &create); err != nil {
			return buserr.WithErr("ErrStructTransform", err)
		}
		upAlert.ID = alertID
		err := a.UpdateAlert(upAlert)
		if err != nil {
			return err
		}
	} else {
		alertInfo.Status = constant.AlertEnable
		if err := copier.Copy(&alertInfo, &create); err != nil {
			return buserr.WithErr("ErrStructTransform", err)
		}

		if err := alertRepo.Create(&alertInfo); err != nil {
			return err
		}
		NewIAlertTaskHelper().InitTask(alertInfo.Type)
	}

	return nil
}

func (a AlertService) UpdateAlert(req dto.AlertUpdate) error {

	upMap := make(map[string]interface{})
	upMap["id"] = req.ID
	upMap["type"] = req.Type
	upMap["cycle"] = req.Cycle
	upMap["count"] = req.Count
	upMap["method"] = req.Method
	upMap["title"] = req.Title
	upMap["project"] = req.Project
	upMap["status"] = req.Status
	upMap["send_count"] = req.SendCount

	if err := alertRepo.Update(upMap, repo.WithByID(req.ID)); err != nil {
		return err
	}
	NewIAlertTaskHelper().InitTask(req.Type)
	return nil
}

func (a AlertService) DeleteAlert(id uint) error {
	return alertRepo.Delete(repo.WithByID(id))
}

func (a AlertService) GetAlert(id uint) (dto.AlertDTO, error) {
	var res dto.AlertDTO
	alertInfo, err := alertRepo.Get(repo.WithByID(id))
	if err != nil {
		return res, err
	}
	_ = copier.Copy(&res, &alertInfo)
	return res, nil
}

func (a AlertService) UpdateStatus(id uint, status string) error {
	alertInfo, _ := alertRepo.Get(repo.WithByID(id))
	if alertInfo.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	err := alertRepo.Update(map[string]interface{}{"status": status}, repo.WithByID(alertInfo.ID))
	if err != nil {
		return err
	}
	alerts, err := a.GetAlerts()
	if err != nil {
		return err
	}
	if len(alerts) > 0 {
		NewIAlertTaskHelper().InitTask(alertInfo.Type)
	} else {
		NewIAlertTaskHelper().StopTask()
	}
	return nil
}

func (a AlertService) GetDisks() ([]dto.DiskDTO, error) {
	var disks []dto.DiskDTO
	excludes := map[string]struct{}{
		"/mnt/cdrom": {}, "/boot": {}, "/boot/efi": {}, "/dev": {}, "/dev/shm": {},
		"/run/lock": {}, "/run": {}, "/run/shm": {}, "/run/user": {},
	}
	stdout, err := executeDiskCommand()
	if err != nil {
		return disks, nil
	}

	lines := strings.Split(stdout, "\n")
	var mounts []dto.AlertDiskInfo

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		mountPoint := strings.Join(fields[6:], " ")
		if shouldExclude(fields, mountPoint, excludes) {
			continue
		}
		mounts = append(mounts, dto.AlertDiskInfo{Type: fields[1], Device: fields[0], Mount: mountPoint})

	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	wg.Add(len(mounts))
	for i := 0; i < len(mounts); i++ {
		go func(timeoutCh <-chan time.Time, mount dto.AlertDiskInfo) {
			defer wg.Done()

			var itemData dto.DiskDTO
			itemData.Path = mount.Mount
			itemData.Type = mount.Type
			itemData.Device = mount.Device
			select {
			case <-timeoutCh:
				mu.Lock()
				disks = append(disks, itemData)
				mu.Unlock()
				global.LOG.Errorf("load disk info from %s failed, err: timeout", mount.Mount)
			default:
				state, err := disk.Usage(mount.Mount)
				if err != nil {
					mu.Lock()
					disks = append(disks, itemData)
					mu.Unlock()
					global.LOG.Errorf("load disk info from %s failed, err: %v", mount.Mount, err)
					return
				}
				itemData.Total = state.Total
				itemData.Free = state.Free
				itemData.Used = state.Used
				itemData.UsedPercent = state.UsedPercent
				itemData.InodesTotal = state.InodesTotal
				itemData.InodesUsed = state.InodesUsed
				itemData.InodesFree = state.InodesFree
				itemData.InodesUsedPercent = state.InodesUsedPercent
				mu.Lock()
				disks = append(disks, itemData)
				mu.Unlock()
			}
		}(time.After(5*time.Second), mounts[i])
	}
	wg.Wait()

	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Path < disks[j].Path
	})
	return disks, nil
}

func executeDiskCommand() (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashC("df -hT -P | grep '/' | grep -v tmpfs | grep -v 'snap/core' | grep -v udev")
	if err != nil {
		cmdMgr2 := cmd.NewCommandMgr(cmd.WithTimeout(1 * time.Second))
		stdout, err = cmdMgr2.RunWithStdoutBashC("df -lhT -P | grep '/' | grep -v tmpfs | grep -v 'snap/core' | grep -v udev")
	}
	return stdout, err
}

func shouldExclude(fields []string, mountPoint string, excludes map[string]struct{}) bool {
	if strings.HasPrefix(mountPoint, "/snap") || len(strings.Split(mountPoint, "/")) > 10 {
		return true
	}
	if strings.TrimSpace(fields[1]) == "tmpfs" {
		return true
	}
	if strings.Contains(fields[2], "K") {
		return true
	}
	if strings.Contains(mountPoint, "docker") {
		return true
	}
	_, excluded := excludes[mountPoint]
	return excluded
}

func (a AlertService) PageAlertLogs(search dto.AlertLogSearch) (int64, []dto.AlertLogDTO, error) {
	var (
		opts   []repo.DBOption
		result []dto.AlertLogDTO
	)
	if search.Status != "" {
		opts = append(opts, repo.WithByStatus(search.Status))
	}
	if search.Count != 0 {
		opts = append(opts, alertRepo.WithByCount(search.Count))
	}
	opts = append(opts, repo.WithOrderBy("created_at desc"))

	total, alerts, err := alertRepo.PageLog(search.Page, search.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	for _, item := range alerts {
		alertLogDTO, err := a.parseAlertLog(item)
		if err != nil {
			return 0, nil, err
		}
		result = append(result, alertLogDTO)
	}

	return total, result, err
}

func (a AlertService) parseAlertLog(item model.AlertLog) (dto.AlertLogDTO, error) {
	var alertDetail dto.AlertDetail
	var alertRule dto.AlertRule

	if err := unmarshalAlertInfo(item.AlertDetail, &alertDetail); err != nil {
		return dto.AlertLogDTO{}, err
	}
	if err := unmarshalAlertInfo(item.AlertRule, &alertRule); err != nil {
		return dto.AlertLogDTO{}, err
	}
	return dto.AlertLogDTO{
		ID:          item.ID,
		Count:       item.Count,
		Type:        item.Type,
		Status:      item.Status,
		Method:      item.Method,
		Message:     item.Message,
		AlertId:     item.AlertId,
		AlertDetail: alertDetail,
		AlertRule:   alertRule,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}, nil
}

func unmarshalAlertInfo(data string, v interface{}) error {
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("unmarshal alert info vars failed, err: %v", err)
	}
	return nil
}

func (a AlertService) CleanAlertLogs() error {
	return alertRepo.CleanAlertLogs()
}

func (a AlertService) GetClams() ([]dto.ClamDTO, error) {
	var clams []dto.ClamDTO
	clamList, err := clamRepo.List()

	for _, clam := range clamList {
		var clamDTO dto.ClamDTO
		clamDTO.ID = clam.ID
		clamDTO.Name = clam.Name
		clamDTO.Path = clam.Path
		clamDTO.Status = clam.Status
		clamDTO.UpdatedAt = clam.UpdatedAt
		clamDTO.CreatedAt = clam.CreatedAt
		clams = append(clams, clamDTO)
	}
	return clams, err
}

func (a AlertService) GetCronJobs(req dto.CronJobReq) ([]dto.CronJobDTO, error) {
	var cronJobs []dto.CronJobDTO
	var (
		opts []repo.DBOption
	)
	if req.Status != "" {
		opts = append(opts, repo.WithByStatus(req.Status))
	}
	if req.Type != "" {
		opts = append(opts, repo.WithByType(req.Type))
	}
	cronjobList, err := cronjobRepo.List(opts...)

	for _, cronJob := range cronjobList {
		var cronJobDTO dto.CronJobDTO
		cronJobDTO.ID = cronJob.ID
		cronJobDTO.Name = cronJob.Name
		cronJobDTO.Status = cronJob.Status
		cronJobDTO.Type = cronJob.Type
		cronJobDTO.UpdatedAt = cronJob.UpdatedAt
		cronJobDTO.CreatedAt = cronJob.CreatedAt
		cronJobs = append(cronJobs, cronJobDTO)
	}
	return cronJobs, err
}

func (a AlertService) GetAlertConfig() ([]model.AlertConfig, error) {
	var (
		opts    []repo.DBOption
		configs []model.AlertConfig
	)
	opts = append(opts, repo.WithByStatus(constant.AlertEnable))
	configs, err := alertRepo.AlertConfigList(opts...)
	return configs, err
}

func (a AlertService) UpdateAlertConfig(req dto.AlertConfigUpdate) error {
	if req.ID != 0 {
		upMap := make(map[string]interface{})
		upMap["id"] = req.ID
		upMap["type"] = req.Type
		upMap["title"] = req.Title
		upMap["status"] = req.Status
		upMap["config"] = req.Config
		if err := alertRepo.UpdateAlertConfig(upMap, repo.WithByID(req.ID)); err != nil {
			return err
		}
	} else {
		var alertConfig model.AlertConfig
		if err := copier.Copy(&alertConfig, &req); err != nil {
			return buserr.WithErr("ErrStructTransform", err)
		}
		if err := alertRepo.CreateAlertConfig(&alertConfig); err != nil {
			return err
		}
	}

	return nil
}

func (a AlertService) DeleteAlertConfig(id uint) error {
	return alertRepo.DeleteAlertConfig(repo.WithByID(id))
}

func (a AlertService) TestAlertConfig(req dto.AlertConfigTest) (bool, error) {
	cfg := email.SMTPConfig{
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Sender,
		Password:   req.Password,
		From:       fmt.Sprintf("%s <%s>", req.DisplayName, req.Sender),
		Encryption: req.Encryption,
		Recipient:  req.Recipient,
	}

	msg := email.EmailMessage{
		Subject: i18n.GetMsgByKey("TestAlertTitle"),
		Body:    i18n.GetMsgByKey("TestAlert"),
		IsHTML:  false,
	}
	transport := xpack.LoadRequestTransport()
	if err := email.SendMail(cfg, msg, transport); err != nil {
		return false, err
	}
	return true, nil
}

func (a AlertService) ExternalUpdateAlert(updateAlert dto.AlertCreate) error {
	upMap := make(map[string]interface{})
	if updateAlert.SendCount == 0 {
		upMap["status"] = constant.AlertDisable
	} else {
		upMap["status"] = constant.AlertEnable
		upMap["send_count"] = updateAlert.SendCount
	}
	upMap["method"] = updateAlert.Method
	alertInfo, _ := alertRepo.Get(alertRepo.WithByType(updateAlert.Type), alertRepo.WithByProject(updateAlert.Project))
	if alertInfo.ID > 0 {
		if err := alertRepo.Update(upMap, alertRepo.WithByProject(updateAlert.Project), alertRepo.WithByType(updateAlert.Type)); err != nil {
			return err
		}
	} else {
		updateAlert.Status = constant.AlertEnable
		err := a.CreateAlert(updateAlert)
		if err != nil {
			return err
		}
	}

	return nil
}
