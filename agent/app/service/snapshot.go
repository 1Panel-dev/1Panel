package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
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

	UpdateDescription(req dto.UpdateDescription) error

	HandleSnapshot(isCronjob bool, req dto.SnapshotCreate, timeNow string, secret string) (string, error)
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
		}
		if item.Label == "system_snapshot" {
			itemBackups[i].IsCheck = false
			for j := 0; j < len(item.Children); j++ {
				itemBackups[i].Children[j].IsCheck = false
			}
		}
	}

	return data, nil
}

func (u *SnapshotService) UpdateDescription(req dto.UpdateDescription) error {
	return snapshotRepo.Update(req.ID, map[string]interface{}{"description": req.Description})
}

type SnapshotJson struct {
	BaseDir       string `json:"baseDir"`
	BackupDataDir string `json:"backupDataDir"`
	Size          uint64 `json:"size"`
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

		if err := snapshotRepo.Delete(commonRepo.WithByID(snap.ID)); err != nil {
			return err
		}
	}
	return nil
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
	openrestyID := 0
	for _, app := range apps {
		if app.App.Key == constant.AppOpenresty {
			openrestyID = int(app.ID)
		}
	}
	websites, err := websiteRepo.List()
	if err != nil {
		return data, err
	}
	appRelationMap := make(map[uint]uint)
	for _, website := range websites {
		if website.Type == constant.Deployment && website.AppInstallID != 0 {
			appRelationMap[uint(openrestyID)] = website.AppInstallID
		}
	}
	appRelations, err := appInstallResourceRepo.GetBy()
	if err != nil {
		return data, err
	}
	for _, relation := range appRelations {
		appRelationMap[uint(relation.AppInstallId)] = relation.LinkId
	}
	appMap := make(map[uint]string)
	for _, app := range apps {
		appMap[app.ID] = fmt.Sprintf("%s-%s", app.App.Key, app.Name)
	}

	appTreeMap := make(map[string]dto.DataTree)
	for _, app := range apps {
		itemApp := dto.DataTree{
			ID:    uuid.NewString(),
			Label: fmt.Sprintf("%s - %s", app.App.Name, app.Name),
			Key:   app.App.Key,
			Name:  app.Name,
		}
		appPath := path.Join(global.CONF.System.BaseDir, "1panel/apps", app.App.Key, app.Name)
		itemAppData := dto.DataTree{ID: uuid.NewString(), Label: "appData", Key: app.App.Key, Name: app.Name, IsCheck: true, Path: appPath}
		if app.App.Key == constant.AppOpenresty && len(websites) != 0 {
			itemAppData.IsDisable = true
		}
		if val, ok := appRelationMap[app.ID]; ok {
			itemAppData.RelationItemID = appMap[val]
		}
		sizeItem, err := fileOp.GetDirSize(appPath)
		if err == nil {
			itemAppData.Size = uint64(sizeItem)
		}
		itemApp.Size += itemAppData.Size
		data = append(data, itemApp)
		appTreeMap[fmt.Sprintf("%s-%s", itemApp.Key, itemApp.Name)] = itemAppData
	}

	for key, val := range appTreeMap {
		if valRelation, ok := appTreeMap[val.RelationItemID]; ok {
			valRelation.IsDisable = true
			appTreeMap[val.RelationItemID] = valRelation

			val.RelationItemID = valRelation.ID
			appTreeMap[key] = val
		}
	}
	for i := 0; i < len(data); i++ {
		if val, ok := appTreeMap[fmt.Sprintf("%s-%s", data[i].Key, data[i].Name)]; ok {
			data[i].Children = append(data[i].Children, val)
		}
	}
	data = loadAppBackup(data, fileOp)
	data = loadAppImage(data)
	return data, nil
}
func loadAppBackup(list []dto.DataTree, fileOp fileUtils.FileOp) []dto.DataTree {
	for i := 0; i < len(list); i++ {
		appBackupPath := path.Join(global.CONF.System.BaseDir, "1panel/backup/app", list[i].Key, list[i].Name)
		itemAppBackupTree, err := loadFile(appBackupPath, 8, fileOp)
		itemAppBackup := dto.DataTree{ID: uuid.NewString(), Label: "appBackup", IsCheck: true, Children: itemAppBackupTree, Path: appBackupPath}
		if err == nil {
			backupSizeItem, err := fileOp.GetDirSize(appBackupPath)
			if err == nil {
				itemAppBackup.Size = uint64(backupSizeItem)
				list[i].Size += itemAppBackup.Size
			}
			list[i].Children = append(list[i].Children, itemAppBackup)
		}
	}
	return list
}
func loadAppImage(list []dto.DataTree) []dto.DataTree {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("new docker client failed, err: %v", err)
		return list
	}
	defer client.Close()
	imageList, err := client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		global.LOG.Errorf("load image list failed, err: %v", err)
		return list
	}

	for i := 0; i < len(list); i++ {
		itemAppImage := dto.DataTree{ID: uuid.NewString(), Label: "appImage"}
		stdout, err := cmd.Execf("cat %s | grep image: ", path.Join(global.CONF.System.BaseDir, "1panel/apps", list[i].Key, list[i].Name, "docker-compose.yml"))
		if err != nil {
			list[i].Children = append(list[i].Children, itemAppImage)
			continue
		}
		itemAppImage.Name = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(stdout), "\n", ""), "image: ", "")
		for _, imageItem := range imageList {
			for _, tag := range imageItem.RepoTags {
				if tag == itemAppImage.Name {
					itemAppImage.Size = uint64(imageItem.Size)
					break
				}
			}
		}
		list[i].Children = append(list[i].Children, itemAppImage)
	}
	return list
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
		case "agent", "conf", "runtime", "docker", "secret", "task":
			itemData.IsDisable = true
		case "clamav":
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
