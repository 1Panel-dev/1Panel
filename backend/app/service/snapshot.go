package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
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

	UpdateDescription(req dto.UpdateDescription) error
	readFromJson(path string) (SnapshotJson, error)
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
		nameItems := strings.Split(snap, "_")
		if !strings.HasPrefix(snap, "1panel_v") || !strings.HasSuffix(snap, ".tar.gz") || len(nameItems) != 3 {
			return fmt.Errorf("incorrect snapshot name format of %s", snap)
		}
		formatTime, err := time.Parse("20060102150405", strings.ReplaceAll(nameItems[2], ".tar.gz", ""))
		if err != nil {
			return fmt.Errorf("incorrect snapshot name format of %s", snap)
		}
		if strings.HasSuffix(snap, ".tar.gz") {
			snap = strings.ReplaceAll(snap, ".tar.gz", "")
		}
		itemSnap := model.Snapshot{
			Name:        snap,
			From:        req.From,
			Version:     nameItems[1],
			Description: req.Description,
			Status:      constant.StatusSuccess,
			BaseModel: model.BaseModel{
				CreatedAt: formatTime,
				UpdatedAt: formatTime,
			},
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
	global.LOG.Info("start to create snapshot now")
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	backup, err := backupRepo.Get(commonRepo.WithByType(req.From))
	if err != nil {
		return err
	}
	backupAccount, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return err
	}

	timeNow := time.Now().Format("20060102150405")
	versionItem, _ := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))
	rootDir := fmt.Sprintf("%s/system/1panel_%s_%s", localDir, versionItem.Value, timeNow)
	backupPanelDir := fmt.Sprintf("%s/1panel", rootDir)
	_ = os.MkdirAll(backupPanelDir, os.ModePerm)
	backupDockerDir := fmt.Sprintf("%s/docker", rootDir)
	_ = os.MkdirAll(backupDockerDir, os.ModePerm)

	_ = settingRepo.Update("SystemStatus", "Snapshoting")
	snap := model.Snapshot{
		Name:        fmt.Sprintf("1panel_%s_%s", versionItem.Value, timeNow),
		Description: req.Description,
		From:        req.From,
		Version:     versionItem.Value,
		Status:      constant.StatusSuccess,
	}
	_ = snapshotRepo.Create(&snap)
	go func() {
		defer func() {
			_ = os.RemoveAll(rootDir)
		}()
		fileOp := files.NewFileOp()

		dockerDataDir, liveRestoreStatus, err := u.loadDockerDataDir()
		if err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		_, _ = cmd.Exec("systemctl stop docker")
		if err := u.handleDockerDatas(fileOp, "snapshot", dockerDataDir, backupDockerDir); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if err := u.handleDaemonJson(fileOp, "snapshot", "", backupDockerDir); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}

		if err := u.handlePanelBinary(fileOp, "snapshot", "", backupPanelDir+"/1panel"); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if err := u.handlePanelctlBinary(fileOp, "snapshot", "", backupPanelDir+"/1pctl"); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if err := u.handlePanelService(fileOp, "snapshot", "", backupPanelDir+"/1panel.service"); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}

		if err := u.handleBackupDatas(fileOp, "snapshot", localDir, backupPanelDir); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}

		if err := u.handlePanelDatas(snap.ID, fileOp, "snapshot", global.CONF.System.BaseDir+"/1panel", backupPanelDir, localDir, dockerDataDir); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		_, _ = cmd.Exec("systemctl restart docker")

		snapJson := SnapshotJson{
			BaseDir:            global.CONF.System.BaseDir,
			DockerDataDir:      dockerDataDir,
			BackupDataDir:      localDir,
			PanelDataDir:       global.CONF.System.BaseDir + "/1panel",
			LiveRestoreEnabled: liveRestoreStatus,
		}
		if err := u.saveJson(snapJson, rootDir); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, fmt.Sprintf("save snapshot json failed, err: %v", err))
			return
		}

		if err := handleTar(rootDir, fmt.Sprintf("%s/system", localDir), fmt.Sprintf("1panel_%s_%s.tar.gz", versionItem.Value, timeNow), ""); err != nil {
			updateSnapshotStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}

		_ = settingRepo.Update("SystemStatus", "Free")

		global.LOG.Infof("start to upload snapshot to %s, please wait", backup.Type)
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusUploading})
		localPath := fmt.Sprintf("%s/system/1panel_%s_%s.tar.gz", localDir, versionItem.Value, timeNow)
		itemBackupPath := strings.TrimLeft(backup.BackupPath, "/")
		itemBackupPath = strings.TrimRight(itemBackupPath, "/")
		if ok, err := backupAccount.Upload(localPath, fmt.Sprintf("%s/system_snapshot/1panel_%s_%s.tar.gz", itemBackupPath, versionItem.Value, timeNow)); err != nil || !ok {
			_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			global.LOG.Errorf("upload snapshot to %s failed, err: %v", backup.Type, err)
			return
		}
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusSuccess})
		_ = os.RemoveAll(fmt.Sprintf("%s/system/1panel_%s_%s.tar.gz", localDir, versionItem.Value, timeNow))

		global.LOG.Infof("upload snapshot to %s success", backup.Type)
	}()
	return nil
}

func (u *SnapshotService) SnapshotRecover(req dto.SnapshotRecover) error {
	global.LOG.Info("start to recvover panel by snapshot now")
	snap, err := snapshotRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if !req.IsNew && len(snap.InterruptStep) != 0 && len(snap.RollbackStatus) != 0 {
		return fmt.Errorf("the snapshot has been rolled back and cannot be restored again")
	}
	isReTry := false
	if len(snap.InterruptStep) != 0 && !req.IsNew {
		isReTry = true
	}
	backup, err := backupRepo.Get(commonRepo.WithByType(snap.From))
	if err != nil {
		return err
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return err
	}
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	baseDir := fmt.Sprintf("%s/system/%s", localDir, snap.Name)
	if _, err := os.Stat(baseDir); err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(baseDir, os.ModePerm)
	}

	_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"recover_status": constant.StatusWaiting})
	_ = settingRepo.Update("SystemStatus", "Recovering")
	go func() {
		operation := "recover"
		if isReTry {
			operation = "re-recover"
		}
		if !isReTry || snap.InterruptStep == "Download" || (isReTry && req.ReDownload) {
			itemBackupPath := strings.TrimLeft(backup.BackupPath, "/")
			itemBackupPath = strings.TrimRight(itemBackupPath, "/")
			ok, err := client.Download(fmt.Sprintf("%s/system_snapshot/%s.tar.gz", itemBackupPath, snap.Name), fmt.Sprintf("%s/%s.tar.gz", baseDir, snap.Name))
			if err != nil || !ok {
				if req.ReDownload {
					updateRecoverStatus(snap.ID, snap.InterruptStep, constant.StatusFailed, fmt.Sprintf("download file %s from %s failed, err: %v", snap.Name, backup.Type, err))
					return
				}
				updateRecoverStatus(snap.ID, "Download", constant.StatusFailed, fmt.Sprintf("download file %s from %s failed, err: %v", snap.Name, backup.Type, err))
				return
			}
			isReTry = false
		}
		fileOp := files.NewFileOp()
		if !isReTry || snap.InterruptStep == "Decompress" || (isReTry && req.ReDownload) {
			if err := handleUnTar(fmt.Sprintf("%s/%s.tar.gz", baseDir, snap.Name), baseDir); err != nil {
				if req.ReDownload {
					updateRecoverStatus(snap.ID, snap.InterruptStep, constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
					return
				}
				updateRecoverStatus(snap.ID, "Decompress", constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
				return
			}
			isReTry = false
		}
		rootDir := fmt.Sprintf("%s/%s", baseDir, snap.Name)

		snapJson, err := u.readFromJson(fmt.Sprintf("%s/snapshot.json", rootDir))
		if err != nil {
			updateRecoverStatus(snap.ID, "Readjson", constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
			return
		}
		if snap.InterruptStep == "Readjson" {
			isReTry = false
		}
		u.OriginalPath = fmt.Sprintf("%s/original_%s", snapJson.BaseDir, snap.Name)
		_ = os.MkdirAll(u.OriginalPath, os.ModePerm)

		snapJson.OldBaseDir = global.CONF.System.BaseDir
		snapJson.OldPanelDataDir = global.CONF.System.BaseDir + "/1panel"
		snapJson.OldBackupDataDir = localDir
		recoverPanelDir := fmt.Sprintf("%s/%s/1panel", baseDir, snap.Name)
		liveRestore := false
		if !isReTry || snap.InterruptStep == "LoadDockerJson" {
			snapJson.OldDockerDataDir, liveRestore, err = u.loadDockerDataDir()
			if err != nil {
				updateRecoverStatus(snap.ID, "LoadDockerJson", constant.StatusFailed, fmt.Sprintf("load docker data dir failed, err: %v", err))
				return
			}
			isReTry = false
		}
		if liveRestore {
			if err := u.updateLiveRestore(false); err != nil {
				updateRecoverStatus(snap.ID, "UpdateLiveRestore", constant.StatusFailed, fmt.Sprintf("update docker daemon.json live-restore conf failed, err: %v", err))
				return
			}
			isReTry = false
		}
		_ = u.saveJson(snapJson, rootDir)

		_, _ = cmd.Exec("systemctl stop docker")
		if !isReTry || snap.InterruptStep == "DockerDir" {
			if err := u.handleDockerDatas(fileOp, operation, rootDir, snapJson.DockerDataDir); err != nil {
				updateRecoverStatus(snap.ID, "DockerDir", constant.StatusFailed, err.Error())
				return
			}
			isReTry = false
		}
		if !isReTry || snap.InterruptStep == "DaemonJson" {
			if err := u.handleDaemonJson(fileOp, operation, rootDir+"/docker/daemon.json", u.OriginalPath); err != nil {
				updateRecoverStatus(snap.ID, "DaemonJson", constant.StatusFailed, err.Error())
				return
			}
			isReTry = false
		}
		_, _ = cmd.Exec("systemctl restart docker")

		if !isReTry || snap.InterruptStep == "1PanelBinary" {
			if err := u.handlePanelBinary(fileOp, operation, recoverPanelDir+"/1panel", u.OriginalPath+"/1panel"); err != nil {
				updateRecoverStatus(snap.ID, "1PanelBinary", constant.StatusFailed, err.Error())
				return
			}
			isReTry = false
		}
		if !isReTry || snap.InterruptStep == "1PctlBinary" {
			if err := u.handlePanelctlBinary(fileOp, operation, recoverPanelDir+"/1pctl", u.OriginalPath+"/1pctl"); err != nil {
				updateRecoverStatus(snap.ID, "1PctlBinary", constant.StatusFailed, err.Error())
				return
			}
			isReTry = false
		}
		if !isReTry || snap.InterruptStep == "1PanelService" {
			if err := u.handlePanelService(fileOp, operation, recoverPanelDir+"/1panel.service", u.OriginalPath+"/1panel.service"); err != nil {
				updateRecoverStatus(snap.ID, "1PanelService", constant.StatusFailed, err.Error())
				return
			}
			isReTry = false
		}

		if !isReTry || snap.InterruptStep == "1PanelBackups" {
			if err := u.handleBackupDatas(fileOp, operation, rootDir, snapJson.BackupDataDir); err != nil {
				updateRecoverStatus(snap.ID, "1PanelBackups", constant.StatusFailed, err.Error())
				return
			}
			isReTry = false
		}

		if !isReTry || snap.InterruptStep == "1PanelData" {
			if err := u.handlePanelDatas(snap.ID, fileOp, operation, rootDir, snapJson.PanelDataDir, localDir, snapJson.OldDockerDataDir); err != nil {
				updateRecoverStatus(snap.ID, "1PanelData", constant.StatusFailed, err.Error())
				return
			}
			isReTry = false
		}
		_ = os.RemoveAll(rootDir)
		global.LOG.Info("recover successful")
		_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
	}()
	return nil
}

func (u *SnapshotService) SnapshotRollback(req dto.SnapshotRecover) error {
	global.LOG.Info("start to rollback now")
	snap, err := snapshotRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if snap.InterruptStep == "Download" || snap.InterruptStep == "Decompress" || snap.InterruptStep == "Readjson" {
		return nil
	}
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()

	rootDir := fmt.Sprintf("%s/system/%s/%s", localDir, snap.Name, snap.Name)

	_ = settingRepo.Update("SystemStatus", "Rollbacking")
	_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"rollback_status": constant.StatusWaiting})
	go func() {
		snapJson, err := u.readFromJson(fmt.Sprintf("%s/snapshot.json", rootDir))
		if err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
			return
		}
		u.OriginalPath = fmt.Sprintf("%s/original_%s", snapJson.OldBaseDir, snap.Name)
		if _, err := os.Stat(u.OriginalPath); err != nil && os.IsNotExist(err) {
			return
		}

		_, _ = cmd.Exec("systemctl stop docker")
		if err := u.handleDockerDatas(fileOp, "rollback", u.OriginalPath, snapJson.OldDockerDataDir); err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if snap.InterruptStep == "DockerDir" {
			_, _ = cmd.Exec("systemctl restart docker")
			return
		}

		if err := u.handleDaemonJson(fileOp, "rollback", u.OriginalPath+"/daemon.json", ""); err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if snap.InterruptStep == "DaemonJson" {
			_, _ = cmd.Exec("systemctl restart docker")
			return
		}
		if snapJson.LiveRestoreEnabled {
			if err := u.updateLiveRestore(true); err != nil {
				updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
				return
			}
		}
		if snap.InterruptStep == "UpdateLiveRestore" {
			_, _ = cmd.Exec("systemctl restart dockere")
			return
		}

		if err := u.handlePanelBinary(fileOp, "rollback", u.OriginalPath+"/1panel", ""); err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if snap.InterruptStep == "1PanelBinary" {
			return
		}

		if err := u.handlePanelctlBinary(fileOp, "rollback", u.OriginalPath+"/1pctl", ""); err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if snap.InterruptStep == "1PctlBinary" {
			return
		}

		if err := u.handlePanelService(fileOp, "rollback", u.OriginalPath+"/1panel.service", ""); err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if snap.InterruptStep == "1PanelService" {
			_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
			return
		}

		if err := u.handleBackupDatas(fileOp, "rollback", u.OriginalPath, snapJson.OldBackupDataDir); err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if snap.InterruptStep == "1PanelBackups" {
			_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
			return
		}

		if err := u.handlePanelDatas(snap.ID, fileOp, "rollback", u.OriginalPath, snapJson.OldPanelDataDir, "", ""); err != nil {
			updateRollbackStatus(snap.ID, constant.StatusFailed, err.Error())
			return
		}
		if snap.InterruptStep == "1PanelData" {
			_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
			return
		}

		_ = os.RemoveAll(rootDir)
		global.LOG.Info("rollback successful")
		_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
	}()
	return nil
}

func (u *SnapshotService) saveJson(snapJson SnapshotJson, path string) error {
	remarkInfo, _ := json.MarshalIndent(snapJson, "", "\t")
	if err := os.WriteFile(fmt.Sprintf("%s/snapshot.json", path), remarkInfo, 0640); err != nil {
		return err
	}
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

func (u *SnapshotService) handleDockerDatas(fileOp files.FileOp, operation string, source, target string) error {
	switch operation {
	case "snapshot":
		if err := u.handleTar(source, target, "docker_data.tar.gz", ""); err != nil {
			return fmt.Errorf("backup docker data failed, err: %v", err)
		}
	case "recover":
		if err := u.handleTar(target, u.OriginalPath, "docker_data.tar.gz", ""); err != nil {
			return fmt.Errorf("backup docker data failed, err: %v", err)
		}
		if err := u.handleUnTar(source+"/docker/docker_data.tar.gz", target); err != nil {
			return fmt.Errorf("recover docker data failed, err: %v", err)
		}
	case "re-recover":
		if err := u.handleUnTar(source+"/docker/docker_data.tar.gz", target); err != nil {
			return fmt.Errorf("re-recover docker data failed, err: %v", err)
		}
	case "rollback":
		if err := u.handleUnTar(source+"/docker_data.tar.gz", target); err != nil {
			return fmt.Errorf("rollback docker data failed, err: %v", err)
		}
	}
	global.LOG.Info("handle docker data dir successful!")
	return nil
}

func (u *SnapshotService) handleDaemonJson(fileOp files.FileOp, operation string, source, target string) error {
	daemonJsonPath := "/etc/docker/daemon.json"
	if operation == "snapshot" || operation == "recover" {
		_, err := os.Stat(daemonJsonPath)
		if os.IsNotExist(err) {
			global.LOG.Info("no daemon.json in snapshot and system now, nothing happened")
		}
		if err == nil {
			if err := fileOp.CopyFile(daemonJsonPath, target); err != nil {
				return fmt.Errorf("backup docker daemon.json failed, err: %v", err)
			}
		}
	}
	if operation == "recover" || operation == "rollback" || operation == "re-recover" {
		_, sourceErr := os.Stat(source)
		if os.IsNotExist(sourceErr) {
			_ = os.Remove(daemonJsonPath)
		}
		if sourceErr == nil {
			if err := fileOp.CopyFile(source, "/etc/docker"); err != nil {
				return fmt.Errorf("recover docker daemon.json failed, err: %v", err)
			}
		}
	}
	global.LOG.Info("handle docker daemon.json successful!")
	return nil
}

func (u *SnapshotService) handlePanelBinary(fileOp files.FileOp, operation string, source, target string) error {
	panelPath := "/usr/local/bin/1panel"
	if operation == "snapshot" || operation == "recover" {
		if _, err := os.Stat(panelPath); err != nil {
			return fmt.Errorf("1panel binary is not found in %s, err: %v", panelPath, err)
		} else {
			if err := cpBinary(panelPath, target); err != nil {
				return fmt.Errorf("backup 1panel binary failed, err: %v", err)
			}
		}
	}
	if operation == "recover" || operation == "rollback" || operation == "re-recover" {
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("1panel binary is not found in snapshot, err: %v", err)
		} else {
			if err := cpBinary(source, "/usr/local/bin/1panel"); err != nil {
				return fmt.Errorf("recover 1panel binary failed, err: %v", err)
			}
		}
	}
	global.LOG.Info("handle binary panel successful!")
	return nil
}
func (u *SnapshotService) handlePanelctlBinary(fileOp files.FileOp, operation string, source, target string) error {
	panelctlPath := "/usr/local/bin/1pctl"
	if operation == "snapshot" || operation == "recover" {
		if _, err := os.Stat(panelctlPath); err != nil {
			return fmt.Errorf("1pctl binary is not found in %s, err: %v", panelctlPath, err)
		} else {
			if err := cpBinary(panelctlPath, target); err != nil {
				return fmt.Errorf("backup 1pctl binary failed, err: %v", err)
			}
		}
	}
	if operation == "recover" || operation == "rollback" || operation == "re-recover" {
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("1pctl binary is not found in snapshot, err: %v", err)
		} else {
			if err := cpBinary(source, "/usr/local/bin/1pctl"); err != nil {
				return fmt.Errorf("recover 1pctl binary failed, err: %v", err)
			}
		}
	}
	global.LOG.Info("handle binary 1pactl successful!")
	return nil
}

func (u *SnapshotService) handlePanelService(fileOp files.FileOp, operation string, source, target string) error {
	panelServicePath := "/etc/systemd/system/1panel.service"
	if operation == "snapshot" || operation == "recover" {
		if _, err := os.Stat(panelServicePath); err != nil {
			return fmt.Errorf("1panel service is not found in %s, err: %v", panelServicePath, err)
		} else {
			if err := cpBinary(panelServicePath, target); err != nil {
				return fmt.Errorf("backup 1panel service failed, err: %v", err)
			}
		}
	}
	if operation == "recover" || operation == "rollback" || operation == "re-recover" {
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("1panel service is not found in snapshot, err: %v", err)
		} else {
			if err := cpBinary(source, "/etc/systemd/system/1panel.service"); err != nil {
				return fmt.Errorf("recover 1panel service failed, err: %v", err)
			}
		}
	}
	global.LOG.Info("handle panel service successful!")
	return nil
}

func (u *SnapshotService) handleBackupDatas(fileOp files.FileOp, operation string, source, target string) error {
	switch operation {
	case "snapshot":
		if err := u.handleTar(source, target, "1panel_backup.tar.gz", "./system;"); err != nil {
			return fmt.Errorf("backup panel local backup dir data failed, err: %v", err)
		}
	case "recover":
		if err := u.handleTar(target, u.OriginalPath, "1panel_backup.tar.gz", "./system;"); err != nil {
			return fmt.Errorf("restore original local backup dir data failed, err: %v", err)
		}
		if err := u.handleUnTar(source+"/1panel/1panel_backup.tar.gz", target); err != nil {
			return fmt.Errorf("recover local backup dir data failed, err: %v", err)
		}
	case "re-recover":
		if err := u.handleUnTar(source+"/1panel/1panel_backup.tar.gz", target); err != nil {
			return fmt.Errorf("retry recover  local backup dir data failed, err: %v", err)
		}
	case "rollback":
		if err := u.handleUnTar(source+"/1panel_backup.tar.gz", target); err != nil {
			return fmt.Errorf("rollback local backup dir data failed, err: %v", err)
		}
	}
	global.LOG.Info("handle backup data successful!")
	return nil
}

func (u *SnapshotService) handlePanelDatas(snapID uint, fileOp files.FileOp, operation string, source, target, backupDir, dockerDir string) error {
	switch operation {
	case "snapshot":
		exclusionRules := "./tmp;./cache;./db/1Panel.db-*;"
		if strings.Contains(backupDir, source) {
			exclusionRules += ("." + strings.ReplaceAll(backupDir, source, "") + ";")
		}
		if strings.Contains(dockerDir, source) {
			exclusionRules += ("." + strings.ReplaceAll(dockerDir, source, "") + ";")
		}

		if err := u.handleTar(source, target, "1panel_data.tar.gz", exclusionRules); err != nil {
			return fmt.Errorf("backup panel data failed, err: %v", err)
		}
	case "recover":
		exclusionRules := "./tmp/;./cache;"
		if strings.Contains(backupDir, target) {
			exclusionRules += ("." + strings.ReplaceAll(backupDir, target, "") + ";")
		}
		if strings.Contains(dockerDir, target) {
			exclusionRules += ("." + strings.ReplaceAll(dockerDir, target, "") + ";")
		}

		_ = snapshotRepo.Update(snapID, map[string]interface{}{"recover_status": ""})
		if err := u.handleTar(target, u.OriginalPath, "1panel_data.tar.gz", exclusionRules); err != nil {
			return fmt.Errorf("restore original panel data failed, err: %v", err)
		}
		_ = snapshotRepo.Update(snapID, map[string]interface{}{"recover_status": constant.StatusWaiting})

		if err := u.handleUnTar(source+"/1panel/1panel_data.tar.gz", target); err != nil {
			return fmt.Errorf("recover panel data failed, err: %v", err)
		}
	case "re-recover":
		if err := u.handleUnTar(source+"/1panel/1panel_data.tar.gz", target); err != nil {
			return fmt.Errorf("retry recover panel data failed, err: %v", err)
		}
	case "rollback":
		if err := u.handleUnTar(source+"/1panel_data.tar.gz", target); err != nil {
			return fmt.Errorf("rollback panel data failed, err: %v", err)
		}
	}
	return nil
}

func (u *SnapshotService) loadDockerDataDir() (string, bool, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", false, fmt.Errorf("new docker client failed, err: %v", err)
	}
	info, err := client.Info(context.Background())
	if err != nil {
		return "", false, fmt.Errorf("load docker info failed, err: %v", err)
	}
	return info.DockerRootDir, info.LiveRestoreEnabled, nil
}

func (u *SnapshotService) Delete(req dto.BatchDeleteReq) error {
	backups, _ := snapshotRepo.GetList(commonRepo.WithIdsIn(req.Ids))
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	for _, snap := range backups {
		if _, err := os.Stat(fmt.Sprintf("%s/system/%s/%s.tar.gz", localDir, snap.Name, snap.Name)); err == nil {
			_ = os.Remove(fmt.Sprintf("%s/system/%s/%s.tar.gz", localDir, snap.Name, snap.Name))
		}
	}
	if err := snapshotRepo.Delete(commonRepo.WithIdsIn(req.Ids)); err != nil {
		return err
	}

	return nil
}

func updateSnapshotStatus(id uint, status string, message string) {
	if status != constant.StatusSuccess {
		global.LOG.Errorf("snapshot failed, err: %s", message)
	}
	if err := snapshotRepo.Update(id, map[string]interface{}{
		"status":  status,
		"message": message,
	}); err != nil {
		global.LOG.Errorf("update snap snapshot status failed, err: %v", err)
	}
	_ = settingRepo.Update("SystemStatus", "Free")
}
func updateRecoverStatus(id uint, interruptStep, status string, message string) {
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
}
func updateRollbackStatus(id uint, status string, message string) {
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

func cpBinary(src, dst string) error {
	stderr, err := cmd.Exec(fmt.Sprintf("\\cp -f %s %s", src, dst))
	if err != nil {
		return errors.New(stderr)
	}
	return nil
}

func (u *SnapshotService) updateLiveRestore(enabled bool) error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil {
		return fmt.Errorf("load docker daemon.json conf failed, err: %v", err)
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	if !enabled {
		delete(daemonMap, "live-restore")
	} else {
		daemonMap["live-restore"] = enabled
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	stdout, err := cmd.Exec("systemctl restart docker")
	if err != nil {
		return errors.New(stdout)
	}

	ticker := time.NewTicker(3 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	for range ticker.C {
		select {
		case <-ctx.Done():
			return errors.New("the docker service cannot be restarted")
		default:
			stdout, err := cmd.Exec("systemctl is-active docker")
			if string(stdout) == "active\n" && err == nil {
				global.LOG.Info("docker restart with new live-restore successful!")
				return nil
			}
		}
	}
	return nil
}

func (u *SnapshotService) handleTar(sourceDir, targetDir, name, exclusionRules string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	exStr := ""
	excludes := strings.Split(exclusionRules, ";")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		exStr += " --exclude "
		exStr += exclude
	}

	commands := fmt.Sprintf("tar --warning=no-file-changed -zcf %s %s -C %s .", targetDir+"/"+name, exStr, sourceDir)
	global.LOG.Debug(commands)
	stdout, err := cmd.ExecWithTimeOut(commands, 30*time.Minute)
	if err != nil {
		global.LOG.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
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
		global.LOG.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
	}
	return nil
}
