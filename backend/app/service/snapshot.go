package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/host"
)

type SnapshotService struct {
	OriginalPath string
}

type ISnapshotService interface {
	SearchWithPage(req dto.PageSnapshot) (int64, interface{}, error)
	LoadSize(req dto.PageSnapshot) ([]dto.SnapshotFile, error)
	SnapshotCreate(req dto.SnapshotCreate) error
	SnapshotRecover(req dto.SnapshotRecover) error
	SnapshotRollback(req dto.SnapshotRecover) error
	SnapshotImport(req dto.SnapshotImport) error
	Delete(req dto.SnapshotBatchDelete) error

	LoadSnapShotStatus(id uint) (*dto.SnapshotStatus, error)

	UpdateDescription(req dto.UpdateDescription) error
	readFromJson(path string) (SnapshotJson, error)

	HandleSnapshot(isCronjob bool, logPath string, req dto.SnapshotCreate, timeNow string, secret string) (string, error)
}

func NewISnapshotService() ISnapshotService {
	return &SnapshotService{}
}

func (u *SnapshotService) SearchWithPage(req dto.PageSnapshot) (int64, interface{}, error) {
	total, records, err := snapshotRepo.Page(req.Page, req.PageSize, commonRepo.WithLikeName(req.Info), commonRepo.WithOrderRuleBy(req.OrderBy, req.Order))
	if err != nil {
		return 0, nil, err
	}
	var data []dto.SnapshotInfo
	for _, record := range records {
		var item dto.SnapshotInfo
		if err := copier.Copy(&item, &record); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		data = append(data, item)
	}
	return total, data, err
}

func (u *SnapshotService) LoadSize(req dto.PageSnapshot) ([]dto.SnapshotFile, error) {
	_, records, err := snapshotRepo.Page(req.Page, req.PageSize, commonRepo.WithLikeName(req.Info), commonRepo.WithOrderRuleBy(req.OrderBy, req.Order))
	if err != nil {
		return nil, err
	}
	data, err := loadSnapSize(records)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (u *SnapshotService) SnapshotImport(req dto.SnapshotImport) error {
	if len(req.Names) == 0 {
		return fmt.Errorf("incorrect snapshot request body: %v", req.Names)
	}
	for _, snapName := range req.Names {
		snap, _ := snapshotRepo.Get(commonRepo.WithByName(strings.ReplaceAll(snapName, ".tar.gz", "")))
		if snap.ID != 0 {
			return constant.ErrRecordExist
		}
	}
	for _, snap := range req.Names {
		shortName := strings.TrimPrefix(snap, "snapshot_")
		nameItems := strings.Split(shortName, "_")
		if !strings.HasPrefix(shortName, "1panel_v") || !strings.HasSuffix(shortName, ".tar.gz") || len(nameItems) < 3 {
			return fmt.Errorf("incorrect snapshot name format of %s", shortName)
		}
		if strings.HasSuffix(snap, ".tar.gz") {
			snap = strings.ReplaceAll(snap, ".tar.gz", "")
		}
		itemSnap := model.Snapshot{
			Name:            snap,
			From:            req.From,
			DefaultDownload: req.From,
			Version:         nameItems[1],
			Description:     req.Description,
			Status:          constant.StatusSuccess,
		}
		if err := snapshotRepo.Create(&itemSnap); err != nil {
			return err
		}
	}
	return nil
}

func (u *SnapshotService) UpdateDescription(req dto.UpdateDescription) error {
	return snapshotRepo.Update(req.ID, map[string]interface{}{"description": req.Description})
}

func (u *SnapshotService) LoadSnapShotStatus(id uint) (*dto.SnapshotStatus, error) {
	var data dto.SnapshotStatus
	status, err := snapshotRepo.GetStatus(id)
	if err != nil {
		return nil, err
	}
	if err := copier.Copy(&data, &status); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return &data, nil
}

type SnapshotJson struct {
	OldBaseDir       string `json:"oldBaseDir"`
	OldDockerDataDir string `json:"oldDockerDataDir"`
	OldBackupDataDir string `json:"oldBackupDataDir"`
	OldPanelDataDir  string `json:"oldPanelDataDir"`

	BaseDir            string `json:"baseDir"`
	DockerDataDir      string `json:"dockerDataDir"`
	BackupDataDir      string `json:"backupDataDir"`
	PanelDataDir       string `json:"panelDataDir"`
	LiveRestoreEnabled bool   `json:"liveRestoreEnabled"`
}

func (u *SnapshotService) SnapshotCreate(req dto.SnapshotCreate) error {
	if _, err := u.HandleSnapshot(false, "", req, time.Now().Format(constant.DateTimeSlimLayout), req.Secret); err != nil {
		return err
	}
	return nil
}

func (u *SnapshotService) SnapshotRecover(req dto.SnapshotRecover) error {
	global.LOG.Info("start to recover panel by snapshot now")
	snap, err := snapshotRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if hasOs(snap.Name) && !strings.Contains(snap.Name, loadOs()) {
		return fmt.Errorf("restoring snapshots(%s) between different server architectures(%s) is not supported", snap.Name, loadOs())
	}
	if !req.IsNew && len(snap.InterruptStep) != 0 && len(snap.RollbackStatus) != 0 {
		return fmt.Errorf("the snapshot has been rolled back and cannot be restored again")
	}

	baseDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("system/%s", snap.Name))
	if _, err := os.Stat(baseDir); err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(baseDir, os.ModePerm)
	}

	_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"recover_status": constant.StatusWaiting})
	_ = settingRepo.Update("SystemStatus", "Recovering")
	go u.HandleSnapshotRecover(snap, true, req)
	return nil
}

func (u *SnapshotService) SnapshotRollback(req dto.SnapshotRecover) error {
	global.LOG.Info("start to rollback now")
	snap, err := snapshotRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	req.IsNew = false
	snap.InterruptStep = "Readjson"
	go u.HandleSnapshotRecover(snap, false, req)
	return nil
}

func (u *SnapshotService) readFromJson(path string) (SnapshotJson, error) {
	var snap SnapshotJson
	if _, err := os.Stat(path); err != nil {
		return snap, fmt.Errorf("find snapshot json file in recover package failed, err: %v", err)
	}
	fileByte, err := os.ReadFile(path)
	if err != nil {
		return snap, fmt.Errorf("read file from path %s failed, err: %v", path, err)
	}
	if err := json.Unmarshal(fileByte, &snap); err != nil {
		return snap, fmt.Errorf("unmarshal snapjson failed, err: %v", err)
	}
	return snap, nil
}

func (u *SnapshotService) HandleSnapshot(isCronjob bool, logPath string, req dto.SnapshotCreate, timeNow string, secret string) (string, error) {
	localDir, err := loadLocalDir()
	if err != nil {
		return "", err
	}
	var (
		rootDir    string
		snap       model.Snapshot
		snapStatus model.SnapshotStatus
	)

	if req.ID == 0 {
		versionItem, _ := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))

		name := fmt.Sprintf("1panel_%s_%s_%s", versionItem.Value, loadOs(), timeNow)
		if isCronjob {
			name = fmt.Sprintf("snapshot_1panel_%s_%s_%s", versionItem.Value, loadOs(), timeNow)
		}
		rootDir = path.Join(localDir, "system", name)

		snap = model.Snapshot{
			Name:            name,
			Description:     req.Description,
			From:            req.From,
			DefaultDownload: req.DefaultDownload,
			Version:         versionItem.Value,
			Status:          constant.StatusWaiting,
		}
		_ = snapshotRepo.Create(&snap)
		snapStatus.SnapID = snap.ID
		_ = snapshotRepo.CreateStatus(&snapStatus)
	} else {
		snap, err = snapshotRepo.Get(commonRepo.WithByID(req.ID))
		if err != nil {
			return "", err
		}
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusWaiting})
		snapStatus, _ = snapshotRepo.GetStatus(snap.ID)
		if snapStatus.ID == 0 {
			snapStatus.SnapID = snap.ID
			_ = snapshotRepo.CreateStatus(&snapStatus)
		}
		rootDir = path.Join(localDir, fmt.Sprintf("system/%s", snap.Name))
	}

	var wg sync.WaitGroup
	itemHelper := snapHelper{SnapID: snap.ID, Status: &snapStatus, Wg: &wg, FileOp: files.NewFileOp(), Ctx: context.Background()}
	backupPanelDir := path.Join(rootDir, "1panel")
	_ = os.MkdirAll(backupPanelDir, os.ModePerm)
	backupDockerDir := path.Join(rootDir, "docker")
	_ = os.MkdirAll(backupDockerDir, os.ModePerm)

	jsonItem := SnapshotJson{
		BaseDir:       global.CONF.System.BaseDir,
		BackupDataDir: localDir,
		PanelDataDir:  path.Join(global.CONF.System.BaseDir, "1panel"),
	}
	loadLogByStatus(snapStatus, logPath)
	if snapStatus.PanelInfo != constant.StatusDone {
		wg.Add(1)
		go snapJson(itemHelper, jsonItem, rootDir)
	}
	if snapStatus.Panel != constant.StatusDone {
		wg.Add(1)
		go snapPanel(itemHelper, backupPanelDir)
	}
	if snapStatus.DaemonJson != constant.StatusDone {
		wg.Add(1)
		go snapDaemonJson(itemHelper, backupDockerDir)
	}
	if snapStatus.AppData != constant.StatusDone {
		wg.Add(1)
		go snapAppData(itemHelper, backupDockerDir)
	}
	if snapStatus.BackupData != constant.StatusDone {
		wg.Add(1)
		go snapBackup(itemHelper, localDir, backupPanelDir)
	}

	if !isCronjob {
		go func() {
			wg.Wait()
			if !checkIsAllDone(snap.ID) {
				_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
				return
			}
			if snapStatus.PanelData != constant.StatusDone {
				snapPanelData(itemHelper, localDir, backupPanelDir)
			}
			if snapStatus.PanelData != constant.StatusDone {
				_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
				return
			}
			if snapStatus.Compress != constant.StatusDone {
				snapCompress(itemHelper, rootDir, secret)
			}
			if snapStatus.Compress != constant.StatusDone {
				_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
				return
			}
			if snapStatus.Upload != constant.StatusDone {
				snapUpload(itemHelper, req.From, fmt.Sprintf("%s.tar.gz", rootDir))
			}
			if snapStatus.Upload != constant.StatusDone {
				_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
				return
			}
			_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusSuccess})
		}()
		return "", nil
	}
	wg.Wait()
	if !checkIsAllDone(snap.ID) {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
		loadLogByStatus(snapStatus, logPath)
		return snap.Name, fmt.Errorf("snapshot %s backup failed", snap.Name)
	}
	loadLogByStatus(snapStatus, logPath)
	snapPanelData(itemHelper, localDir, backupPanelDir)
	if snapStatus.PanelData != constant.StatusDone {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
		loadLogByStatus(snapStatus, logPath)
		return snap.Name, fmt.Errorf("snapshot %s 1panel data failed", snap.Name)
	}
	loadLogByStatus(snapStatus, logPath)
	snapCompress(itemHelper, rootDir, secret)
	if snapStatus.Compress != constant.StatusDone {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
		loadLogByStatus(snapStatus, logPath)
		return snap.Name, fmt.Errorf("snapshot %s compress failed", snap.Name)
	}
	loadLogByStatus(snapStatus, logPath)
	snapUpload(itemHelper, req.From, fmt.Sprintf("%s.tar.gz", rootDir))
	if snapStatus.Upload != constant.StatusDone {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
		loadLogByStatus(snapStatus, logPath)
		return snap.Name, fmt.Errorf("snapshot %s upload failed", snap.Name)
	}
	_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusSuccess})
	loadLogByStatus(snapStatus, logPath)
	return snap.Name, nil
}

func (u *SnapshotService) Delete(req dto.SnapshotBatchDelete) error {
	snaps, _ := snapshotRepo.GetList(commonRepo.WithIdsIn(req.Ids))
	for _, snap := range snaps {
		if req.DeleteWithFile {
			targetAccounts, err := loadClientMap(snap.From)
			if err != nil {
				return err
			}
			for _, item := range targetAccounts {
				global.LOG.Debugf("remove snapshot file %s.tar.gz from %s", snap.Name, item.backType)
				_, _ = item.client.Delete(path.Join(item.backupPath, "system_snapshot", snap.Name+".tar.gz"))
			}
		}

		_ = snapshotRepo.DeleteStatus(snap.ID)
		if err := snapshotRepo.Delete(commonRepo.WithByID(snap.ID)); err != nil {
			return err
		}
	}
	return nil
}

func updateRecoverStatus(id uint, isRecover bool, interruptStep, status, message string) {
	if isRecover {
		if status != constant.StatusSuccess {
			global.LOG.Errorf("recover failed, err: %s", message)
		}
		if err := snapshotRepo.Update(id, map[string]interface{}{
			"interrupt_step":    interruptStep,
			"recover_status":    status,
			"recover_message":   message,
			"last_recovered_at": time.Now().Format(constant.DateTimeLayout),
		}); err != nil {
			global.LOG.Errorf("update snap recover status failed, err: %v", err)
		}
		_ = settingRepo.Update("SystemStatus", "Free")
		return
	}
	_ = settingRepo.Update("SystemStatus", "Free")
	if status == constant.StatusSuccess {
		if err := snapshotRepo.Update(id, map[string]interface{}{
			"recover_status":     "",
			"recover_message":    "",
			"interrupt_step":     "",
			"rollback_status":    "",
			"rollback_message":   "",
			"last_rollbacked_at": time.Now().Format(constant.DateTimeLayout),
		}); err != nil {
			global.LOG.Errorf("update snap recover status failed, err: %v", err)
		}
		return
	}
	global.LOG.Errorf("rollback failed, err: %s", message)
	if err := snapshotRepo.Update(id, map[string]interface{}{
		"rollback_status":    status,
		"rollback_message":   message,
		"last_rollbacked_at": time.Now().Format(constant.DateTimeLayout),
	}); err != nil {
		global.LOG.Errorf("update snap recover status failed, err: %v", err)
	}
}

func (u *SnapshotService) handleUnTar(sourceDir, targetDir string, secret string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}
	commands := ""
	if len(secret) != 0 {
		extraCmd := "openssl enc -d -aes-256-cbc -k '" + secret + "' -in " + sourceDir + " | "
		commands = fmt.Sprintf("%s tar -zxvf - -C %s", extraCmd, targetDir+" > /dev/null 2>&1")
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar zxvfC %s %s", sourceDir, targetDir)
		global.LOG.Debug(commands)
	}
	stdout, err := cmd.ExecWithTimeOut(commands, 30*time.Minute)
	if err != nil {
		if len(stdout) != 0 {
			global.LOG.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
			return fmt.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
		}
	}
	return nil
}

func rebuildAllAppInstall() error {
	global.LOG.Debug("start to rebuild all app")
	appInstalls, err := appInstallRepo.ListBy()
	if err != nil {
		global.LOG.Errorf("get all app installed for rebuild failed, err: %v", err)
		return err
	}
	var wg sync.WaitGroup
	for i := 0; i < len(appInstalls); i++ {
		wg.Add(1)
		appInstalls[i].Status = constant.Rebuilding
		_ = appInstallRepo.Save(context.Background(), &appInstalls[i])
		go func(app model.AppInstall) {
			defer wg.Done()
			dockerComposePath := app.GetComposePath()
			_, _ = compose.Up(dockerComposePath)
			app.Status = constant.Running
			_ = appInstallRepo.Save(context.Background(), &app)
		}(appInstalls[i])
	}
	wg.Wait()
	return nil
}

func checkIsAllDone(snapID uint) bool {
	status, err := snapshotRepo.GetStatus(snapID)
	if err != nil {
		return false
	}
	isOK, _ := checkAllDone(status)
	return isOK
}

func checkAllDone(status model.SnapshotStatus) (bool, string) {
	if status.Panel != constant.StatusDone {
		return false, status.Panel
	}
	if status.PanelInfo != constant.StatusDone {
		return false, status.PanelInfo
	}
	if status.DaemonJson != constant.StatusDone {
		return false, status.DaemonJson
	}
	if status.AppData != constant.StatusDone {
		return false, status.AppData
	}
	if status.BackupData != constant.StatusDone {
		return false, status.BackupData
	}
	return true, ""
}

func loadLogByStatus(status model.SnapshotStatus, logPath string) {
	logs := ""
	logs += fmt.Sprintf("Write 1Panel basic information: %s \n", status.PanelInfo)
	logs += fmt.Sprintf("Backup 1Panel system files: %s \n", status.Panel)
	logs += fmt.Sprintf("Backup Docker configuration file: %s \n", status.DaemonJson)
	logs += fmt.Sprintf("Backup installed apps from 1Panel: %s \n", status.AppData)
	logs += fmt.Sprintf("Backup 1Panel data directory: %s \n", status.PanelData)
	logs += fmt.Sprintf("Backup local backup directory for 1Panel: %s \n", status.BackupData)
	logs += fmt.Sprintf("Create snapshot file: %s \n", status.Compress)
	logs += fmt.Sprintf("Snapshot size: %s \n", status.Size)
	logs += fmt.Sprintf("Upload snapshot file: %s \n", status.Upload)

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write([]byte(logs))
}

func hasOs(name string) bool {
	return strings.Contains(name, "amd64") ||
		strings.Contains(name, "arm64") ||
		strings.Contains(name, "armv7") ||
		strings.Contains(name, "ppc64le") ||
		strings.Contains(name, "s390x")
}

func loadOs() string {
	hostInfo, _ := host.Info()
	switch hostInfo.KernelArch {
	case "x86_64":
		return "amd64"
	case "armv7l":
		return "armv7"
	default:
		return hostInfo.KernelArch
	}
}

func loadSnapSize(records []model.Snapshot) ([]dto.SnapshotFile, error) {
	var datas []dto.SnapshotFile
	clientMap := make(map[string]loadSizeHelper)
	var wg sync.WaitGroup
	for i := 0; i < len(records); i++ {
		item := dto.SnapshotFile{Name: records[i].Name, ID: records[i].ID}
		itemPath := fmt.Sprintf("system_snapshot/%s.tar.gz", item.Name)
		if _, ok := clientMap[records[i].DefaultDownload]; !ok {
			backup, err := backupRepo.Get(commonRepo.WithByType(records[i].DefaultDownload))
			if err != nil {
				global.LOG.Errorf("load backup model %s from db failed, err: %v", records[i].DefaultDownload, err)
				clientMap[records[i].DefaultDownload] = loadSizeHelper{}
				datas = append(datas, item)
				continue
			}
			client, err := NewIBackupService().NewClient(&backup)
			if err != nil {
				global.LOG.Errorf("load backup client %s from db failed, err: %v", records[i].DefaultDownload, err)
				clientMap[records[i].DefaultDownload] = loadSizeHelper{}
				datas = append(datas, item)
				continue
			}
			item.Size, _ = client.Size(path.Join(strings.TrimLeft(backup.BackupPath, "/"), itemPath))
			datas = append(datas, item)
			clientMap[records[i].DefaultDownload] = loadSizeHelper{backupPath: strings.TrimLeft(backup.BackupPath, "/"), client: client, isOk: true}
			continue
		}
		if clientMap[records[i].DefaultDownload].isOk {
			wg.Add(1)
			go func(index int) {
				item.Size, _ = clientMap[records[index].DefaultDownload].client.Size(path.Join(clientMap[records[index].DefaultDownload].backupPath, itemPath))
				datas = append(datas, item)
				wg.Done()
			}(i)
		} else {
			datas = append(datas, item)
		}
	}
	wg.Wait()
	return datas, nil
}
