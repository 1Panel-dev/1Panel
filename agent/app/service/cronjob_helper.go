package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/ntp"
	"github.com/pkg/errors"
)

func (u *CronjobService) HandleJob(cronjob *model.Cronjob) {
	var (
		message []byte
		err     error
	)
	record := cronjobRepo.StartRecords(cronjob.ID, "")
	go func() {
		switch cronjob.Type {
		case "shell":
			if len(cronjob.Script) == 0 {
				return
			}
			record.Records = u.generateLogsPath(*cronjob, record.StartTime)
			_ = cronjobRepo.UpdateRecords(record.ID, map[string]interface{}{"records": record.Records})
			err = u.handleShell(*cronjob, record.Records)
			u.removeExpiredLog(*cronjob)
		case "curl":
			if len(cronjob.URL) == 0 {
				return
			}
			record.Records = u.generateLogsPath(*cronjob, record.StartTime)
			_ = cronjobRepo.UpdateRecords(record.ID, map[string]interface{}{"records": record.Records})
			err = cmd.ExecShell(record.Records, 24*time.Hour, "bash", "-c", "curl", cronjob.URL)
			u.removeExpiredLog(*cronjob)
		case "ntp":
			err = u.handleNtpSync()
			u.removeExpiredLog(*cronjob)
		case "cutWebsiteLog":
			var messageItem []string
			messageItem, record.File, err = u.handleCutWebsiteLog(cronjob, record.StartTime)
			message = []byte(strings.Join(messageItem, "\n"))
		case "clean":
			messageItem := ""
			messageItem, err = u.handleSystemClean()
			message = []byte(messageItem)
			u.removeExpiredLog(*cronjob)
		case "website":
			err = u.handleWebsite(*cronjob, record.StartTime)
		case "app":
			err = u.handleApp(*cronjob, record.StartTime)
		case "database":
			err = u.handleDatabase(*cronjob, record.StartTime)
		case "directory":
			if len(cronjob.SourceDir) == 0 {
				return
			}
			err = u.handleDirectory(*cronjob, record.StartTime)
		case "log":
			err = u.handleSystemLog(*cronjob, record.StartTime)
		case "snapshot":
			_ = cronjobRepo.UpdateRecords(record.ID, map[string]interface{}{"records": record.Records})
			err = u.handleSnapshot(*cronjob, record.StartTime)
		}

		if err != nil {
			if len(message) != 0 {
				record.Records, _ = mkdirAndWriteFile(cronjob, record.StartTime, message)
			}
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), record.Records)
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

func (u *CronjobService) handleShell(cronjob model.Cronjob, logPath string) error {
	if len(cronjob.ContainerName) != 0 {
		command := "sh"
		if len(cronjob.Command) != 0 {
			command = cronjob.Command
		}
		scriptFile, _ := os.ReadFile(cronjob.Script)
		return cmd.ExecShell(logPath, 24*time.Hour, "docker", "exec", cronjob.ContainerName, command, "-c", strings.ReplaceAll(string(scriptFile), "\"", "\\\""))
	}
	if len(cronjob.Executor) == 0 {
		cronjob.Executor = "bash"
	}
	if cronjob.ScriptMode == "input" {
		fileItem := path.Join(global.CONF.System.BaseDir, "1panel", "task", "shell", cronjob.Name, cronjob.Name+".sh")
		_ = os.MkdirAll(path.Dir(fileItem), os.ModePerm)
		shellFile, err := os.OpenFile(fileItem, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer shellFile.Close()
		if _, err := shellFile.WriteString(cronjob.Script); err != nil {
			return err
		}
		if len(cronjob.User) == 0 {
			return cmd.ExecShell(logPath, 24*time.Hour, cronjob.Executor, fileItem)
		}
		return cmd.ExecShell(logPath, 24*time.Hour, "sudo", "-u", cronjob.User, cronjob.Executor, fileItem)
	}
	if len(cronjob.User) == 0 {
		return cmd.ExecShell(logPath, 24*time.Hour, cronjob.Executor, cronjob.Script)
	}
	return cmd.ExecShell(logPath, 24*time.Hour, "sudo", "-u", cronjob.User, cronjob.Executor, cronjob.Script)
}

func (u *CronjobService) handleNtpSync() error {
	ntpServer, err := settingRepo.Get(settingRepo.WithByKey("NtpSite"))
	if err != nil {
		return err
	}
	ntime, err := ntp.GetRemoteTime(ntpServer.Value)
	if err != nil {
		return err
	}
	if err := ntp.UpdateSystemTime(ntime.Format(constant.DateTimeLayout)); err != nil {
		return err
	}
	return nil
}

func handleTar(sourceDir, targetDir, name, exclusionRules string, secret string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	excludes := strings.Split(exclusionRules, ",")
	excludeRules := ""
	excludes = append(excludes, "*.sock")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		excludeRules += " --exclude " + exclude
	}
	path := ""
	if strings.Contains(sourceDir, "/") {
		itemDir := strings.ReplaceAll(sourceDir[strings.LastIndex(sourceDir, "/"):], "/", "")
		aheadDir := sourceDir[:strings.LastIndex(sourceDir, "/")]
		if len(aheadDir) == 0 {
			aheadDir = "/"
		}
		path += fmt.Sprintf("-C %s %s", aheadDir, itemDir)
	} else {
		path = sourceDir
	}

	commands := ""

	if len(secret) != 0 {
		extraCmd := "| openssl enc -aes-256-cbc -salt -k '" + secret + "' -out"
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf %s %s %s %s", " -"+excludeRules, path, extraCmd, targetDir+"/"+name)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf %s %s %s", targetDir+"/"+name, excludeRules, path)
		global.LOG.Debug(commands)
	}
	stdout, err := cmd.ExecWithTimeOut(commands, 24*time.Hour)
	if err != nil {
		if len(stdout) != 0 {
			global.LOG.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
			return fmt.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
		}
	}
	return nil
}

func handleUnTar(sourceFile, targetDir string, secret string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}
	commands := ""
	if len(secret) != 0 {
		extraCmd := "openssl enc -d -aes-256-cbc -k '" + secret + "' -in " + sourceFile + " | "
		commands = fmt.Sprintf("%s tar -zxvf - -C %s", extraCmd, targetDir+" > /dev/null 2>&1")
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar zxvfC %s %s", sourceFile, targetDir)
		global.LOG.Debug(commands)
	}

	stdout, err := cmd.ExecWithTimeOut(commands, 24*time.Hour)
	if err != nil {
		global.LOG.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
	}
	return nil
}

func (u *CronjobService) handleCutWebsiteLog(cronjob *model.Cronjob, startTime time.Time) ([]string, string, error) {
	var (
		err       error
		filePaths []string
		msgs      []string
	)
	websites := loadWebsForJob(*cronjob)
	nginx, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return msgs, "", nil
	}
	baseDir := path.Join(nginx.GetPath(), "www", "sites")
	fileOp := files.NewFileOp()
	for _, website := range websites {
		websiteLogDir := path.Join(baseDir, website.Alias, "log")
		srcAccessLogPath := path.Join(websiteLogDir, "access.log")
		srcErrorLogPath := path.Join(websiteLogDir, "error.log")
		dstLogDir := path.Join(global.CONF.System.Backup, "log", "website", website.Alias)
		if !fileOp.Stat(dstLogDir) {
			_ = os.MkdirAll(dstLogDir, 0755)
		}

		dstName := fmt.Sprintf("%s_log_%s.gz", website.PrimaryDomain, startTime.Format(constant.DateTimeSlimLayout))
		dstFilePath := path.Join(dstLogDir, dstName)
		filePaths = append(filePaths, dstFilePath)

		if err = backupLogFile(dstFilePath, websiteLogDir, fileOp); err != nil {
			websiteErr := buserr.WithNameAndErr("ErrCutWebsiteLog", website.PrimaryDomain, err)
			err = websiteErr
			msgs = append(msgs, websiteErr.Error())
			global.LOG.Error(websiteErr.Error())
			continue
		} else {
			_ = fileOp.WriteFile(srcAccessLogPath, strings.NewReader(""), 0755)
			_ = fileOp.WriteFile(srcErrorLogPath, strings.NewReader(""), 0755)
		}
		msg := i18n.GetMsgWithMap("CutWebsiteLogSuccess", map[string]interface{}{"name": website.PrimaryDomain, "path": dstFilePath})
		global.LOG.Infof(msg)
		msgs = append(msgs, msg)
	}
	u.removeExpiredLog(*cronjob)
	return msgs, strings.Join(filePaths, ","), err
}

func backupLogFile(dstFilePath, websiteLogDir string, fileOp files.FileOp) error {
	if err := cmd.ExecCmd(fmt.Sprintf("tar -czf %s -C %s %s", dstFilePath, websiteLogDir, strings.Join([]string{"access.log", "error.log"}, " "))); err != nil {
		dstDir := path.Dir(dstFilePath)
		if err = fileOp.Copy(path.Join(websiteLogDir, "access.log"), dstDir); err != nil {
			return err
		}
		if err = fileOp.Copy(path.Join(websiteLogDir, "error.log"), dstDir); err != nil {
			return err
		}
		if err = cmd.ExecCmd(fmt.Sprintf("tar -czf %s -C %s %s", dstFilePath, dstDir, strings.Join([]string{"access.log", "error.log"}, " "))); err != nil {
			return err
		}
		_ = fileOp.DeleteFile(path.Join(dstDir, "access.log"))
		_ = fileOp.DeleteFile(path.Join(dstDir, "error.log"))
		return nil
	}
	return nil
}

func (u *CronjobService) handleSystemClean() (string, error) {
	return NewIDeviceService().CleanForCronjob()
}

func (u *CronjobService) uploadCronjobBackFile(cronjob model.Cronjob, accountMap map[string]backupClientHelper, file string) (string, error) {
	defer func() {
		_ = os.Remove(file)
	}()
	accounts := strings.Split(cronjob.SourceAccountIDs, ",")
	cloudSrc := strings.TrimPrefix(file, global.CONF.System.TmpDir+"/")
	for _, account := range accounts {
		if len(account) != 0 {
			global.LOG.Debugf("start upload file to %s, dir: %s", accountMap[account].name, path.Join(accountMap[account].backupPath, cloudSrc))
			if _, err := accountMap[account].client.Upload(file, path.Join(accountMap[account].backupPath, cloudSrc)); err != nil {
				return "", err
			}
			global.LOG.Debugf("upload successful!")
		}
	}
	return cloudSrc, nil
}

func (u *CronjobService) removeExpiredBackup(cronjob model.Cronjob, accountMap map[string]backupClientHelper, record model.BackupRecord) {
	global.LOG.Infof("start to handle remove expired, retain copies: %d", cronjob.RetainCopies)
	var opts []repo.DBOption
	opts = append(opts, commonRepo.WithByFrom("cronjob"))
	opts = append(opts, backupRepo.WithByCronID(cronjob.ID))
	opts = append(opts, commonRepo.WithOrderBy("created_at desc"))
	if record.ID != 0 {
		opts = append(opts, commonRepo.WithByType(record.Type))
		opts = append(opts, commonRepo.WithByName(record.Name))
		opts = append(opts, commonRepo.WithByDetailName(record.DetailName))
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
					_, _ = accountMap[account].client.Delete(path.Join(accountMap[account].backupPath, "system_snapshot", records[i].FileName))
				}
			}
			_ = snapshotRepo.Delete(commonRepo.WithByName(strings.TrimSuffix(records[i].FileName, ".tar.gz")))
		} else {
			for _, account := range accounts {
				if len(account) != 0 {
					_, _ = accountMap[account].client.Delete(path.Join(accountMap[account].backupPath, records[i].FileDir, records[i].FileName))
				}
			}
		}
		_ = backupRepo.DeleteRecord(context.Background(), commonRepo.WithByID(records[i].ID))
	}
}

func (u *CronjobService) removeExpiredLog(cronjob model.Cronjob) {
	global.LOG.Infof("start to handle remove expired, retain copies: %d", cronjob.RetainCopies)
	records, _ := cronjobRepo.ListRecord(cronjobRepo.WithByJobID(int(cronjob.ID)), commonRepo.WithOrderBy("created_at desc"))
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
		_ = cronjobRepo.DeleteRecord(commonRepo.WithByID(uint(records[i].ID)))
		_ = os.Remove(records[i].Records)
	}
}

func (u *CronjobService) generateLogsPath(cronjob model.Cronjob, startTime time.Time) string {
	dir := fmt.Sprintf("%s/task/%s/%s", constant.DataDir, cronjob.Type, cronjob.Name)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(dir, os.ModePerm)
	}

	path := fmt.Sprintf("%s/%s.log", dir, startTime.Format(constant.DateTimeSlimLayout))
	return path
}

func hasBackup(cronjobType string) bool {
	return cronjobType == "app" || cronjobType == "database" || cronjobType == "website" || cronjobType == "directory" || cronjobType == "snapshot" || cronjobType == "log"
}
