package service

import (
	"fmt"
	"os"
	"path"
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
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) RedisBackup(req dto.CommonBackup) error {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", req.Name)
	if err != nil {
		return err
	}
	appendonly, err := configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly")
	if err != nil {
		return err
	}
	global.LOG.Infof("appendonly in redis conf is %s", appendonly)

	timeNow := time.Now().Format(constant.DateTimeSlimLayout) + common.RandStrAndNum(5)
	fileName := fmt.Sprintf("%s.rdb", timeNow)
	if appendonly == "yes" {
		if strings.HasPrefix(redisInfo.Version, "6.") {
			fileName = fmt.Sprintf("%s.aof", timeNow)
		} else {
			fileName = fmt.Sprintf("%s.tar.gz", timeNow)
		}
	}
	itemDir := fmt.Sprintf("database/redis/%s", redisInfo.Name)
	backupDir := path.Join(global.CONF.System.Backup, itemDir)
	if err := handleRedisBackup(redisInfo, nil, backupDir, fileName, req.Secret, req.TaskID); err != nil {
		return err
	}
	record := &model.BackupRecord{
		Type:              "redis",
		Name:              req.Name,
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

func (u *BackupService) RedisRecover(req dto.CommonRecover) error {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", req.Name)
	if err != nil {
		return err
	}
	global.LOG.Infof("recover redis from backup file %s", req.File)
	if err := handleRedisRecover(redisInfo, nil, req.File, false, req.Secret, req.TaskID); err != nil {
		return err
	}
	return nil
}

func handleRedisBackup(redisInfo *repo.RootInfo, parentTask *task.Task, backupDir, fileName, secret, taskID string) error {
	var (
		err      error
		itemTask *task.Task
	)
	itemTask = parentTask
	if parentTask == nil {
		itemTask, err = task.NewTaskWithOps("Redis", task.TaskBackup, task.TaskScopeDatabase, taskID, redisInfo.ID)
		if err != nil {
			return err
		}
	}

	backupDatabase := func(t *task.Task) error {
		fileOp := files.NewFileOp()
		if !fileOp.Stat(backupDir) {
			if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", backupDir, err)
			}
		}

		stdout, err := cmd.Execf("docker exec %s redis-cli -a %s --no-auth-warning save", redisInfo.ContainerName, redisInfo.Password)
		if err != nil {
			return errors.New(string(stdout))
		}

		if strings.HasSuffix(fileName, ".tar.gz") {
			redisDataDir := fmt.Sprintf("%s/%s/%s/data/appendonlydir", constant.AppInstallDir, "redis", redisInfo.Name)
			if err := handleTar(redisDataDir, backupDir, fileName, "", secret); err != nil {
				return err
			}
			return nil
		}
		if strings.HasSuffix(fileName, ".aof") {
			stdout1, err := cmd.Execf("docker cp %s:/data/appendonly.aof %s/%s", redisInfo.ContainerName, backupDir, fileName)
			if err != nil {
				return errors.New(string(stdout1))
			}
			return nil
		}

		stdout1, err1 := cmd.Execf("docker cp %s:/data/dump.rdb %s/%s", redisInfo.ContainerName, backupDir, fileName)
		if err1 != nil {
			return errors.New(string(stdout1))
		}
		return nil
	}
	itemTask.AddSubTask(i18n.GetMsgByKey("TaskBackup"), backupDatabase, nil)
	if parentTask != nil {
		return backupDatabase(parentTask)
	}

	return itemTask.Execute()
}

func handleRedisRecover(redisInfo *repo.RootInfo, parentTask *task.Task, recoverFile string, isRollback bool, secret, taskID string) error {
	var (
		err      error
		itemTask *task.Task
	)
	itemTask = parentTask
	if parentTask == nil {
		itemTask, err = task.NewTaskWithOps("Redis", task.TaskRecover, task.TaskScopeDatabase, taskID, redisInfo.ID)
		if err != nil {
			return err
		}
	}

	recoverDatabase := func(t *task.Task) error {
		fileOp := files.NewFileOp()
		if !fileOp.Stat(recoverFile) {
			return buserr.WithName("ErrFileNotFound", recoverFile)
		}

		appendonly, err := configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly")
		if err != nil {
			return err
		}

		if appendonly == "yes" {
			if strings.HasPrefix(redisInfo.Version, "6.") && !strings.HasSuffix(recoverFile, ".aof") {
				return buserr.New(constant.ErrTypeOfRedis)
			}
			if strings.HasPrefix(redisInfo.Version, "7.") && !strings.HasSuffix(recoverFile, ".tar.gz") {
				return buserr.New(constant.ErrTypeOfRedis)
			}
		} else {
			if !strings.HasSuffix(recoverFile, ".rdb") {
				return buserr.New(constant.ErrTypeOfRedis)
			}
		}

		global.LOG.Infof("appendonly in redis conf is %s", appendonly)
		isOk := false
		if !isRollback {
			suffix := "rdb"
			if appendonly == "yes" {
				if strings.HasPrefix(redisInfo.Version, "6.") {
					suffix = "aof"
				} else {
					suffix = "tar.gz"
				}
			}
			rollbackFile := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("database/redis/%s_%s.%s", redisInfo.Name, time.Now().Format(constant.DateTimeSlimLayout), suffix))
			if err := handleRedisBackup(redisInfo, nil, path.Dir(rollbackFile), path.Base(rollbackFile), secret, ""); err != nil {
				return fmt.Errorf("backup database %s for rollback before recover failed, err: %v", redisInfo.Name, err)
			}
			defer func() {
				if !isOk {
					global.LOG.Info("recover failed, start to rollback now")
					if err := handleRedisRecover(redisInfo, itemTask, rollbackFile, true, secret, ""); err != nil {
						global.LOG.Errorf("rollback redis from %s failed, err: %v", rollbackFile, err)
						return
					}
					global.LOG.Infof("rollback redis from %s successful", rollbackFile)
					_ = os.RemoveAll(rollbackFile)
				} else {
					_ = os.RemoveAll(rollbackFile)
				}
			}()
		}
		composeDir := fmt.Sprintf("%s/redis/%s", constant.AppInstallDir, redisInfo.Name)
		if _, err := compose.Down(composeDir + "/docker-compose.yml"); err != nil {
			return err
		}
		if appendonly == "yes" && strings.HasPrefix(redisInfo.Version, "7.") {
			redisDataDir := fmt.Sprintf("%s/%s/%s/data", constant.AppInstallDir, "redis", redisInfo.Name)
			if err := handleUnTar(recoverFile, redisDataDir, secret); err != nil {
				return err
			}
		} else {
			itemName := "dump.rdb"
			if appendonly == "yes" && strings.HasPrefix(redisInfo.Version, "6.") {
				itemName = "appendonly.aof"
			}
			input, err := os.ReadFile(recoverFile)
			if err != nil {
				return err
			}
			if err = os.WriteFile(composeDir+"/data/"+itemName, input, 0640); err != nil {
				return err
			}
		}
		if _, err := compose.Up(composeDir + "/docker-compose.yml"); err != nil {
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
