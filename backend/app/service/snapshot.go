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
	SearchWithPage(req dto.SearchWithPage) (int64, interface{}, error)
	SnapshotCreate(req dto.SnapshotCreate) error
	SnapshotRecover(req dto.SnapshotRecover) error
	SnapshotRollback(req dto.SnapshotRecover) error
	SnapshotImport(req dto.SnapshotImport) error
	Delete(req dto.BatchDeleteReq) error

	LoadSnapShotStatus(id uint) (*dto.SnapshotStatus, error)

	UpdateDescription(req dto.UpdateDescription) error
	readFromJson(path string) (SnapshotJson, error)

	HandleSnapshot(isCronjob bool, logPath string, req dto.SnapshotCreate, timeNow string) (string, error)
}

func NewISnapshotService() ISnapshotService {
	return &SnapshotService{}
}

func (u *SnapshotService) SearchWithPage(req dto.SearchWithPage) (int64, interface{}, error) {
	total, systemBackups, err := snapshotRepo.Page(req.Page, req.PageSize, commonRepo.WithLikeName(req.Info))
	var dtoSnap []dto.SnapshotInfo
	for _, systemBackup := range systemBackups {
		var item dto.SnapshotInfo
		if err := copier.Copy(&item, &systemBackup); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoSnap = append(dtoSnap, item)
	}
	return total, dtoSnap, err
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
	if _, err := u.HandleSnapshot(false, "", req, time.Now().Format("20060102150405")); err != nil {
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
		return fmt.Errorf("Restoring snapshots(%s) between different server architectures(%s) is not supported.", snap.Name, loadOs())
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

func (u *SnapshotService) HandleSnapshot(isCronjob bool, logPath string, req dto.SnapshotCreate, timeNow string) (string, error) {
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
				snapCompress(itemHelper, rootDir)
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
	snapCompress(itemHelper, rootDir)
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

func (u *SnapshotService) Delete(req dto.BatchDeleteReq) error {
	backups, _ := snapshotRepo.GetList(commonRepo.WithIdsIn(req.Ids))
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	for _, snap := range backups {
		itemFile := path.Join(localDir, "system", snap.Name)
		_ = os.RemoveAll(itemFile)

		itemTarFile := path.Join(global.CONF.System.TmpDir, "system", snap.Name+".tar.gz")
		_ = os.Remove(itemTarFile)

		_ = snapshotRepo.DeleteStatus(snap.ID)
	}
	if err := snapshotRepo.Delete(commonRepo.WithIdsIn(req.Ids)); err != nil {
		return err
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
			"last_recovered_at": time.Now().Format("2006-01-02 15:04:05"),
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
			"last_rollbacked_at": time.Now().Format("2006-01-02 15:04:05"),
		}); err != nil {
			global.LOG.Errorf("update snap recover status failed, err: %v", err)
		}
		return
	}
	global.LOG.Errorf("rollback failed, err: %s", message)
	if err := snapshotRepo.Update(id, map[string]interface{}{
		"rollback_status":    status,
		"rollback_message":   message,
		"last_rollbacked_at": time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {
		global.LOG.Errorf("update snap recover status failed, err: %v", err)
	}
}

func cpBinary(src []string, dst string) error {
	global.LOG.Debugf(fmt.Sprintf("\\cp -f %s %s", strings.Join(src, " "), dst))
	stdout, err := cmd.Exec(fmt.Sprintf("\\cp -f %s %s", strings.Join(src, " "), dst))
	if err != nil {
		return fmt.Errorf("cp file failed, stdout: %v, err: %v", stdout, err)
	}
	return nil
}

func (u *SnapshotService) handleUnTar(sourceDir, targetDir string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	commands := fmt.Sprintf("tar -zxf %s -C %s .", sourceDir, targetDir)
	global.LOG.Debug(commands)
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
			out, err := compose.Down(dockerComposePath)
			if err != nil {
				_ = handleErr(app, err, out)
				return
			}
			out, err = compose.Up(dockerComposePath)
			if err != nil {
				_ = handleErr(app, err, out)
				return
			}
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
