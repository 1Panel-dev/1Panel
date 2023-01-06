package service

import (
	"context"
	"fmt"
	"os"
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

type SnapshotService struct{}

type ISnapshotService interface {
	SearchWithPage(req dto.PageInfo) (int64, interface{}, error)
	Create(req dto.SnapshotCreate) error
}

func NewISnapshotService() ISnapshotService {
	return &SnapshotService{}
}

func (u *SnapshotService) SearchWithPage(req dto.PageInfo) (int64, interface{}, error) {
	total, snapshots, err := snapshotRepo.Page(req.Page, req.PageSize)
	var dtoSnap []dto.SnapshotInfo
	for _, snapshot := range snapshots {
		var item dto.SnapshotInfo
		if err := copier.Copy(&item, &snapshot); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoSnap = append(dtoSnap, item)
	}
	return total, dtoSnap, err
}

func (u *SnapshotService) Create(req dto.SnapshotCreate) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	backup, err := backupRepo.Get(commonRepo.WithByType(req.BackupType))
	if err != nil {
		return err
	}
	backupAccont, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return err
	}

	timeNow := time.Now().Format("20060102150405")
	rootDir := fmt.Sprintf("/tmp/songliu/1panel_backup_%s", timeNow)
	backupPanelDir := fmt.Sprintf("%s/1panel", rootDir)
	_ = os.MkdirAll(backupPanelDir, os.ModePerm)
	backupDockerDir := fmt.Sprintf("%s/docker", rootDir)
	_ = os.MkdirAll(backupDockerDir, os.ModePerm)

	defer func() {
		_, _ = cmd.Exec("systemctl start docker")
		_ = os.RemoveAll(rootDir)
	}()

	fileOp := files.NewFileOp()
	if err := fileOp.Compress([]string{localDir}, backupPanelDir, "1panel_backup.tar.gz", files.TarGz); err != nil {
		global.LOG.Errorf("snapshot backup 1panel backup datas %s failed, err: %v", localDir, err)
		return err
	}
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	info, err := client.Info(ctx)
	if err != nil {
		return err
	}
	dataDir := info.DockerRootDir
	stdout, err := cmd.Exec("systemctl stop docker")
	if err != nil {
		return errors.New(stdout)
	}

	if _, err := os.Stat("/etc/systemd/system/1panel.service"); err == nil {
		if err := fileOp.Compress([]string{dataDir}, backupDockerDir, "docker_data.tar.gz", files.TarGz); err != nil {
			global.LOG.Errorf("snapshot backup docker data dir %s failed, err: %v", dataDir, err)
			return err
		}
	}
	if _, err := os.Stat(constant.DaemonJsonPath); err == nil {
		if err := fileOp.CopyFile(constant.DaemonJsonPath, backupDockerDir); err != nil {
			global.LOG.Errorf("snapshot backup daemon.json failed, err: %v", err)
			return err
		}
	}

	if _, err := os.Stat("/Users/slooop/go/bin/swag"); err == nil {
		if err := fileOp.CopyFile("/Users/slooop/go/bin/swag", backupPanelDir); err != nil {
			global.LOG.Errorf("snapshot backup 1panel failed, err: %v", err)
			return err
		}
	}
	if _, err := os.Stat("/etc/systemd/system/1panel.service"); err == nil {
		if err := fileOp.CopyFile("/etc/systemd/system/1panel.service", backupPanelDir); err != nil {
			global.LOG.Errorf("snapshot backup 1panel.service failed, err: %v", err)
			return err
		}
	}
	if _, err := os.Stat("/usr/local/bin/1panelctl"); err == nil {
		if err := fileOp.CopyFile("/usr/local/bin/1panelctl", backupPanelDir); err != nil {
			global.LOG.Errorf("snapshot backup 1panelctl failed, err: %v", err)
			return err
		}
	}
	if _, err := os.Stat(global.CONF.System.DataDir); err == nil {
		if err := fileOp.Compress([]string{global.CONF.System.DataDir}, backupPanelDir, "1panel_data.tar.gz", files.TarGz); err != nil {
			global.LOG.Errorf("snapshot backup 1panel data %s failed, err: %v", global.CONF.System.DataDir, err)
			return err
		}
	}
	if err := fileOp.Compress([]string{rootDir}, fmt.Sprintf("%s/system", localDir), fmt.Sprintf("1panel_backup_%s.tar.gz", timeNow), files.TarGz); err != nil {
		return err
	}

	snap := model.Snapshot{
		Name:        "1panel_backup_" + timeNow,
		Description: req.Description,
		BackupType:  req.BackupType,
		Status:      constant.StatusWaiting,
	}
	_ = snapshotRepo.Create(&snap)
	go func() {
		localPath := fmt.Sprintf("%s/system/1panel_backup_%s.tar.gz", localDir, timeNow)
		if ok, err := backupAccont.Upload(localPath, fmt.Sprintf("system_snapshot/1panel_backup_%s.tar.gz", timeNow)); err != nil || !ok {
			_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			global.LOG.Errorf("upload snapshot to %s failed, err: %v", backup.Type, err)
			return
		}
		snap.Status = constant.StatusSuccess
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusSuccess})
		global.LOG.Infof("upload snapshot to %s success", backup.Type)
	}()
	return nil
}
