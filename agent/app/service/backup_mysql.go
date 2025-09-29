package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"

	"github.com/1Panel-dev/1Panel/agent/buserr"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/mysql/client"
)

func (u *BackupService) MysqlBackup(req dto.CommonBackup) error {
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
		return err
	}

	databaseHelper := DatabaseHelper{Database: req.Name, DBType: req.Type, Name: req.DetailName}
	if err := handleMysqlBackup(databaseHelper, nil, record.ID, targetDir, fileName, req.TaskID); err != nil {
		backupRepo.UpdateRecordByMap(record.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
		return err
	}
	return nil
}

func (u *BackupService) MysqlRecover(req dto.CommonRecover) error {
	return handleMysqlRecover(req, nil, false, req.TaskID)
}

func (u *BackupService) MysqlRecoverByUpload(req dto.CommonRecover) error {
	recoveFile, err := loadSqlFile(req.File)
	if err != nil {
		return err
	}
	req.File = recoveFile

	if err := handleMysqlRecover(req, nil, false, req.TaskID); err != nil {
		return err
	}
	return nil
}

func handleMysqlBackup(db DatabaseHelper, parentTask *task.Task, recordID uint, targetDir, fileName, taskID string) error {
	var (
		err        error
		backupTask *task.Task
	)
	backupTask = parentTask
	dbInfo, err := mysqlRepo.Get(repo.WithByName(db.Name), mysqlRepo.WithByMysqlName(db.Database))
	if err != nil {
		return err
	}
	itemName := fmt.Sprintf("%s[%s] - %s", db.Database, db.DBType, db.Name)
	if parentTask == nil {
		backupTask, err = task.NewTaskWithOps(itemName, task.TaskBackup, task.TaskScopeBackup, taskID, dbInfo.ID)
		if err != nil {
			return err
		}
	}

	itemHandler := func() error { return doMysqlBackup(db, targetDir, fileName) }
	if parentTask != nil {
		return itemHandler()
	}
	backupTask.AddSubTaskWithOps(task.GetTaskName(itemName, task.TaskBackup, task.TaskScopeBackup), func(t *task.Task) error { return itemHandler() }, nil, 3, time.Hour)
	go func() {
		if err := backupTask.Execute(); err != nil {
			backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			return
		}
		backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusSuccess})
	}()
	return nil
}

func handleMysqlRecover(req dto.CommonRecover, parentTask *task.Task, isRollback bool, taskID string) error {
	var (
		err      error
		itemTask *task.Task
	)
	itemTask = parentTask
	dbInfo, err := mysqlRepo.Get(repo.WithByName(req.DetailName), mysqlRepo.WithByMysqlName(req.Name))
	if err != nil {
		return err
	}
	itemName := fmt.Sprintf("%s[%s] - %s", req.Name, req.Type, req.DetailName)
	if parentTask == nil {
		itemTask, err = task.NewTaskWithOps(itemName, task.TaskRecover, task.TaskScopeBackup, taskID, dbInfo.ID)
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
		dbInfo, err := mysqlRepo.Get(repo.WithByName(req.DetailName), mysqlRepo.WithByMysqlName(req.Name))
		if err != nil {
			return err
		}
		cli, version, err := LoadMysqlClientByFrom(req.Name)
		if err != nil {
			return err
		}

		if !isRollback {
			rollbackFile := path.Join(global.Dir.TmpDir, fmt.Sprintf("database/%s/%s_%s.sql.gz", req.Type, req.DetailName, time.Now().Format(constant.DateTimeSlimLayout)))
			if err := cli.Backup(client.BackupInfo{
				Name:      req.DetailName,
				Type:      req.Type,
				Version:   version,
				Format:    dbInfo.Format,
				TargetDir: path.Dir(rollbackFile),
				FileName:  path.Base(rollbackFile),
			}); err != nil {
				return fmt.Errorf("backup mysql db %s for rollback before recover failed, err: %v", req.DetailName, err)
			}
			defer func() {
				if !isOk {
					global.LOG.Info("recover failed, start to rollback now")
					if err := cli.Recover(client.RecoverInfo{
						Name:       req.DetailName,
						Type:       req.Type,
						Version:    version,
						Format:     dbInfo.Format,
						SourceFile: rollbackFile,
					}); err != nil {
						global.LOG.Errorf("rollback mysql db %s from %s failed, err: %v", req.DetailName, rollbackFile, err)
					} else {
						global.LOG.Infof("rollback mysql db %s from %s successful", req.DetailName, rollbackFile)
					}
					_ = os.RemoveAll(rollbackFile)
				} else {
					_ = os.RemoveAll(rollbackFile)
				}
			}()
		}
		if err := cli.Recover(client.RecoverInfo{
			Name:       req.DetailName,
			Type:       req.Type,
			Version:    version,
			Format:     dbInfo.Format,
			SourceFile: req.File,
		}); err != nil {
			global.LOG.Errorf("recover mysql db %s from %s failed, err: %v", req.DetailName, req.File, err)
			return err
		}
		isOk = true
		return nil
	}
	if parentTask != nil {
		return recoverDatabase(parentTask)
	}

	itemTask.AddSubTaskWithOps(i18n.GetMsgByKey("TaskRecover"), recoverDatabase, nil, 3, time.Hour)
	go func() {
		_ = itemTask.Execute()
	}()
	return nil
}

func doMysqlBackup(db DatabaseHelper, targetDir, fileName string) error {
	dbInfo, err := mysqlRepo.Get(repo.WithByName(db.Name), mysqlRepo.WithByMysqlName(db.Database))
	if err != nil {
		return err
	}
	cli, version, err := LoadMysqlClientByFrom(db.Database)
	if err != nil {
		return err
	}
	backupInfo := client.BackupInfo{
		Name:      db.Name,
		Type:      db.DBType,
		Version:   version,
		Format:    dbInfo.Format,
		TargetDir: targetDir,
		FileName:  fileName,
	}
	return cli.Backup(backupInfo)
}

func loadSqlFile(file string) (string, error) {
	if !strings.HasSuffix(file, ".tar.gz") && !strings.HasSuffix(file, ".zip") {
		return file, nil
	}
	fileName := path.Base(file)
	fileDir := path.Dir(file)
	fileNameItem := time.Now().Format(constant.DateTimeSlimLayout)
	dstDir := fmt.Sprintf("%s/%s", fileDir, fileNameItem)
	_ = os.Mkdir(dstDir, constant.DirPerm)
	if strings.HasSuffix(fileName, ".tar.gz") {
		fileOp := files.NewFileOp()
		if err := fileOp.TarGzExtractPro(file, dstDir, ""); err != nil {
			_ = os.RemoveAll(dstDir)
			return "", err
		}
	}
	if strings.HasSuffix(fileName, ".zip") {
		archiver, err := files.NewShellArchiver(files.Zip)
		if err != nil {
			_ = os.RemoveAll(dstDir)
			return "", err
		}
		if err := archiver.Extract(file, dstDir, ""); err != nil {
			_ = os.RemoveAll(dstDir)
			return "", err
		}
	}
	global.LOG.Infof("decompress file %s successful, now start to check test.sql is exist", file)
	var sqlFiles []string
	hasTestSql := false
	_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			sqlFiles = append(sqlFiles, path)
			if info.Name() == "test.sql" {
				hasTestSql = true
			}
		}
		return nil
	})
	if len(sqlFiles) == 1 {
		return sqlFiles[0], nil
	}
	if !hasTestSql {
		_ = os.RemoveAll(dstDir)
		return "", fmt.Errorf("no such file named test.sql in %s", fileName)
	}
	return "", nil
}
