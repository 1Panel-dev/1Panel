package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/docker/docker/api/types/container"
	_ "github.com/go-sql-driver/mysql"
)

type RedisService struct{}

type IRedisService interface {
	UpdateConf(req dto.RedisConfUpdate) error
	UpdatePersistenceConf(req dto.RedisConfPersistenceUpdate) error
	ChangePassword(info dto.ChangeRedisPass) error

	LoadStatus(req dto.OperationWithName) (*dto.RedisStatus, error)
	LoadConf(req dto.OperationWithName) (*dto.RedisConf, error)
	LoadPersistenceConf(req dto.OperationWithName) (*dto.RedisPersistence, error)

	CheckHasCli() bool
	InstallCli() error
}

func NewIRedisService() IRedisService {
	return &RedisService{}
}

func (u *RedisService) UpdateConf(req dto.RedisConfUpdate) error {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", req.Database)
	if err != nil {
		return err
	}

	var confs []redisConfig
	confs = append(confs, redisConfig{key: "timeout", value: req.Timeout})
	confs = append(confs, redisConfig{key: "maxclients", value: req.Maxclients})
	confs = append(confs, redisConfig{key: "maxmemory", value: req.Maxmemory})
	if err := confSet(redisInfo.Name, "", confs); err != nil {
		return err
	}
	if _, err := compose.Restart(fmt.Sprintf("%s/redis/%s/docker-compose.yml", constant.AppInstallDir, redisInfo.Name)); err != nil {
		return err
	}

	return nil
}

func (u *RedisService) CheckHasCli() bool {
	client, err := docker.NewDockerClient()
	if err != nil {
		return false
	}
	defer client.Close()
	containerLists, err := client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return false
	}
	for _, item := range containerLists {
		if strings.ReplaceAll(item.Names[0], "/", "") == "1Panel-redis-cli-tools" {
			return true
		}
	}
	return false
}

func (u *RedisService) InstallCli() error {
	item := dto.ContainerOperate{
		Name:    "1Panel-redis-cli-tools",
		Image:   "redis:7.2.4",
		Network: "1panel-network",
	}
	return NewIContainerService().ContainerCreate(item)
}

func (u *RedisService) ChangePassword(req dto.ChangeRedisPass) error {
	if err := updateInstallInfoInDB("redis", req.Database, "password", req.Value); err != nil {
		return err
	}
	remote, err := databaseRepo.Get(commonRepo.WithByName(req.Database))
	if err != nil {
		return err
	}
	if remote.From == "local" {
		pass, err := encrypt.StringEncrypt(req.Value)
		if err != nil {
			return fmt.Errorf("decrypt database password failed, err: %v", err)
		}
		_ = databaseRepo.Update(remote.ID, map[string]interface{}{"password": pass})
	}

	return nil
}

func (u *RedisService) UpdatePersistenceConf(req dto.RedisConfPersistenceUpdate) error {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", req.Database)
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
	if err := confSet(redisInfo.Name, req.Type, confs); err != nil {
		return err
	}
	if _, err := compose.Restart(fmt.Sprintf("%s/redis/%s/docker-compose.yml", constant.AppInstallDir, redisInfo.Name)); err != nil {
		return err
	}

	return nil
}

func (u *RedisService) LoadStatus(req dto.OperationWithName) (*dto.RedisStatus, error) {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", req.Name)
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

func (u *RedisService) LoadConf(req dto.OperationWithName) (*dto.RedisConf, error) {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", req.Name)
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

func (u *RedisService) LoadPersistenceConf(req dto.OperationWithName) (*dto.RedisPersistence, error) {
	redisInfo, err := appInstallRepo.LoadBaseInfo("redis", req.Name)
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

func confSet(redisName string, updateType string, changeConf []redisConfig) error {
	path := fmt.Sprintf("%s/redis/%s/conf/redis.conf", constant.AppInstallDir, redisName)
	lineBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	files := strings.Split(string(lineBytes), "\n")

	startIndex, endIndex, emptyLine := 0, 0, 0
	var newFiles []string
	for i := 0; i < len(files); i++ {
		if files[i] == "# Redis configuration rewrite by 1Panel" {
			startIndex = i
			newFiles = append(newFiles, files[i])
			continue
		}
		if files[i] == "# End Redis configuration rewrite by 1Panel" {
			endIndex = i
			break
		}
		if startIndex == 0 && strings.HasPrefix(files[i], "save ") {
			newFiles = append(newFiles, "#   "+files[i])
			continue
		}
		if startIndex != 0 && endIndex == 0 && (len(files[i]) == 0 || (updateType == "rbd" && strings.HasPrefix(files[i], "save "))) {
			emptyLine++
			continue
		}
		newFiles = append(newFiles, files[i])
	}
	endIndex = endIndex - emptyLine
	for _, item := range changeConf {
		if item.key == "save" {
			saveVal := strings.Split(item.value, ",")
			for i := 0; i < len(saveVal); i++ {
				newFiles = append(newFiles, "save "+saveVal[i])
			}
			continue
		}

		isExist := false
		for i := startIndex; i < endIndex; i++ {
			if strings.HasPrefix(newFiles[i], item.key) || strings.HasPrefix(newFiles[i], "# "+item.key) {
				if item.value == "0" || len(item.value) == 0 {
					newFiles[i] = fmt.Sprintf("# %s %s", item.key, item.value)
				} else {
					newFiles[i] = fmt.Sprintf("%s %s", item.key, item.value)
				}
				isExist = true
				break
			}
		}
		if !isExist {
			if item.value == "0" || len(item.value) == 0 {
				newFiles = append(newFiles, fmt.Sprintf("# %s %s", item.key, item.value))
			} else {
				newFiles = append(newFiles, fmt.Sprintf("%s %s", item.key, item.value))
			}
		}
	}
	newFiles = append(newFiles, "# End Redis configuration rewrite by 1Panel")

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
