package service

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
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
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	dockerImage "github.com/docker/docker/api/types/image"
	dockerClient "github.com/docker/docker/client"
)

func (u *BackupService) MongodbBackup(req dto.CommonBackup) error {
	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	itemDir := fmt.Sprintf("database/%s/%s/%s", req.Type, req.Name, req.DetailName)
	targetDir := path.Join(global.Dir.LocalBackupDir, itemDir)
	fileName := fmt.Sprintf("%s_%s.gz", req.DetailName, timeNow+common.RandStrAndNum(5))

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

	if err := handleMongodbBackup(req, nil, record.ID, targetDir, fileName); err != nil {
		backupRepo.UpdateRecordByMap(record.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
		return err
	}
	return nil
}

func (u *BackupService) MongodbRecover(req dto.CommonRecover) error {
	return handleMongodbRecover(req, nil, false)
}

func (u *BackupService) MongodbRecoverByUpload(req dto.CommonRecover) error {
	return handleMongodbRecover(req, nil, false)
}

func handleMongodbBackup(req dto.CommonBackup, parentTask *task.Task, recordID uint, targetDir, fileName string) error {
	dbItem, err := mongodbRepo.Get(repo.WithByName(req.DetailName), mongodbRepo.WithByMongodbName(req.Name))
	if err != nil {
		return err
	}
	itemName := fmt.Sprintf("%s[%s] - %s", req.Name, req.Type, req.DetailName)
	backupTask := parentTask
	if backupTask == nil {
		backupTask, err = task.NewTaskWithOps(itemName, task.TaskBackup, task.TaskScopeBackup, req.TaskID, dbItem.ID)
		if err != nil {
			return err
		}
	}

	itemHandler := func(t *task.Task) error {
		return doMongodbBackup(req.Name, req.Type, req.DetailName, targetDir, fileName, req.Secret, t)
	}
	if parentTask != nil {
		return itemHandler(parentTask)
	}

	backupTask.AddSubTaskWithOps(
		task.GetTaskName(itemName, task.TaskBackup, task.TaskScopeBackup),
		func(t *task.Task) error { return itemHandler(t) },
		nil,
		0,
		3*time.Hour,
	)
	go func() {
		if err := backupTask.Execute(); err != nil {
			backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			return
		}
		backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusSuccess})
	}()
	return nil
}

func handleMongodbRecover(req dto.CommonRecover, parentTask *task.Task, isRollback bool) error {
	dbItem, err := mongodbRepo.Get(repo.WithByName(req.DetailName), mongodbRepo.WithByMongodbName(req.Name))
	if err != nil {
		return err
	}
	itemName := fmt.Sprintf("%s[%s] - %s", req.Name, req.Type, req.DetailName)
	recoverTask := parentTask
	if recoverTask == nil {
		recoverTask, err = task.NewTaskWithOps(itemName, task.TaskRecover, task.TaskScopeBackup, req.TaskID, dbItem.ID)
		if err != nil {
			return err
		}
	}

	recoverDatabase := func(t *task.Task) error {
		fileOp := files.NewFileOp()
		if !fileOp.Stat(req.File) {
			return buserr.WithName("ErrFileNotFound", req.File)
		}

		restoreFile := req.File
		if len(req.Secret) != 0 {
			if err := files.OpensslDecrypt(req.File, req.Secret); err != nil {
				return err
			}
			restoreFile = path.Join(path.Dir(req.File), "tmp_"+path.Base(req.File))
			defer os.Remove(restoreFile)
			t.LogWithStatus(i18n.GetMsgByKey("Decrypt"), nil)
		}

		isOk := false
		if !isRollback {
			rollbackFile := path.Join(
				global.Dir.TmpDir,
				fmt.Sprintf("database/%s/%s_%s.gz", req.Type, req.DetailName, time.Now().Format(constant.DateTimeSlimLayout)),
			)
			if err := doMongodbBackup(req.Name, req.Type, req.DetailName, path.Dir(rollbackFile), path.Base(rollbackFile), "", t); err != nil {
				return fmt.Errorf("backup mongodb db %s for rollback before recover failed, err: %v", req.DetailName, err)
			}
			defer func() {
				if !isOk {
					global.LOG.Info("recover failed, start to rollback now")
					if err := doMongodbRestore(req.Name, req.Type, req.DetailName, rollbackFile, t); err != nil {
						global.LOG.Errorf("rollback mongodb db %s from %s failed, err: %v", req.DetailName, rollbackFile, err)
					} else {
						global.LOG.Infof("rollback mongodb db %s from %s successful", req.DetailName, rollbackFile)
					}
				}
				_ = os.RemoveAll(rollbackFile)
			}()
		}

		if err := doMongodbRestore(req.Name, req.Type, req.DetailName, restoreFile, t); err != nil {
			global.LOG.Errorf("recover mongodb db %s from %s failed, err: %v", req.DetailName, restoreFile, err)
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
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("TaskRecover"), recoverDatabase, nil, 0, timeout)
	go func() {
		_ = recoverTask.Execute()
	}()
	return nil
}

func doMongodbBackup(database, dbType, dbName, targetDir, fileName, secret string, taskItem *task.Task) error {
	dbItem, err := mongodbRepo.Get(repo.WithByName(dbName), mongodbRepo.WithByMongodbName(database))
	if err == nil && dbItem.From == constant.AppResourceRemote {
		if err := doRemoteMongodbBackup(database, dbName, targetDir, fileName, taskItem); err != nil {
			return err
		}
		if len(secret) != 0 {
			return files.OpensslEncrypt(path.Join(targetDir, fileName), secret)
		}
		return nil
	}

	appInfo, err := appInstallRepo.LoadBaseInfo(dbType, database)
	if err != nil {
		return err
	}
	if appInfo.ContainerName == "" {
		return fmt.Errorf("mongodb container not found for database %s", database)
	}
	if err := os.MkdirAll(targetDir, constant.DirPerm); err != nil {
		return err
	}

	targetFile := path.Join(targetDir, fileName)
	containerFile := path.Join("/tmp", fileName)
	defer func() {
		_ = cmd.NewCommandMgr().Run("docker", "exec", appInfo.ContainerName, "rm", "-f", containerFile)
	}()

	uri := buildMongodbDumpURI(appInfo.UserName, appInfo.Password, dbName)
	cmdMgr := mongodbCmdMgr(taskItem)
	if err := cmdMgr.Run(
		"docker",
		"exec",
		appInfo.ContainerName,
		"mongodump",
		"--uri="+uri,
		"--archive="+containerFile,
		"--gzip",
	); err != nil {
		return err
	}
	if err := cmdMgr.Run("docker", "cp", fmt.Sprintf("%s:%s", appInfo.ContainerName, containerFile), targetFile); err != nil {
		return err
	}
	if len(secret) != 0 {
		return files.OpensslEncrypt(targetFile, secret)
	}
	return nil
}

func doMongodbRestore(database, dbType, dbName, sourceFile string, taskItem *task.Task) error {
	dbItem, err := mongodbRepo.Get(repo.WithByName(dbName), mongodbRepo.WithByMongodbName(database))
	if err == nil && dbItem.From == constant.AppResourceRemote {
		return doRemoteMongodbRestore(database, dbName, sourceFile, taskItem)
	}

	appInfo, err := appInstallRepo.LoadBaseInfo(dbType, database)
	if err != nil {
		return err
	}
	if appInfo.ContainerName == "" {
		return fmt.Errorf("mongodb container not found for database %s", database)
	}

	containerFile := path.Join("/tmp", fmt.Sprintf("1panel-mongodb-restore-%s.gz", common.RandStrAndNum(8)))
	defer func() {
		_ = cmd.NewCommandMgr().Run("docker", "exec", appInfo.ContainerName, "rm", "-f", containerFile)
	}()

	cmdMgr := mongodbCmdMgr(taskItem)
	if err := cmdMgr.Run("docker", "cp", sourceFile, fmt.Sprintf("%s:%s", appInfo.ContainerName, containerFile)); err != nil {
		return err
	}

	uri := buildMongodbRestoreURI(appInfo.UserName, appInfo.Password)
	if err := cmdMgr.Run(
		"docker",
		"exec",
		appInfo.ContainerName,
		"mongorestore",
		"--uri="+uri,
		"--nsInclude="+buildMongodbNamespace(sourceFile, dbName),
		"--nsFrom="+buildMongodbNamespace(sourceFile, dbName),
		"--nsTo="+dbName+".*",
		"--archive="+containerFile,
		"--gzip",
		"--drop",
	); err != nil {
		return err
	}
	return nil
}

func buildMongodbDumpURI(username, password, dbName string) string {
	return (&url.URL{
		Scheme:   "mongodb",
		User:     url.UserPassword(username, password),
		Host:     "127.0.0.1:27017",
		Path:     "/" + dbName,
		RawQuery: "authSource=admin",
	}).String()
}

func buildMongodbRestoreURI(username, password string) string {
	return (&url.URL{
		Scheme:   "mongodb",
		User:     url.UserPassword(username, password),
		Host:     "127.0.0.1:27017",
		Path:     "/",
		RawQuery: "authSource=admin",
	}).String()
}

func mongodbCmdMgr(taskItem *task.Task) *cmd.CommandHelper {
	if taskItem == nil {
		return cmd.NewCommandMgr(cmd.WithTimeout(3 * time.Hour))
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(3*time.Hour), cmd.WithTask(*taskItem))
}

func doRemoteMongodbBackup(database, dbName, targetDir, fileName string, taskItem *task.Task) error {
	info, err := loadRemoteMongodbConnection(database)
	if err != nil {
		return err
	}
	imageTag, err := ensureMongodbImage(database, taskItem)
	if err != nil {
		return err
	}
	logRemoteMongodbImage(taskItem, "backup", database, dbName, imageTag, info)
	logRemoteMongodbStep(taskItem, fmt.Sprintf("local image %s is ready, start backup", imageTag))
	if err := os.MkdirAll(targetDir, constant.DirPerm); err != nil {
		return err
	}

	targetFile, err := os.OpenFile(path.Join(targetDir, fileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, constant.DirPerm)
	if err != nil {
		return fmt.Errorf("open file %s failed, err: %v", path.Join(targetDir, fileName), err)
	}
	defer func() { _ = targetFile.Close() }()

	backupCmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"--net=host",
		"-i",
		imageTag,
		"mongodump",
		"--uri="+buildRemoteMongodbURI(info),
		"--db="+dbName,
		"--archive",
		"--gzip",
	)
	backupCmd.Stdout = targetFile
	var stderr bytes.Buffer
	backupCmd.Stderr = &stderr
	if err := backupCmd.Run(); err != nil {
		return fmt.Errorf("handle backup mongodb database failed, err: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

func doRemoteMongodbRestore(database, dbName, sourceFile string, taskItem *task.Task) error {
	info, err := loadRemoteMongodbConnection(database)
	if err != nil {
		return err
	}
	imageTag, err := ensureMongodbImage(database, taskItem)
	if err != nil {
		return err
	}
	logRemoteMongodbImage(taskItem, "restore", database, dbName, imageTag, info)
	logRemoteMongodbStep(taskItem, fmt.Sprintf("local image %s is ready, start restore", imageTag))

	fi, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer func() { _ = fi.Close() }()

	restoreCmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"--net=host",
		"-i",
		imageTag,
		"mongorestore",
		"--uri="+buildRemoteMongodbURI(info),
		"--nsInclude="+buildMongodbNamespace(sourceFile, dbName),
		"--nsFrom="+buildMongodbNamespace(sourceFile, dbName),
		"--nsTo="+dbName+".*",
		"--archive",
		"--gzip",
		"--drop",
	)
	restoreCmd.Stdin = fi
	var stderr bytes.Buffer
	restoreCmd.Stderr = &stderr
	if err := restoreCmd.Run(); err != nil {
		return fmt.Errorf("handle recover mongodb database failed, err: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

func ensureMongodbImage(database string, taskItem *task.Task) (string, error) {
	imageTag, exists, err := loadMongodbImageTag(database)
	if err != nil {
		return "", err
	}
	logRemoteMongodbStep(taskItem, fmt.Sprintf("check local image %s", imageTag))
	if exists {
		logRemoteMongodbStep(taskItem, fmt.Sprintf("local image %s exists", imageTag))
		return imageTag, nil
	}
	logRemoteMongodbStep(taskItem, fmt.Sprintf("local image %s not found, start docker pull", imageTag))
	if err := mongodbCmdMgr(taskItem).Run("docker", "pull", imageTag); err != nil {
		return "", err
	}
	logRemoteMongodbStep(taskItem, fmt.Sprintf("docker pull %s finished", imageTag))
	return imageTag, nil
}

func loadMongodbImageTag(database string) (string, bool, error) {
	databaseInfo, err := databaseRepo.Get(repo.WithByName(database))
	if err != nil {
		return "", false, err
	}

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return "", false, err
	}
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), dockerImage.ListOptions{})
	if err != nil {
		return "", false, err
	}
	imagePrefix := "mongo:" + loadMongodbImageMajor(databaseInfo.Version)
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if strings.HasPrefix(tag, imagePrefix) {
				return tag, true, nil
			}
		}
	}
	return imagePrefix, false, nil
}

func buildMongodbNamespace(sourceFile, targetDB string) string {
	sourceDB := loadMongodbBackupDBName(sourceFile, targetDB)
	return sourceDB + ".*"
}

func loadMongodbBackupDBName(sourceFile, defaultDB string) string {
	baseName := path.Base(sourceFile)
	if strings.HasSuffix(baseName, ".gz") {
		baseName = strings.TrimSuffix(baseName, ".gz")
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^1panel_mongodb_(.+)_\d{14}[A-Za-z0-9]*$`),
		regexp.MustCompile(`^db_(.+)_\d{14}[A-Za-z0-9]*$`),
		regexp.MustCompile(`^(.+)_\d{14}[A-Za-z0-9]*$`),
	}
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(baseName)
		if len(matches) == 2 && len(matches[1]) != 0 {
			return matches[1]
		}
	}
	return defaultDB
}

func loadMongodbImageMajor(version string) string {
	switch {
	case strings.HasPrefix(version, "6"):
		return "6"
	case strings.HasPrefix(version, "7"):
		return "7"
	default:
		return "8"
	}
}

func logRemoteMongodbImage(taskItem *task.Task, action, database, dbName, imageTag string, info mongodbConnectionInfo) {
	message := fmt.Sprintf(
		"use local docker image %s to %s remote mongodb %s/%s via %s:%d",
		imageTag,
		action,
		database,
		dbName,
		info.Address,
		info.Port,
	)
	global.LOG.Info(message)
	if taskItem != nil {
		taskItem.Log(message)
	}
}

func logRemoteMongodbStep(taskItem *task.Task, message string) {
	global.LOG.Info(message)
	if taskItem != nil {
		taskItem.Log(message)
	}
}
