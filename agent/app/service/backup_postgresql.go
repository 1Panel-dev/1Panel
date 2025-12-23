package service

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/postgresql/client"
)

func (u *BackupService) PostgresqlBackup(req dto.CommonBackup) error {
	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	itemDir := fmt.Sprintf("database/%s/%s/%s", req.Type, req.Name, req.DetailName)
	targetDir := path.Join(global.Dir.LocalBackupDir, itemDir)
	fileName := fmt.Sprintf("%s_%s.sql.gz", req.DetailName, timeNow+common.RandStrAndNum(5))

	record := &model.BackupRecord{
		Type:              req.Type,
		Name:              req.Name,
		DetailName:        req.DetailName,
		SourceAccountIDs:  "1",
		DownloadAccountID: 1,
		FileDir:           itemDir,
		FileName:          fileName,
		TaskID:            req.TaskID,
		Status:            constant.StatusWaiting,
		Description:       req.Description,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
	}

	databaseHelper := DatabaseHelper{Database: req.Name, DBType: req.Type, Name: req.DetailName}
	if err := handlePostgresqlBackup(databaseHelper, nil, record.ID, targetDir, fileName, req.TaskID, req.Secret); err != nil {
		return err
	}
	return nil
}
func (u *BackupService) PostgresqlRecover(req dto.CommonRecover) error {
	if err := handlePostgresqlRecover(req, nil, false); err != nil {
		return err
	}
	return nil
}

func (u *BackupService) PostgresqlRecoverByUpload(req dto.CommonRecover) error {
	recoverFile, err := loadSqlFile(req.File)
	if err != nil {
		return err
	}
	req.File = recoverFile
	if err := handlePostgresqlRecover(req, nil, false); err != nil {
		return err
	}
	global.LOG.Info("recover from uploads successful!")
	return nil
}

func handlePostgresqlBackup(db DatabaseHelper, parentTask *task.Task, recordID uint, targetDir, fileName, taskID, secret string) error {
	var (
		err        error
		backupTask *task.Task
	)
	backupTask = parentTask
	itemName := fmt.Sprintf("%s - %s", db.Database, db.Name)
	if parentTask == nil {
		backupTask, err = task.NewTaskWithOps(itemName, task.TaskBackup, task.TaskScopeBackup, taskID, db.ID)
		if err != nil {
			return err
		}
	}

	itemHandler := func() error { return doPostgresqlBackup(db, targetDir, fileName, secret, backupTask) }
	if parentTask != nil {
		return itemHandler()
	}
	backupTask.AddSubTaskWithOps(task.GetTaskName(itemName, task.TaskBackup, task.TaskScopeBackup), func(t *task.Task) error { return itemHandler() }, nil, 0, 3*time.Hour)
	go func() {
		if err := backupTask.Execute(); err != nil {
			backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			return
		}
		backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusSuccess})
	}()
	return nil
}

func handlePostgresqlRecover(req dto.CommonRecover, parentTask *task.Task, isRollback bool) error {
	var (
		err      error
		itemTask *task.Task
	)
	dbInfo, err := postgresqlRepo.Get(repo.WithByName(req.DetailName), postgresqlRepo.WithByPostgresqlName(req.Name))
	if err != nil {
		return err
	}
	itemTask = parentTask
	if parentTask == nil {
		itemTask, err = task.NewTaskWithOps(req.Name, task.TaskRecover, task.TaskScopeBackup, req.TaskID, dbInfo.ID)
		if err != nil {
			return err
		}
	}

	recoverDatabase := func(t *task.Task) error {
		isOk := false
		fileOp := files.NewFileOp()
		if !fileOp.Stat(req.File) {
			return buserr.WithName("ErrFileNotFound", req.File)
		}

		cli, err := LoadPostgresqlClientByFrom(req.Name)
		if err != nil {
			return err
		}
		defer cli.Close()

		if !isRollback {
			rollbackFile := path.Join(global.Dir.TmpDir, fmt.Sprintf("database/%s/%s_%s.sql.gz", req.Type, req.DetailName, time.Now().Format(constant.DateTimeSlimLayout)))
			if err := cli.Backup(client.BackupInfo{
				Database:  req.Name,
				Name:      req.DetailName,
				TargetDir: path.Dir(rollbackFile),
				FileName:  path.Base(rollbackFile),

				Task:    t,
				Timeout: 300,
			}); err != nil {
				return fmt.Errorf("backup postgresql db %s for rollback before recover failed, err: %v", req.DetailName, err)
			}
			defer func() {
				if !isOk {
					global.LOG.Info("recover failed, start to rollback now")
					if err := cli.Recover(client.RecoverInfo{
						Database:   req.Name,
						Name:       req.DetailName,
						SourceFile: rollbackFile,

						Task:    t,
						Timeout: 300,
					}); err != nil {
						global.LOG.Errorf("rollback postgresql db %s from %s failed, err: %v", req.DetailName, rollbackFile, err)
					} else {
						global.LOG.Infof("rollback postgresql db %s from %s successful", req.DetailName, rollbackFile)
					}
					_ = os.RemoveAll(rollbackFile)
				} else {
					_ = os.RemoveAll(rollbackFile)
				}
			}()
		}
		if len(req.Secret) != 0 {
			err = files.OpensslDecrypt(req.File, req.Secret)
			if err != nil {
				return err
			}
			req.File = path.Join(path.Dir(req.File), "tmp_"+path.Base(req.File))
			defer os.Remove(req.File)
			t.LogWithStatus(i18n.GetMsgByKey("Decrypt"), err)
		}
		if err := cli.Recover(client.RecoverInfo{
			Database:   req.Name,
			Name:       req.DetailName,
			SourceFile: req.File,
			Username:   dbInfo.Username,
			Task:       t,
			Timeout:    300,
		}); err != nil {
			global.LOG.Errorf("recover postgresql db %s from %s failed, err: %v", req.DetailName, req.File, err)
			return err
		}
		isOk = true
		return nil
	}
	if parentTask != nil {
		return recoverDatabase(parentTask)
	}

	var timeout time.Duration
	switch req.Timeout {
	case -1:
		timeout = 0
	case 0:
		timeout = 3 * time.Hour
	default:
		timeout = time.Duration(req.Timeout) * time.Second
	}
	itemTask.AddSubTaskWithOps(i18n.GetMsgByKey("TaskRecover"), recoverDatabase, nil, 0, timeout)
	go func() {
		_ = itemTask.Execute()
	}()
	return nil
}

func doPostgresqlBackup(db DatabaseHelper, targetDir, fileName, secret string, task *task.Task) error {
	cli, err := LoadPostgresqlClientByFrom(db.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	backupInfo := client.BackupInfo{
		Database:  db.Database,
		Name:      db.Name,
		TargetDir: targetDir,
		FileName:  fileName,

		Task:    task,
		Timeout: 300,
	}
	if err := cli.Backup(backupInfo); err != nil {
		return err
	}
	if len(secret) != 0 {
		return files.OpensslEncrypt(path.Join(targetDir, fileName), secret)
	}
	return nil
}
