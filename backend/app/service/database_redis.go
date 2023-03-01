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

	SearchBackupListWithPage(req dto.PageInfo) (int64, interface{}, error)
}

func NewIRedisService() IRedisService {
	return &RedisService{}
}

func (u *RedisService) UpdateConf(req dto.RedisConfUpdate) error {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
	if err != nil {
		return err
	}

	var confs []redisConfig
	confs = append(confs, redisConfig{key: "timeout", value: req.Timeout})
	confs = append(confs, redisConfig{key: "maxclients", value: req.Maxclients})
	confs = append(confs, redisConfig{key: "maxmemory", value: req.Maxmemory})
	if err := confSet(redisInfo.Name, confs); err != nil {
		return err
	}
	if _, err := compose.Restart(fmt.Sprintf("%s/redis/%s/docker-compose.yml", constant.AppInstallDir, redisInfo.Name)); err != nil {
		return err
	}

	return nil
}

func (u *RedisService) ChangePassword(req dto.ChangeDBInfo) error {
	if err := updateInstallInfoInDB("redis", "", "password", true, req.Value); err != nil {
		return err
	}
	if err := updateInstallInfoInDB("redis-commander", "", "password", true, req.Value); err != nil {
		return err
	}

	return nil
}

func (u *RedisService) UpdatePersistenceConf(req dto.RedisConfPersistenceUpdate) error {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
	if err != nil {
		return err
	}

	var confs []redisConfig
	if req.Type == "rbd" {
		confs = append(confs, redisConfig{key: "save", value: req.Save})
	} else {
		confs = append(confs, redisConfig{key: "appendonly", value: req.Appendonly})
		confs = append(confs, redisConfig{key: "appendfsync", value: req.Appendfsync})
	}
	if err := confSet(redisInfo.Name, confs); err != nil {
		return err
	}
	if _, err := compose.Restart(fmt.Sprintf("%s/redis/%s/docker-compose.yml", constant.AppInstallDir, redisInfo.Name)); err != nil {
		return err
	}

	return nil
}

func (u *RedisService) LoadStatus() (*dto.RedisStatus, error) {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
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
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
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
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
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

func (u *RedisService) SearchBackupListWithPage(req dto.PageInfo) (int64, interface{}, error) {
	var (
		list      []dto.DatabaseFileRecords
		backDatas []dto.DatabaseFileRecords
	)
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", "")
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

type redisConfig struct {
	key   string
	value string
}

func confSet(redisName string, changeConf []redisConfig) error {
	path := fmt.Sprintf("%s/redis/%s/conf/redis.conf", constant.AppInstallDir, redisName)
	lineBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	files := strings.Split(string(lineBytes), "\n")

	isStartRange := false
	isEndRange := false
	var newFiles []string
	for _, line := range files {
		if !isStartRange {
			if line == "# Redis configuration rewrite by 1Panel" {
				isStartRange = true
			}
			newFiles = append(newFiles, line)
			continue
		}
		if !isEndRange {
			isExist := false
			for _, item := range changeConf {
				if strings.HasPrefix(line, item.key) || strings.HasPrefix(line, "# "+item.key) {
					if item.value == "0" || len(item.value) == 0 {
						newFiles = append(newFiles, fmt.Sprintf("# %s %s", item.key, item.value))
					} else {
						newFiles = append(newFiles, fmt.Sprintf("%s %s", item.key, item.value))
					}
					isExist = true
					break
				}
			}
			if isExist {
				continue
			}
			newFiles = append(newFiles, line)
			if line == "# End Redis configuration rewrite by 1Panel" {
				isEndRange = true
			}
			continue
		}
		newFiles = append(newFiles, line)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(newFiles, "\n"))
	if err != nil {
		return err
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
