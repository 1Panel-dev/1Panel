package service

import (
	"context"
	"fmt"
	"os"
	pathUtils "path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/ntp"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
)

func (u *CronjobService) HandleJob(cronjob *model.Cronjob) {
	record := cronjobRepo.StartRecords(cronjob.ID, "", cronjob.Type)
	go func() {
		taskItem, err := task.NewTaskWithOps(fmt.Sprintf("cronjob-%s", cronjob.Name), task.TaskHandle, task.TaskScopeCronjob, record.TaskID, cronjob.ID)
		if err != nil {
			global.LOG.Errorf("new task for exec shell failed, err: %v", err)
			return
		}
		err = u.loadTask(cronjob, &record, taskItem)
		if cronjob.Type == "snapshot" {
			if err != nil {
				taskItem, _ := taskRepo.GetFirst(taskRepo.WithByID(record.TaskID))
				if len(taskItem.ID) == 0 {
					record.TaskID = ""
				}
				cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), record.Records)
				handleCronJobAlert(cronjob)
				return
			}
			cronjobRepo.EndRecords(record, constant.StatusSuccess, "", record.Records)
			return
		}
		if err != nil {
			global.LOG.Debugf("preper to handle cron job [%s] %s failed, err: %v", cronjob.Type, cronjob.Name, err)
			record.TaskID = ""
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), record.Records)
			return
		}
		if err := taskItem.Execute(); err != nil {
			taskItem, _ := taskRepo.GetFirst(taskRepo.WithByID(record.TaskID))
			if len(taskItem.ID) == 0 {
				record.TaskID = ""
			}
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), record.Records)
			handleCronJobAlert(cronjob)
		} else {
			cronjobRepo.EndRecords(record, constant.StatusSuccess, "", record.Records)
		}
	}()
}

func (u *CronjobService) loadTask(cronjob *model.Cronjob, record *model.JobRecords, taskItem *task.Task) error {
	var err error
	switch cronjob.Type {
	case "shell":
		if cronjob.ScriptMode == "library" {
			scriptItem, _ := scriptRepo.Get(repo.WithByID(cronjob.ScriptID))
			if scriptItem.ID == 0 {
				return fmt.Errorf("load script from db failed, err: %v", err)
			}
			cronjob.Script = scriptItem.Script
			cronjob.ScriptMode = "input"
		}
		if len(cronjob.Script) == 0 {
			return fmt.Errorf("the script content is empty and is skipped")
		}
		u.handleShell(*cronjob, taskItem)
	case "curl":
		if len(cronjob.URL) == 0 {
			return fmt.Errorf("the url is empty and is skipped")
		}
		u.handleCurl(*cronjob, taskItem)
	case "ntp":
		u.handleNtpSync(*cronjob, taskItem)
	case "cutWebsiteLog":
		err = u.handleCutWebsiteLog(cronjob, record.StartTime, taskItem)
	case "clean":
		u.handleSystemClean(*cronjob, taskItem)
	case "website":
		err = u.handleWebsite(*cronjob, record.StartTime, taskItem)
	case "app":
		err = u.handleApp(*cronjob, record.StartTime, taskItem)
	case "database":
		err = u.handleDatabase(*cronjob, record.StartTime, taskItem)
	case "directory":
		if len(cronjob.SourceDir) == 0 {
			return fmt.Errorf("the source dir is empty and is skipped")
		}
		err = u.handleDirectory(*cronjob, record.StartTime, taskItem)
	case "log":
		err = u.handleSystemLog(*cronjob, record.StartTime, taskItem)
	case "snapshot":
		taskItem.Task.Type = task.TaskScopeSnapshot
		_ = cronjobRepo.UpdateRecords(record.ID, map[string]interface{}{"records": record.Records})
		err = u.handleSnapshot(*cronjob, *record, taskItem)
	}
	return err
}

func (u *CronjobService) handleShell(cronjob model.Cronjob, taskItem *task.Task) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTask(*taskItem))
	taskItem.AddSubTaskWithOps(i18n.GetWithName("HandleShell", cronjob.Name), func(t *task.Task) error {
		if len(cronjob.ContainerName) != 0 {
			command := "sh"
			if len(cronjob.Command) != 0 {
				command = cronjob.Command
			}
			return cmdMgr.Run("docker", "exec", cronjob.ContainerName, command, "-c", strings.ReplaceAll(cronjob.Script, "\"", "\\\""))
		}
		if len(cronjob.Executor) == 0 {
			cronjob.Executor = "bash"
		}
		if cronjob.ScriptMode == "input" {
			fileItem := pathUtils.Join(global.Dir.DataDir, "task", "shell", cronjob.Name, cronjob.Name+".sh")
			_ = os.MkdirAll(pathUtils.Dir(fileItem), os.ModePerm)
			shellFile, err := os.OpenFile(fileItem, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
			if err != nil {
				return err
			}
			defer shellFile.Close()
			if _, err := shellFile.WriteString(cronjob.Script); err != nil {
				return err
			}
			if len(cronjob.User) == 0 {
				return cmdMgr.Run(cronjob.Executor, fileItem)
			}
			return cmdMgr.Run("sudo", "-u", cronjob.User, cronjob.Executor, fileItem)
		}
		if len(cronjob.User) == 0 {
			return cmdMgr.Run(cronjob.Executor, cronjob.Script)
		}
		if err := cmdMgr.Run("sudo", "-u", cronjob.User, cronjob.Executor, cronjob.Script); err != nil {
			return err
		}
		return nil
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
}

func (u *CronjobService) handleCurl(cronjob model.Cronjob, taskItem *task.Task) {
	taskItem.AddSubTaskWithOps(i18n.GetWithName("HandleShell", cronjob.Name), func(t *task.Task) error {
		cmdMgr := cmd.NewCommandMgr(cmd.WithTask(*taskItem))
		return cmdMgr.Run("curl", cronjob.URL)
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
}

func (u *CronjobService) handleNtpSync(cronjob model.Cronjob, taskItem *task.Task) {
	taskItem.AddSubTaskWithOps(i18n.GetMsgByKey("HandleNtpSync"), func(t *task.Task) error {
		ntpServer, err := settingRepo.Get(settingRepo.WithByKey("NtpSite"))
		if err != nil {
			return err
		}
		taskItem.Logf("ntp server: %s", ntpServer.Value)
		ntime, err := ntp.GetRemoteTime(ntpServer.Value)
		if err != nil {
			return err
		}
		if err := ntp.UpdateSystemTime(ntime.Format(constant.DateTimeLayout)); err != nil {
			return err
		}
		return nil
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
}

func (u *CronjobService) handleCutWebsiteLog(cronjob *model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	taskItem.AddSubTaskWithOps(i18n.GetWithName("CutWebsiteLog", cronjob.Name), func(t *task.Task) error {
		websites := loadWebsForJob(*cronjob)
		fileOp := files.NewFileOp()
		baseDir := GetOpenrestyDir(SitesRootDir)
		clientMap, err := NewBackupClientMap([]string{fmt.Sprintf("%v", cronjob.DownloadAccountID)})
		if err != nil {
			return fmt.Errorf("load local backup client failed, err: %v", err)
		}
		for _, website := range websites {
			taskItem.Log(website.Alias)
			var record model.BackupRecord
			record.From = "cronjob"
			record.Type = "cut-website-log"
			record.CronjobID = cronjob.ID
			record.Name = website.Alias
			record.DetailName = website.Alias
			record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
			backupDir := pathUtils.Join(global.Dir.LocalBackupDir, "log", "website", website.Alias)
			if !fileOp.Stat(backupDir) {
				_ = os.MkdirAll(backupDir, constant.DirPerm)
			}
			record.FileDir = strings.TrimPrefix(backupDir, global.Dir.LocalBackupDir+"/")
			record.FileName = fmt.Sprintf("%s_log_%s.gz", website.PrimaryDomain, startTime.Format(constant.DateTimeSlimLayout))
			if err := backupRepo.CreateRecord(&record); err != nil {
				global.LOG.Errorf("save backup record failed, err: %v", err)
				return err
			}

			websiteLogDir := pathUtils.Join(baseDir, website.Alias, "log")
			srcAccessLogPath := pathUtils.Join(websiteLogDir, "access.log")
			srcErrorLogPath := pathUtils.Join(websiteLogDir, "error.log")

			dstFilePath := pathUtils.Join(backupDir, record.FileName)
			if err := backupLogFile(dstFilePath, websiteLogDir, fileOp); err != nil {
				taskItem.LogFailedWithErr("CutWebsiteLog", err)
				continue
			} else {
				_ = fileOp.WriteFile(srcAccessLogPath, strings.NewReader(""), constant.DirPerm)
				_ = fileOp.WriteFile(srcErrorLogPath, strings.NewReader(""), constant.DirPerm)
			}
			taskItem.Log(i18n.GetMsgWithMap("CutWebsiteLogSuccess", map[string]interface{}{"name": website.PrimaryDomain, "path": dstFilePath}))
			u.removeExpiredBackup(*cronjob, clientMap, record)
		}
		return nil
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	return nil
}

func backupLogFile(dstFilePath, websiteLogDir string, fileOp files.FileOp) error {
	cmdMgr := cmd.NewCommandMgr()
	if err := cmdMgr.RunBashCf("tar -czf %s -C %s %s", dstFilePath, websiteLogDir, strings.Join([]string{"access.log", "error.log"}, " ")); err != nil {
		dstDir := pathUtils.Dir(dstFilePath)
		if err = fileOp.Copy(pathUtils.Join(websiteLogDir, "access.log"), dstDir); err != nil {
			return err
		}
		if err = fileOp.Copy(pathUtils.Join(websiteLogDir, "error.log"), dstDir); err != nil {
			return err
		}
		if err = cmdMgr.RunBashCf("tar -czf %s -C %s %s", dstFilePath, dstDir, strings.Join([]string{"access.log", "error.log"}, " ")); err != nil {
			return err
		}
		_ = fileOp.DeleteFile(pathUtils.Join(dstDir, "access.log"))
		_ = fileOp.DeleteFile(pathUtils.Join(dstDir, "error.log"))
		return nil
	}
	return nil
}

func (u *CronjobService) handleSystemClean(cronjob model.Cronjob, taskItem *task.Task) {
	cleanTask := doSystemClean(taskItem)
	taskItem.AddSubTaskWithOps(i18n.GetMsgByKey("HandleSystemClean"), cleanTask, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
}

func (u *CronjobService) uploadCronjobBackFile(cronjob model.Cronjob, accountMap map[string]backupClientHelper, file string) (string, error) {
	defer func() {
		_ = os.Remove(file)
	}()
	accounts := strings.Split(cronjob.SourceAccountIDs, ",")
	cloudSrc := strings.TrimPrefix(file, global.Dir.TmpDir+"/")
	for _, account := range accounts {
		if len(account) != 0 {
			global.LOG.Debugf("start upload file to %s, dir: %s", accountMap[account].name, pathUtils.Join(accountMap[account].backupPath, cloudSrc))
			if _, err := accountMap[account].client.Upload(file, pathUtils.Join(accountMap[account].backupPath, cloudSrc)); err != nil {
				global.LOG.Errorf("upload file to %s failed, err: %v", accountMap[account].name, err)
				continue
			}
			global.LOG.Debugf("upload successful!")
		}
	}
	return cloudSrc, nil
}

func (u *CronjobService) removeExpiredBackup(cronjob model.Cronjob, accountMap map[string]backupClientHelper, record model.BackupRecord) {
	var opts []repo.DBOption
	opts = append(opts, repo.WithByFrom("cronjob"))
	opts = append(opts, backupRepo.WithByCronID(cronjob.ID))
	opts = append(opts, repo.WithOrderBy("created_at desc"))
	if record.ID != 0 {
		opts = append(opts, repo.WithByType(record.Type))
		opts = append(opts, repo.WithByName(record.Name))
		opts = append(opts, repo.WithByDetailName(record.DetailName))
	}
	records, _ := backupRepo.ListRecord(opts...)
	if len(records) <= int(cronjob.RetainCopies) {
		return
	}
	for i := int(cronjob.RetainCopies); i < len(records); i++ {
		accounts := strings.Split(cronjob.SourceAccountIDs, ",")
		if cronjob.Type == "snapshot" {
			for _, account := range accounts {
				if len(account) != 0 {
					if _, ok := accountMap[account]; !ok {
						continue
					}
					_, _ = accountMap[account].client.Delete(pathUtils.Join(accountMap[account].backupPath, "system_snapshot", records[i].FileName))
				}
			}
			_ = snapshotRepo.Delete(repo.WithByName(strings.TrimSuffix(records[i].FileName, ".tar.gz")))
		} else {
			for _, account := range accounts {
				if len(account) != 0 {
					if _, ok := accountMap[account]; !ok {
						continue
					}
					_, _ = accountMap[account].client.Delete(pathUtils.Join(accountMap[account].backupPath, records[i].FileDir, records[i].FileName))
				}
			}
		}
		_ = backupRepo.DeleteRecord(context.Background(), repo.WithByID(records[i].ID))
	}
}

func hasBackup(cronjobType string) bool {
	return cronjobType == "app" || cronjobType == "database" || cronjobType == "website" || cronjobType == "directory" || cronjobType == "snapshot" || cronjobType == "log" || cronjobType == "cutWebsiteLog"
}

func handleCronJobAlert(cronjob *model.Cronjob) {
	pushAlert := dto.PushAlert{
		TaskName:  cronjob.Name,
		AlertType: cronjob.Type,
		EntryID:   cronjob.ID,
		Param:     cronjob.Type,
	}
	err := xpack.PushAlert(pushAlert)
	if err != nil {
		global.LOG.Errorf("cronjob alert push failed, err: %v", err)
		return
	}
}
