package service

import (
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

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

func handleRollback(name string) error {
	rootDir := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, name)
	baseDir := path.Join(rootDir, "base")

	FileOp := files.NewFileOp()
	if err := FileOp.CopyDir(path.Join(rootDir, "1panel"), global.CONF.System.BaseDir); err != nil {
		return err
	}
	if err := FileOp.CopyDir(path.Join(rootDir, "backup"), path.Dir(global.CONF.System.Backup)); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1pctl"), "/usr/local/bin/1pctl"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel"), "/usr/local/bin/1panel"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel_agent"), "/usr/local/bin/1panel_agent"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel.service"), "/etc/systemd/system/1panel.service"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel_agent.service"), "/etc/systemd/system/1panel_agent.service"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "daemon.json"), "/etc/docker/daemon.json"); err != nil {
		return err
	}
	_ = os.RemoveAll(rootDir)
	return nil
}
