package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	_ "github.com/go-sql-driver/mysql"
)

type RedisService struct{}

type IRedisService interface {
	UpdateConf(req dto.RedisConfUpdate) error
	UpdatePersistenceConf(req dto.RedisConfPersistenceUpdate) error
	ChangePassword(info dto.ChangeDBInfo) error

	LoadStatus() (*dto.RedisStatus, error)
	LoadConf() (*dto.RedisConf, error)
	LoadPersistenceConf() (*dto.RedisPersistence, error)

	Backup() error
	SearchBackupListWithPage(req dto.PageInfo) (int64, interface{}, error)
	Recover(req dto.RedisBackupRecover) error
}

func NewIRedisService() IRedisService {
	return &RedisService{}
}

func (u *RedisService) UpdateConf(req dto.RedisConfUpdate) error {
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return err
	}
	if err := configSetStr(redisInfo.ContainerName, redisInfo.Password, "timeout", req.Timeout); err != nil {
		return err
	}
	if err := configSetStr(redisInfo.ContainerName, redisInfo.Password, "maxclients", req.Maxclients); err != nil {
		return err
	}

	if err := configSetStr(redisInfo.ContainerName, redisInfo.Password, "maxmemory", req.Maxmemory); err != nil {
		return err
	}

	commands := append(redisExec(redisInfo.ContainerName, redisInfo.Password), []string{"config", "rewrite"}...)
	cmd := exec.Command("docker", commands...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (u *RedisService) ChangePassword(req dto.ChangeDBInfo) error {
	if err := updateInstallInfoInDB("redis", "password", true, req.Value); err != nil {
		return err
	}
	if err := updateInstallInfoInDB("redis-commander", "password", true, req.Value); err != nil {
		return err
	}

	return nil
}

func (u *RedisService) UpdatePersistenceConf(req dto.RedisConfPersistenceUpdate) error {
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return err
	}

	if req.Type == "rbd" {
		if err := configSetStr(redisInfo.ContainerName, redisInfo.Password, "save", req.Save); err != nil {
			return err
		}
	} else {
		if err := configSetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly", req.Appendonly); err != nil {
			return err
		}
		if err := configSetStr(redisInfo.ContainerName, redisInfo.Password, "appendfsync", req.Appendfsync); err != nil {
			return err
		}
	}
	commands := append(redisExec(redisInfo.ContainerName, redisInfo.Password), []string{"config", "rewrite"}...)
	cmd := exec.Command("docker", commands...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (u *RedisService) LoadStatus() (*dto.RedisStatus, error) {
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return nil, err
	}
	commands := append(redisExec(redisInfo.ContainerName, redisInfo.Password), "info")
	cmd := exec.Command("docker", commands...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New(string(stdout))
	}
	rows := strings.Split(string(stdout), "\r\n")
	rowMap := make(map[string]string)
	for _, v := range rows {
		itemRow := strings.Split(v, ":")
		if len(itemRow) == 2 {
			rowMap[itemRow[0]] = itemRow[1]
		}
	}
	var info dto.RedisStatus
	arr, err := json.Marshal(rowMap)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(arr, &info)
	return &info, nil
}

func (u *RedisService) LoadConf() (*dto.RedisConf, error) {
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return nil, err
	}

	var item dto.RedisConf
	item.ContainerName = redisInfo.ContainerName
	item.Name = redisInfo.Name
	item.Port = redisInfo.Port
	item.Requirepass = redisInfo.Password
	item.Timeout, _ = configGetStr(redisInfo.ContainerName, redisInfo.Password, "timeout")
	item.Maxclients, _ = configGetStr(redisInfo.ContainerName, redisInfo.Password, "maxclients")
	item.Maxmemory, _ = configGetStr(redisInfo.ContainerName, redisInfo.Password, "maxmemory")
	return &item, nil
}

func (u *RedisService) LoadPersistenceConf() (*dto.RedisPersistence, error) {
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return nil, err
	}

	var item dto.RedisPersistence
	if item.Appendonly, err = configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly"); err != nil {
		return nil, err
	}
	if item.Appendfsync, err = configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendfsync"); err != nil {
		return nil, err
	}
	if item.Save, err = configGetStr(redisInfo.ContainerName, redisInfo.Password, "save"); err != nil {
		return nil, err
	}
	return &item, nil
}

func (u *RedisService) Backup() error {
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return err
	}
	commands := append(redisExec(redisInfo.ContainerName, redisInfo.Password), "save")
	cmd := exec.Command("docker", commands...)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		return errors.New(string(stdout))
	}
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	backupDir := fmt.Sprintf("database/redis/%s/", redisInfo.Name)
	fullDir := fmt.Sprintf("%s/%s", localDir, backupDir)
	if _, err := os.Stat(fullDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(fullDir, os.ModePerm); err != nil {
			if err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", fullDir, err)
			}
		}
	}

	appendonly, err := configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly")
	if err != nil {
		return err
	}
	if appendonly == "yes" {
		redisDataDir := fmt.Sprintf("%s/%s/%s/data", constant.AppInstallDir, "redis", redisInfo.Name)
		name := fmt.Sprintf("%s.tar.gz", time.Now().Format("20060102150405"))
		if err := handleTar(redisDataDir+"/appendonlydir", fullDir, name, ""); err != nil {
			return err
		}
		return nil
	}

	name := fmt.Sprintf("%s.rdb", time.Now().Format("20060102150405"))
	cmd2 := exec.Command("docker", "cp", fmt.Sprintf("%s:/data/dump.rdb", redisInfo.ContainerName), fmt.Sprintf("%s/%s", fullDir, name))
	if stdout, err := cmd2.CombinedOutput(); err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (u *RedisService) Recover(req dto.RedisBackupRecover) error {
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return err
	}
	appendonly, err := configGetStr(redisInfo.ContainerName, redisInfo.Password, "appendonly")
	if err != nil {
		return err
	}

	composeDir := fmt.Sprintf("%s/redis/%s", constant.AppInstallDir, redisInfo.Name)
	if _, err := compose.Down(composeDir + "/docker-compose.yml"); err != nil {
		return err
	}
	fullName := fmt.Sprintf("%s/%s", req.FileDir, req.FileName)
	if appendonly == "yes" {
		redisDataDir := fmt.Sprintf("%s/%s/%s/data/", constant.AppInstallDir, "redis", redisInfo.Name)
		if err := handleUnTar(fullName, redisDataDir); err != nil {
			return err
		}
	} else {
		input, err := ioutil.ReadFile(fullName)
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

func (u *RedisService) SearchBackupListWithPage(req dto.PageInfo) (int64, interface{}, error) {
	var (
		list      []dto.DatabaseFileRecords
		backDatas []dto.DatabaseFileRecords
	)
	redisInfo, err := appInstallRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return 0, nil, err
	}
	localDir, err := loadLocalDir()
	if err != nil {
		return 0, nil, err
	}
	backupDir := fmt.Sprintf("%s/database/redis/%s", localDir, redisInfo.Name)
	_ = filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			list = append(list, dto.DatabaseFileRecords{
				CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
				Size:      int(info.Size()),
				FileDir:   backupDir,
				FileName:  info.Name(),
			})
		}
		return nil
	})
	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]dto.DatabaseFileRecords, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = list[start:end]
	}
	return int64(total), backDatas, nil
}

func configGetStr(containerName, password, param string) (string, error) {
	commands := append(redisExec(containerName, password), []string{"config", "get", param}...)
	cmd := exec.Command("docker", commands...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(stdout))
	}
	rows := strings.Split(string(stdout), "\r\n")
	for _, v := range rows {
		itemRow := strings.Split(v, "\n")
		if len(itemRow) == 3 {
			return itemRow[1], nil
		}
	}
	return "", nil
}
func configSetStr(containerName, password, param, value string) error {
	commands := append(redisExec(containerName, password), []string{"config", "set", param, value}...)
	cmd := exec.Command("docker", commands...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func redisExec(containerName, password string) []string {
	cmds := []string{"exec", containerName, "redis-cli", "-a", password, "--no-auth-warning"}
	if len(password) == 0 {
		cmds = []string{"exec", containerName, "redis-cli"}
	}
	return cmds
}
