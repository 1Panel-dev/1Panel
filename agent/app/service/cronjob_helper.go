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
	"github.com/1Panel-dev/1Panel/agent/buserr"
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
		var (
			message []byte
			err     error
		)
		cronjob.RetryTimes = cronjob.RetryTimes + 1
		for i := uint(0); i < cronjob.RetryTimes; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cronjob.Timeout)*time.Second)
			done := make(chan error)
			go func() {
				message, err = u.handleJob(cronjob, record)
				if err != nil {
					global.LOG.Debugf("try handle cron job [%s] %s failed %d/%d, err: %v", cronjob.Type, cronjob.Name, i+1, cronjob.RetryTimes, err)
				}
				close(done)
			}()
			select {
			case <-done:
				cancel()
			case <-ctx.Done():
				global.LOG.Debugf("try handle cron job [%s] %s failed %d/%d, err: timeout", cronjob.Type, cronjob.Name, i+1, cronjob.RetryTimes)
				err = fmt.Errorf("handle timeout")
				cancel()
				continue
			}
		}
		if err != nil {
			if len(message) != 0 {
				record.Records, _ = mkdirAndWriteFile(cronjob, record.StartTime, message)
			}
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), record.Records)
			handleCronJobAlert(cronjob)
			return
		}
		if len(message) != 0 {
			record.Records, err = mkdirAndWriteFile(cronjob, record.StartTime, message)
			if err != nil {
				global.LOG.Errorf("save file %s failed, err: %v", record.Records, err)
			}
		}
		cronjobRepo.EndRecords(record, constant.StatusSuccess, "", record.Records)
	}()
}

func (u *CronjobService) handleJob(cronjob *model.Cronjob, record model.JobRecords) ([]byte, error) {
	var (
		message []byte
		err     error
	)
	switch cronjob.Type {
	case "shell":
		if len(cronjob.Script) == 0 {
			return nil, fmt.Errorf("the script content is empty and is skipped")
		}
		err = u.handleShell(*cronjob, record.TaskID)
	case "curl":
		if len(cronjob.URL) == 0 {
			return nil, fmt.Errorf("the url is empty and is skipped")
		}
		err = u.handleCurl(*cronjob, record.TaskID)
	case "ntp":
		err = u.handleNtpSync(*cronjob, record.TaskID)
	case "cutWebsiteLog":
		var messageItem []string
		messageItem, record.File, err = u.handleCutWebsiteLog(cronjob, record.StartTime)
		message = []byte(strings.Join(messageItem, "\n"))
	case "clean":
		err = u.handleSystemClean(*cronjob, record.TaskID)
	case "website":
		err = u.handleWebsite(*cronjob, record.StartTime, record.TaskID)
	case "app":
		err = u.handleApp(*cronjob, record.StartTime, record.TaskID)
	case "database":
		err = u.handleDatabase(*cronjob, record.StartTime, record.TaskID)
	case "directory":
		if len(cronjob.SourceDir) == 0 {
			return nil, fmt.Errorf("the source dir is empty and is skipped")
		}
		err = u.handleDirectory(*cronjob, record.StartTime)
	case "log":
		err = u.handleSystemLog(*cronjob, record.StartTime)
	case "snapshot":
		_ = cronjobRepo.UpdateRecords(record.ID, map[string]interface{}{"records": record.Records})
		err = u.handleSnapshot(*cronjob, record)
	}
	return message, err
}

func (u *CronjobService) handleShell(cronjob model.Cronjob, taskID string) error {
	taskItem, err := task.NewTaskWithOps(fmt.Sprintf("cronjob-%s", cronjob.Name), task.TaskHandle, task.TaskScopeCronjob, taskID, cronjob.ID)
	if err != nil {
		global.LOG.Errorf("new task for exec shell failed, err: %v", err)
		return err
	}

	cmdMgr := cmd.NewCommandMgr(cmd.WithTask(*taskItem), cmd.WithTimeout(24*time.Hour))
	taskItem.AddSubTask(i18n.GetWithName("HandleShell", cronjob.Name), func(t *task.Task) error {
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
	},
		nil,
	)
	return taskItem.Execute()
}

func (u *CronjobService) handleCurl(cronjob model.Cronjob, taskID string) error {
	taskItem, err := task.NewTaskWithOps(fmt.Sprintf("cronjob-%s", cronjob.Name), task.TaskHandle, task.TaskScopeCronjob, taskID, cronjob.ID)
	if err != nil {
		global.LOG.Errorf("new task for exec shell failed, err: %v", err)
		return err
	}

	taskItem.AddSubTask(i18n.GetWithName("HandleShell", cronjob.Name), func(t *task.Task) error {
		cmdMgr := cmd.NewCommandMgr(cmd.WithTask(*taskItem), cmd.WithTimeout(time.Hour))
		return cmdMgr.Run("curl", cronjob.URL)
	},
		nil,
	)
	return taskItem.Execute()
}

func (u *CronjobService) handleNtpSync(cronjob model.Cronjob, taskID string) error {
	taskItem, err := task.NewTaskWithOps(fmt.Sprintf("cronjob-%s", cronjob.Name), task.TaskHandle, task.TaskScopeCronjob, taskID, cronjob.ID)
	if err != nil {
		global.LOG.Errorf("new task for exec shell failed, err: %v", err)
		return err
	}

	taskItem.AddSubTask(i18n.GetMsgByKey("HandleNtpSync"), func(t *task.Task) error {
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
	}, nil)
	return taskItem.Execute()
}

func (u *CronjobService) handleCutWebsiteLog(cronjob *model.Cronjob, startTime time.Time) ([]string, string, error) {
	var (
		err       error
		filePaths []string
		msgs      []string
	)
	websites := loadWebsForJob(*cronjob)
	fileOp := files.NewFileOp()
	baseDir := GetOpenrestyDir(SitesRootDir)
	for _, website := range websites {
		websiteLogDir := pathUtils.Join(baseDir, website.Alias, "log")
		srcAccessLogPath := pathUtils.Join(websiteLogDir, "access.log")
		srcErrorLogPath := pathUtils.Join(websiteLogDir, "error.log")
		dstLogDir := pathUtils.Join(global.Dir.LocalBackupDir, "log", "website", website.Alias)
		if !fileOp.Stat(dstLogDir) {
			_ = os.MkdirAll(dstLogDir, constant.DirPerm)
		}

		dstName := fmt.Sprintf("%s_log_%s.gz", website.PrimaryDomain, startTime.Format(constant.DateTimeSlimLayout))
		dstFilePath := pathUtils.Join(dstLogDir, dstName)
		filePaths = append(filePaths, dstFilePath)

		if err = backupLogFile(dstFilePath, websiteLogDir, fileOp); err != nil {
			websiteErr := buserr.WithNameAndErr("ErrCutWebsiteLog", website.PrimaryDomain, err)
			err = websiteErr
			msgs = append(msgs, websiteErr.Error())
			global.LOG.Error(websiteErr.Error())
			continue
		} else {
			_ = fileOp.WriteFile(srcAccessLogPath, strings.NewReader(""), constant.DirPerm)
			_ = fileOp.WriteFile(srcErrorLogPath, strings.NewReader(""), constant.DirPerm)
		}
		msg := i18n.GetMsgWithMap("CutWebsiteLogSuccess", map[string]interface{}{"name": website.PrimaryDomain, "path": dstFilePath})
		msgs = append(msgs, msg)
	}
	u.removeExpiredLog(*cronjob)
	return msgs, strings.Join(filePaths, ","), err
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

func (u *CronjobService) handleSystemClean(cronjob model.Cronjob, taskID string) error {
	taskItem, err := task.NewTaskWithOps(fmt.Sprintf("cronjob-%s", cronjob.Name), task.TaskHandle, task.TaskScopeCronjob, taskID, cronjob.ID)
	if err != nil {
		global.LOG.Errorf("new task for system clean failed, err: %v", err)
		return err
	}
	return systemClean(taskItem)
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
				return "", err
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
					_, _ = accountMap[account].client.Delete(pathUtils.Join(accountMap[account].backupPath, "system_snapshot", records[i].FileName))
				}
			}
			_ = snapshotRepo.Delete(repo.WithByName(strings.TrimSuffix(records[i].FileName, ".tar.gz")))
		} else {
			for _, account := range accounts {
				if len(account) != 0 {
					_, _ = accountMap[account].client.Delete(pathUtils.Join(accountMap[account].backupPath, records[i].FileDir, records[i].FileName))
				}
			}
		}
		_ = backupRepo.DeleteRecord(context.Background(), repo.WithByID(records[i].ID))
	}
}

func (u *CronjobService) removeExpiredLog(cronjob model.Cronjob) {
	records, _ := cronjobRepo.ListRecord(cronjobRepo.WithByJobID(int(cronjob.ID)), repo.WithOrderBy("created_at desc"))
	if len(records) <= int(cronjob.RetainCopies) {
		return
	}
	for i := int(cronjob.RetainCopies); i < len(records); i++ {
		if len(records[i].File) != 0 {
			files := strings.Split(records[i].File, ",")
			for _, file := range files {
				_ = os.Remove(file)
			}
		}
		_ = cronjobRepo.DeleteRecord(repo.WithByID(records[i].ID))
		_ = os.Remove(records[i].Records)
	}
}

func hasBackup(cronjobType string) bool {
	return cronjobType == "app" || cronjobType == "database" || cronjobType == "website" || cronjobType == "directory" || cronjobType == "snapshot" || cronjobType == "log"
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
