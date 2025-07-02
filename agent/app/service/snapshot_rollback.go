package service

import (
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

func (u *SnapshotService) SnapshotRollback(req dto.SnapshotRecover) error {
	global.LOG.Info("start to rollback now")
	snap, err := snapshotRepo.Get(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if len(snap.TaskRollbackID) != 0 {
		req.TaskID = snap.TaskRollbackID
	} else {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"task_rollback_id": req.TaskID})
	}
	taskItem, err := task.NewTaskWithOps(snap.Name, task.TaskRollback, task.TaskScopeSnapshot, req.TaskID, snap.ID)
	if err != nil {
		global.LOG.Errorf("new task for create snapshot failed, err: %v", err)
		return err
	}
	go func() {
		rootDir := fmt.Sprintf("%s/1panel_original/original_%s", global.Dir.BaseDir, snap.Name)
		baseDir := path.Join(rootDir, "base")

		FileOp := files.NewFileOp()
		taskItem.AddSubTask(
			i18n.GetWithName("SnapCopy", "/usr/local/bin/1pctl"),
			func(t *task.Task) error {
				return FileOp.CopyFile(path.Join(baseDir, "1pctl"), "/usr/local/bin")
			},
			nil,
		)
		if global.IsMaster {
			taskItem.AddSubTask(
				i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel-core"),
				func(t *task.Task) error {
					return FileOp.CopyFile(path.Join(baseDir, "1panel"), "/usr/local/bin")
				},
				nil,
			)
		}
		taskItem.AddSubTask(
			i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel-agent"),
			func(t *task.Task) error {
				return FileOp.CopyFile(path.Join(baseDir, "1panel-agent"), "/usr/local/bin")
			},
			nil,
		)
		if global.IsMaster {
			taskItem.AddSubTask(
				i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel-core.service"),
				func(t *task.Task) error {
					return FileOp.CopyFile(path.Join(baseDir, "1panel-core.service"), "/etc/systemd/system")
				},
				nil,
			)
		}
		taskItem.AddSubTask(
			i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel-agent.service"),
			func(t *task.Task) error {
				return FileOp.CopyFile(path.Join(baseDir, "1panel-agent.service"), "/etc/systemd/system")
			},
			nil,
		)
		taskItem.AddSubTask(
			i18n.GetWithName("SnapCopy", constant.DaemonJsonPath),
			func(t *task.Task) error {
				if FileOp.Stat(path.Join(baseDir, "daemon.json")) {
					return FileOp.CopyFile(path.Join(baseDir, "daemon.json"), path.Dir(constant.DaemonJsonPath))
				}
				return nil
			},
			nil,
		)
		taskItem.AddSubTask(
			i18n.GetWithName("SnapCopy", global.Dir.LocalBackupDir),
			func(t *task.Task) error {
				return FileOp.CopyDir(path.Join(rootDir, "backup"), global.Dir.LocalBackupDir)
			},
			nil,
		)
		openrestyDir, _ := settingRepo.GetValueByKey("WEBSITE_DIR")
		if len(openrestyDir) != 0 {
			taskItem.AddSubTask(
				i18n.GetWithName("SnapCopy", openrestyDir),
				func(t *task.Task) error {
					return FileOp.CopyDir(path.Join(rootDir, "www"), openrestyDir)
				},
				nil,
			)
		}
		taskItem.AddSubTask(
			i18n.GetWithName("SnapCopy", global.Dir.BaseDir),
			func(t *task.Task) error {
				return FileOp.CopyDir(path.Join(rootDir, "1panel"), global.Dir.BaseDir)
			},
			nil,
		)
		if err := taskItem.Execute(); err != nil {
			_ = snapshotRepo.Update(req.ID, map[string]interface{}{"rollback_status": constant.StatusFailed, "rollback_message": err.Error()})
			return
		}
		_ = snapshotRepo.Update(req.ID, map[string]interface{}{
			"recover_status":   "",
			"recover_message":  "",
			"rollback_status":  "",
			"rollback_message": "",
			"interrupt_step":   "",
		})
		_ = os.RemoveAll(rootDir)
	}()
	return nil
}
