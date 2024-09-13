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

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	fileUtils "github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/host"
)

type SnapshotService struct {
	OriginalPath string
}

type ISnapshotService interface {
	SearchWithPage(req dto.SearchWithPage) (int64, interface{}, error)
	LoadSnapshotData() (dto.SnapshotData, error)
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

func (u *SnapshotService) SearchWithPage(req dto.SearchWithPage) (int64, interface{}, error) {
	total, systemBackups, err := snapshotRepo.Page(req.Page, req.PageSize, commonRepo.WithByLikeName(req.Info))
	if err != nil {
		return 0, nil, err
	}
	dtoSnap, err := loadSnapSize(systemBackups)
	if err != nil {
		return 0, nil, err
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
			Name:              snap,
			SourceAccountIDs:  fmt.Sprintf("%v", req.BackupAccountID),
			DownloadAccountID: req.BackupAccountID,
			Version:           nameItems[1],
			Description:       req.Description,
			Status:            constant.StatusSuccess,
		}
		if err := snapshotRepo.Create(&itemSnap); err != nil {
			return err
		}
	}
	return nil
}

func (u *SnapshotService) LoadSnapshotData() (dto.SnapshotData, error) {
	var (
		data dto.SnapshotData
		err  error
	)
	fileOp := fileUtils.NewFileOp()
	data.AppData, err = loadApps(fileOp)
	if err != nil {
		return data, err
	}
	data.PanelData, err = loadPanelFile(fileOp)
	if err != nil {
		return data, err
	}
	itemBackups, err := loadFile(global.CONF.System.Backup, 8, fileOp)
	if err != nil {
		return data, err
	}
	for i, item := range itemBackups {
		if item.Label == "app" {
			data.BackupData = append(itemBackups[:i], itemBackups[i+1:]...)
			break
		}
	}
	data.WithLoginLog = true
	data.WithOperationLog = true
	data.WithMonitorData = false

	return data, nil
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
	BaseDir       string `json:"baseDir"`
	BackupDataDir string `json:"backupDataDir"`
	Size          uint64 `json:"size"`
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
	go u.HandleSnapshotRecover(snap, req)
	return nil
}

func (u *SnapshotService) SnapshotRollback(req dto.SnapshotRecover) error {
	global.LOG.Info("start to rollback now")
	snap, err := snapshotRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	go func() {
		if err := handleRollback(snap.Name); err != nil {
			global.LOG.Errorf("handle roll back snapshot failed, err: %v", err)
		}
	}()
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
	var (
		rootDir    string
		snap       model.Snapshot
		snapStatus model.SnapshotStatus
		err        error
	)

	if req.ID == 0 {
		versionItem, _ := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))

		name := fmt.Sprintf("1panel_%s_%s_%s", versionItem.Value, loadOs(), timeNow)
		if isCronjob {
			name = fmt.Sprintf("snapshot_1panel_%s_%s_%s", versionItem.Value, loadOs(), timeNow)
		}
		rootDir = path.Join(global.CONF.System.BaseDir, "1panel/tmp/system", name)

		appItem, _ := json.Marshal(req.AppData)
		panelItem, _ := json.Marshal(req.PanelData)
		backupItem, _ := json.Marshal(req.BackupData)
		snap = model.Snapshot{
			Name:              name,
			Description:       req.Description,
			SourceAccountIDs:  req.SourceAccountIDs,
			DownloadAccountID: req.DownloadAccountID,

			AppData:          string(appItem),
			PanelData:        string(panelItem),
			BackupData:       string(backupItem),
			WithMonitorData:  req.WithMonitorData,
			WithLoginLog:     req.WithLoginLog,
			WithOperationLog: req.WithOperationLog,

			Version: versionItem.Value,
			Status:  constant.StatusWaiting,
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
		rootDir = path.Join(global.CONF.System.BaseDir, "1panel/tmp/system", snap.Name)
	}

	var wg sync.WaitGroup
	itemHelper := snapHelper{SnapID: snap.ID, Status: &snapStatus, Wg: &wg, FileOp: files.NewFileOp(), Ctx: context.Background()}
	baseDir := path.Join(rootDir, "base")
	_ = os.MkdirAll(baseDir, os.ModePerm)
	if err := loadDbConn(&itemHelper, rootDir, req); err != nil {
		return "", fmt.Errorf("load snapshot db conn failed, err: %v", err)
	}

	loadLogByStatus(snapStatus, logPath)
	if snapStatus.BaseData != constant.StatusDone {
		wg.Add(1)
		go snapBaseData(itemHelper, baseDir)
	}
	if snapStatus.AppImage != constant.StatusDone {
		wg.Add(1)
		go snapAppImage(itemHelper, req, rootDir)
	}
	if snapStatus.BackupData != constant.StatusDone {
		wg.Add(1)
		go snapBackupData(itemHelper, req, rootDir)
	}
	if snapStatus.PanelData != constant.StatusDone {
		wg.Add(1)
		go snapPanelData(itemHelper, req, rootDir)
	}
	if !isCronjob {
		go func() {
			wg.Wait()
			closeDatabase(itemHelper.snapAgentDB)
			closeDatabase(itemHelper.snapCoreDB)
			if !checkIsAllDone(snap.ID) {
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
				snapUpload(itemHelper, req.SourceAccountIDs, fmt.Sprintf("%s.tar.gz", rootDir))
			}
			if snapStatus.Upload != constant.StatusDone {
				_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
				return
			}
			_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusSuccess})
			// _ = snapshotRepo.DeleteStatus(itemHelper.SnapID)
			_ = os.RemoveAll(rootDir)
		}()
		return "", nil
	}
	wg.Wait()
	closeDatabase(itemHelper.snapAgentDB)
	closeDatabase(itemHelper.snapCoreDB)
	if !checkIsAllDone(snap.ID) {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
		loadLogByStatus(snapStatus, logPath)
		return snap.Name, fmt.Errorf("snapshot %s backup failed", snap.Name)
	}
	loadLogByStatus(snapStatus, logPath)
	snapCompress(itemHelper, rootDir, secret)
	if snapStatus.Compress != constant.StatusDone {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
		loadLogByStatus(snapStatus, logPath)
		return snap.Name, fmt.Errorf("snapshot %s compress failed", snap.Name)
	}
	loadLogByStatus(snapStatus, logPath)
	snapUpload(itemHelper, req.SourceAccountIDs, fmt.Sprintf("%s.tar.gz", rootDir))
	if snapStatus.Upload != constant.StatusDone {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed})
		loadLogByStatus(snapStatus, logPath)
		return snap.Name, fmt.Errorf("snapshot %s upload failed", snap.Name)
	}
	_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusSuccess})
	loadLogByStatus(snapStatus, logPath)
	_ = os.RemoveAll(rootDir)
	return snap.Name, nil
}

func (u *SnapshotService) Delete(req dto.SnapshotBatchDelete) error {
	snaps, _ := snapshotRepo.GetList(commonRepo.WithByIDs(req.Ids))
	for _, snap := range snaps {
		if req.DeleteWithFile {
			accounts, err := NewBackupClientMap(strings.Split(snap.SourceAccountIDs, ","))
			if err != nil {
				return err
			}
			for _, item := range accounts {
				global.LOG.Debugf("remove snapshot file %s.tar.gz from %s", snap.Name, item.name)
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
	if status.BaseData != constant.StatusDone {
		return false, status.BaseData
	}
	if status.PanelData != constant.StatusDone {
		return false, status.PanelData
	}
	if status.AppImage != constant.StatusDone {
		return false, status.AppImage
	}
	if status.BackupData != constant.StatusDone {
		return false, status.BackupData
	}
	return true, ""
}

func loadLogByStatus(status model.SnapshotStatus, logPath string) {
	logs := ""
	logs += fmt.Sprintf("Backup 1Panel base files: %s \n", status.BaseData)
	logs += fmt.Sprintf("Backup installed apps from 1Panel: %s \n", status.AppImage)
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

func loadSnapSize(records []model.Snapshot) ([]dto.SnapshotInfo, error) {
	var datas []dto.SnapshotInfo
	clientMap := make(map[uint]loadSizeHelper)
	var wg sync.WaitGroup
	for i := 0; i < len(records); i++ {
		var item dto.SnapshotInfo
		if err := copier.Copy(&item, &records[i]); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		itemPath := fmt.Sprintf("system_snapshot/%s.tar.gz", item.Name)
		if _, ok := clientMap[records[i].DownloadAccountID]; !ok {
			backup, client, err := NewBackupClientWithID(records[i].DownloadAccountID)
			if err != nil {
				global.LOG.Errorf("load backup client from db failed, err: %v", err)
				clientMap[records[i].DownloadAccountID] = loadSizeHelper{}
				datas = append(datas, item)
				continue
			}
			item.Size, _ = client.Size(path.Join(strings.TrimLeft(backup.BackupPath, "/"), itemPath))
			datas = append(datas, item)
			clientMap[records[i].DownloadAccountID] = loadSizeHelper{backupPath: strings.TrimLeft(backup.BackupPath, "/"), client: client, isOk: true}
			continue
		}
		if clientMap[records[i].DownloadAccountID].isOk {
			wg.Add(1)
			go func(index int) {
				item.Size, _ = clientMap[records[index].DownloadAccountID].client.Size(path.Join(clientMap[records[index].DownloadAccountID].backupPath, itemPath))
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

func loadApps(fileOp fileUtils.FileOp) ([]dto.DataTree, error) {
	var data []dto.DataTree
	apps, err := appInstallRepo.ListBy()
	if err != nil {
		return data, err
	}
	client, err := docker.NewDockerClient()
	hasDockerClient := true
	if err != nil {
		hasDockerClient = false
		global.LOG.Errorf("new docker client failed, err: %v", err)
	} else {
		defer client.Close()
	}
	imageList, err := client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		hasDockerClient = false
		global.LOG.Errorf("load image list failed, err: %v", err)
	}

	for _, app := range apps {
		itemApp := dto.DataTree{ID: uuid.NewString(), Label: fmt.Sprintf("%s - %s", app.App.Name, app.Name), Key: app.App.Key, Name: app.Name}
		appPath := path.Join(global.CONF.System.BaseDir, "1panel/apps", app.App.Key, app.Name)
		itemAppData := dto.DataTree{ID: uuid.NewString(), Label: "appData", Key: app.App.Key, Name: app.Name, IsCheck: true, Path: appPath}
		sizeItem, err := fileOp.GetDirSize(appPath)
		if err == nil {
			itemAppData.Size = uint64(sizeItem)
		}
		itemApp.Size += itemAppData.Size
		itemApp.Children = append(itemApp.Children, itemAppData)

		appBackupPath := path.Join(global.CONF.System.BaseDir, "1panel/backup/app", app.App.Key, app.Name)
		itemAppBackupTree, err := loadFile(appBackupPath, 8, fileOp)
		itemAppBackup := dto.DataTree{ID: uuid.NewString(), Label: "appBackup", IsCheck: true, Children: itemAppBackupTree, Path: appBackupPath}
		if err == nil {
			backupSizeItem, err := fileOp.GetDirSize(appBackupPath)
			if err == nil {
				itemAppBackup.Size = uint64(backupSizeItem)
				itemApp.Size += itemAppBackup.Size
			}
			itemApp.Children = append(itemApp.Children, itemAppBackup)
		}

		itemAppImage := dto.DataTree{ID: uuid.NewString(), Label: "appImage"}
		stdout, err := cmd.Execf("cat %s | grep image: ", path.Join(global.CONF.System.BaseDir, "1panel/apps", app.App.Key, app.Name, "docker-compose.yml"))
		if err != nil {
			itemApp.Children = append(itemApp.Children, itemAppImage)
			data = append(data, itemApp)
			continue
		}
		itemAppImage.Name = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(stdout), "\n", ""), "image: ", "")
		if !hasDockerClient {
			itemApp.Children = append(itemApp.Children, itemAppImage)
			data = append(data, itemApp)
			continue
		}
		for _, imageItem := range imageList {
			for _, tag := range imageItem.RepoTags {
				if tag == itemAppImage.Name {
					itemAppImage.Size = uint64(imageItem.Size)
					break
				}
			}
		}
		itemApp.Children = append(itemApp.Children, itemAppImage)
		data = append(data, itemApp)
	}
	return data, nil
}

func loadPanelFile(fileOp fileUtils.FileOp) ([]dto.DataTree, error) {
	var data []dto.DataTree
	snapFiles, err := os.ReadDir(path.Join(global.CONF.System.BaseDir, "1panel"))
	if err != nil {
		return data, err
	}
	for _, fileItem := range snapFiles {
		itemData := dto.DataTree{
			ID:      uuid.NewString(),
			Label:   fileItem.Name(),
			IsCheck: true,
			Path:    path.Join(global.CONF.System.BaseDir, "1panel", fileItem.Name()),
		}
		switch itemData.Label {
		case "agent", "conf", "db", "runtime", "secret":
			itemData.IsDisable = true
		case "log", "docker", "task", "clamav":
			panelPath := path.Join(global.CONF.System.BaseDir, "1panel", itemData.Label)
			itemData.Children, _ = loadFile(panelPath, 5, fileOp)
		default:
			continue
		}
		if fileItem.IsDir() {
			sizeItem, err := fileOp.GetDirSize(path.Join(global.CONF.System.BaseDir, "1panel", itemData.Label))
			if err != nil {
				continue
			}
			itemData.Size = uint64(sizeItem)
		} else {
			fileInfo, err := fileItem.Info()
			if err != nil {
				continue
			}
			itemData.Size = uint64(fileInfo.Size())
		}
		if itemData.IsCheck && itemData.Size == 0 {
			itemData.IsCheck = false
			itemData.IsDisable = true
		}

		data = append(data, itemData)
	}

	return data, nil
}

func loadFile(pathItem string, index int, fileOp fileUtils.FileOp) ([]dto.DataTree, error) {
	var data []dto.DataTree
	snapFiles, err := os.ReadDir(pathItem)
	if err != nil {
		return data, err
	}
	i := 0
	for _, fileItem := range snapFiles {
		itemData := dto.DataTree{
			ID:      uuid.NewString(),
			Label:   fileItem.Name(),
			Name:    fileItem.Name(),
			Path:    path.Join(pathItem, fileItem.Name()),
			IsCheck: true,
		}
		if fileItem.IsDir() {
			sizeItem, err := fileOp.GetDirSize(path.Join(pathItem, itemData.Label))
			if err != nil {
				continue
			}
			itemData.Size = uint64(sizeItem)
			itemData.Children, _ = loadFile(path.Join(pathItem, itemData.Label), index-1, fileOp)
		} else {
			fileInfo, err := fileItem.Info()
			if err != nil {
				continue
			}
			itemData.Size = uint64(fileInfo.Size())
		}
		data = append(data, itemData)
		i++
	}
	return data, nil
}
