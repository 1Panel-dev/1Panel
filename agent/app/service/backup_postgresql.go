package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	pgclient "github.com/1Panel-dev/1Panel/agent/utils/postgresql/client"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/postgresql/client"
)

func (u *BackupService) PostgresqlBackup(req dto.CommonBackup) error {
	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	itemDir := fmt.Sprintf("database/%s/%s/%s", req.Type, req.Name, req.DetailName)
	targetDir := path.Join(global.CONF.System.Backup, itemDir)
	fileName := fmt.Sprintf("%s_%s.sql.gz", req.DetailName, timeNow+common.RandStrAndNum(5))

	databaseHelper := DatabaseHelper{Database: req.Name, DBType: req.Type, Name: req.DetailName}
	if err := handlePostgresqlBackup(databaseHelper, nil, targetDir, fileName, req.TaskID); err != nil {
		return err
	}

	record := &model.BackupRecord{
		Type:              req.Type,
		Name:              req.Name,
		DetailName:        req.DetailName,
		SourceAccountIDs:  "1",
		DownloadAccountID: 1,
		FileDir:           itemDir,
		FileName:          fileName,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
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
	file := req.File
	fileName := path.Base(req.File)
	if strings.HasSuffix(fileName, ".tar.gz") {
		fileNameItem := time.Now().Format(constant.DateTimeSlimLayout)
		dstDir := fmt.Sprintf("%s/%s", path.Dir(req.File), fileNameItem)
		fileOp := files.NewFileOp()
		if !fileOp.Stat(dstDir) {
			if err := fileOp.CreateDir(dstDir, os.ModePerm); err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", dstDir, err)
			}
		}
		if err := fileOp.TarGzExtractPro(req.File, dstDir, ""); err != nil {
			_ = os.RemoveAll(dstDir)
			return err
		}
		global.LOG.Infof("decompress file %s successful, now start to check test.sql is exist", req.File)
		hasTestSql := false
		_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "test.sql" {
				hasTestSql = true
				file = path
				fileName = "test.sql"
			}
			return nil
		})
		if !hasTestSql {
			_ = os.RemoveAll(dstDir)
			return fmt.Errorf("no such file named test.sql in %s", fileName)
		}
		defer func() {
			_ = os.RemoveAll(dstDir)
		}()
	}

	req.File = path.Dir(file) + "/" + fileName
	if err := handlePostgresqlRecover(req, nil, false); err != nil {
		return err
	}
	global.LOG.Info("recover from uploads successful!")
	return nil
}

func handlePostgresqlBackup(db DatabaseHelper, parentTask *task.Task, targetDir, fileName, taskID string) error {
	var (
		err      error
		itemTask *task.Task
	)
	itemTask = parentTask
	itemName := fmt.Sprintf("%s - %s", db.Database, db.Name)
	if parentTask == nil {
		itemTask, err = task.NewTaskWithOps(itemName, task.TaskBackup, task.TaskScopeDatabase, taskID, db.ID)
		if err != nil {
			return err
		}
	}
	backupDatabase := func(t *task.Task) error {
		cli, err := LoadPostgresqlClientByFrom(db.Database)
		if err != nil {
			return err
		}
		defer cli.Close()
		backupInfo := pgclient.BackupInfo{
			Name:      db.Name,
			TargetDir: targetDir,
			FileName:  fileName,

			Timeout: 300,
		}
		return cli.Backup(backupInfo)
	}
	itemTask.AddSubTask(i18n.GetMsgByKey("TaskBackup"), backupDatabase, nil)
	if parentTask != nil {
		return backupDatabase(parentTask)
	}

	return itemTask.Execute()
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
		itemTask, err = task.NewTaskWithOps("Redis", task.TaskRecover, task.TaskScopeDatabase, req.TaskID, dbInfo.ID)
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
			rollbackFile := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("database/%s/%s_%s.sql.gz", req.Type, req.DetailName, time.Now().Format(constant.DateTimeSlimLayout)))
			if err := cli.Backup(client.BackupInfo{
				Name:      req.DetailName,
				TargetDir: path.Dir(rollbackFile),
				FileName:  path.Base(rollbackFile),

				Timeout: 300,
			}); err != nil {
				return fmt.Errorf("backup postgresql db %s for rollback before recover failed, err: %v", req.DetailName, err)
			}
			defer func() {
				if !isOk {
					global.LOG.Info("recover failed, start to rollback now")
					if err := cli.Recover(client.RecoverInfo{
						Name:       req.DetailName,
						SourceFile: rollbackFile,

						Timeout: 300,
					}); err != nil {
						global.LOG.Errorf("rollback postgresql db %s from %s failed, err: %v", req.DetailName, rollbackFile, err)
					}
					global.LOG.Infof("rollback postgresql db %s from %s successful", req.DetailName, rollbackFile)
					_ = os.RemoveAll(rollbackFile)
				} else {
					_ = os.RemoveAll(rollbackFile)
				}
			}()
		}
		if err := cli.Recover(client.RecoverInfo{
			Name:       req.DetailName,
			SourceFile: req.File,
			Username:   dbInfo.Username,
			Timeout:    300,
		}); err != nil {
			return err
		}
		isOk = true
		return nil
	}
	itemTask.AddSubTask(i18n.GetMsgByKey("TaskRecover"), recoverDatabase, nil)
	if parentTask != nil {
		return recoverDatabase(parentTask)
	}

	return itemTask.Execute()
}
