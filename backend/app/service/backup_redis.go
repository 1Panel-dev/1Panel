package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) RedisBackup() error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
	if err != nil {
		return err
	}
	appendonly, err := configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly")
	if err != nil {
		return err
	}
	global.LOG.Infof("appendonly in redis conf is %s", appendonly)

	timeNow := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s.rdb", timeNow)
	if appendonly == "yes" {
		fileName = fmt.Sprintf("%s.tar.gz", timeNow)
	}
	backupDir := fmt.Sprintf("%s/database/redis/%s", localDir, redisInfo.Name)
	if err := handleRedisBackup(redisInfo, backupDir, fileName); err != nil {
		return err
	}
	record := &model.BackupRecord{
		Type:       "redis",
		Source:     "LOCAL",
		BackupType: "LOCAL",
		FileDir:    backupDir,
		FileName:   fileName,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
	}

	return nil
}

func (u *BackupService) RedisRecover(req dto.CommonRecover) error {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
	if err != nil {
		return err
	}
	global.LOG.Infof("recover redis from backup file %s", req.File)
	if err := handleRedisRecover(redisInfo, req.File, false); err != nil {
		return err
	}
	return nil
}

func handleRedisBackup(redisInfo *repo.RootInfo, backupDir, fileName string) error {
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
		if err := handleTar(redisDataDir, backupDir, fileName, ""); err != nil {
			return err
		}
		return nil
	}

	stdout1, err1 := cmd.Execf("docker cp %s:/data/dump.rdb %s/%s", redisInfo.ContainerName, backupDir, fileName)
	if err1 != nil {
		return errors.New(string(stdout1))
	}
	return nil
}

func handleRedisRecover(redisInfo *repo.RootInfo, recoverFile string, isRollback bool) error {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(recoverFile) {
		return fmt.Errorf("%s file is not exist", recoverFile)
	}

	appendonly, err := configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly")
	if err != nil {
		return err
	}
	if (appendonly == "yes" && !strings.HasSuffix(recoverFile, ".tar.gz")) || (appendonly != "yes" && !strings.HasSuffix(recoverFile, ".rdb")) {
		return fmt.Errorf("recover redis with appendonly=%s from file %s format error", appendonly, recoverFile)
	}
	global.LOG.Infof("appendonly in redis conf is %s", appendonly)
	isOk := false
	if !isRollback {
		suffix := "tar.gz"
		if appendonly != "yes" {
			suffix = "rdb"
		}
		rollbackFile := fmt.Sprintf("%s/original/database/redis/%s_%s.%s", global.CONF.System.BaseDir, redisInfo.Name, time.Now().Format("20060102150405"), suffix)
		if err := handleRedisBackup(redisInfo, path.Dir(rollbackFile), path.Base(rollbackFile)); err != nil {
			return fmt.Errorf("backup database %s for rollback before recover failed, err: %v", redisInfo.Name, err)
		}
		defer func() {
			if !isOk {
				global.LOG.Info("recover failed, start to rollback now")
				if err := handleRedisRecover(redisInfo, rollbackFile, true); err != nil {
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
	if appendonly == "yes" {
		redisDataDir := fmt.Sprintf("%s/%s/%s/data/", constant.AppInstallDir, "redis", redisInfo.Name)
		if err := handleUnTar(recoverFile, redisDataDir); err != nil {
			return err
		}
	} else {
		input, err := ioutil.ReadFile(recoverFile)
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(composeDir+"/data/dump.rdb", input, 0640); err != nil {
			return err
		}
	}
	if _, err := compose.Up(composeDir + "/docker-compose.yml"); err != nil {
		return err
	}
	return nil
}
