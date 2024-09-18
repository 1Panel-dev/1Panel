package service

import (
	"bufio"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/net/context"
)

const composeProjectLabel = "com.docker.compose.project"
const composeConfigLabel = "com.docker.compose.project.config_files"
const composeWorkdirLabel = "com.docker.compose.project.working_dir"
const composeCreatedBy = "createdBy"

type DockerCompose struct {
	Version  string                            `yaml:"version"`
	Services map[string]map[string]interface{} `yaml:"services"`
	Networks map[string]interface{}            `yaml:"networks"`
}

func (u *ContainerService) PageCompose(req dto.SearchWithPage) (int64, interface{}, error) {
	var (
		records   []dto.ComposeInfo
		BackDatas []dto.ComposeInfo
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	defer client.Close()

	options := container.ListOptions{All: true}
	options.Filters = filters.NewArgs()
	options.Filters.Add("label", composeProjectLabel)

	list, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return 0, nil, err
	}

	composeCreatedByLocal, _ := composeRepo.ListRecord()

	composeLocalMap := make(map[string]dto.ComposeInfo)
	for _, localItem := range composeCreatedByLocal {
		composeItemLocal := dto.ComposeInfo{
			ContainerNumber: 0,
			CreatedAt:       localItem.CreatedAt.Format(constant.DateTimeLayout),
			ConfigFile:      localItem.Path,
			Workdir:         strings.TrimSuffix(localItem.Path, "/docker-compose.yml"),
		}
		composeItemLocal.CreatedBy = "1Panel"
		composeItemLocal.Path = localItem.Path
		composeLocalMap[localItem.Name] = composeItemLocal
	}

	composeMap := make(map[string]dto.ComposeInfo)
	for _, container := range list {
		if name, ok := container.Labels[composeProjectLabel]; ok {
			containerItem := dto.ComposeContainer{
				ContainerID: container.ID,
				Name:        container.Names[0][1:],
				State:       container.State,
				CreateTime:  time.Unix(container.Created, 0).Format(constant.DateTimeLayout),
			}
			if compose, has := composeMap[name]; has {
				compose.ContainerNumber++
				compose.Containers = append(compose.Containers, containerItem)
				composeMap[name] = compose
			} else {
				config := container.Labels[composeConfigLabel]
				workdir := container.Labels[composeWorkdirLabel]
				composeItem := dto.ComposeInfo{
					ContainerNumber: 1,
					CreatedAt:       time.Unix(container.Created, 0).Format(constant.DateTimeLayout),
					ConfigFile:      config,
					Workdir:         workdir,
					Containers:      []dto.ComposeContainer{containerItem},
				}
				createdBy, ok := container.Labels[composeCreatedBy]
				if ok {
					composeItem.CreatedBy = createdBy
				}
				if len(config) != 0 && len(workdir) != 0 && strings.Contains(config, workdir) {
					composeItem.Path = config
				} else {
					composeItem.Path = workdir
				}
				for i := 0; i < len(composeCreatedByLocal); i++ {
					if composeCreatedByLocal[i].Name == name {
						composeItem.CreatedBy = "1Panel"
						composeCreatedByLocal = append(composeCreatedByLocal[:i], composeCreatedByLocal[i+1:]...)
						break
					}
				}
				composeMap[name] = composeItem
			}
		}
	}

	mergedMap := make(map[string]dto.ComposeInfo)
	for key, localItem := range composeLocalMap {
		mergedMap[key] = localItem
	}
	for key, item := range composeMap {
		if existingItem, exists := mergedMap[key]; exists {
			if item.ContainerNumber > 0 {
				if existingItem.ContainerNumber <= 0 {
					mergedMap[key] = item
				}
			}
		} else {
			mergedMap[key] = item
		}
	}

	for key, value := range mergedMap {
		value.Name = key
		records = append(records, value)
	}
	if len(req.Info) != 0 {
		length, count := len(records), 0
		for count < length {
			if !strings.Contains(records[count].Name, req.Info) {
				records = append(records[:count], records[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	total, start, end := len(records), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		BackDatas = make([]dto.ComposeInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		BackDatas = records[start:end]
	}
	return int64(total), BackDatas, nil
}

func (u *ContainerService) TestCompose(req dto.ComposeCreate) (bool, error) {
	if cmd.CheckIllegal(req.Path) {
		return false, buserr.New(constant.ErrCmdIllegal)
	}
	composeItem, _ := composeRepo.GetRecord(commonRepo.WithByName(req.Name))
	if composeItem.ID != 0 {
		return false, constant.ErrRecordExist
	}
	if err := u.loadPath(&req); err != nil {
		return false, err
	}
	cmd := exec.Command("docker-compose", "-f", req.Path, "config")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.New(string(stdout))
	}
	return true, nil
}

func formatYAML(data []byte) []byte {
	return []byte(strings.ReplaceAll(string(data), "\t", "  "))
}

func updateDockerComposeWithEnv(req dto.ComposeCreate) error {
	data, err := ioutil.ReadFile(req.Path)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose.yml: %v", err)
	}
	var composeItem DockerCompose
	if err := yaml.Unmarshal(data, &composeItem); err != nil {
		return fmt.Errorf("failed to parse docker-compose.yml: %v", err)
	}
	for serviceName, service := range composeItem.Services {
		envMap := make(map[string]string)
		if existingEnv, exists := service["environment"].([]interface{}); exists {
			for _, env := range existingEnv {
				envStr := env.(string)
				parts := strings.SplitN(envStr, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}
		}
		for _, env := range req.Env {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
		envVars := []string{}
		for key, value := range envMap {
			envVars = append(envVars, key+"="+value)
		}
		service["environment"] = envVars
		composeItem.Services[serviceName] = service
	}
	if composeItem.Networks != nil {
		for key := range composeItem.Networks {
			composeItem.Networks[key] = map[string]interface{}{}
		}
	}
	newData, err := yaml.Marshal(&composeItem)
	if err != nil {
		return fmt.Errorf("failed to marshal docker-compose.yml: %v", err)
	}
	formattedData := formatYAML(newData)
	if err := ioutil.WriteFile(req.Path, formattedData, 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %v", err)
	}
	return nil
}

func (u *ContainerService) CreateCompose(req dto.ComposeCreate) (string, error) {
	if cmd.CheckIllegal(req.Name, req.Path) {
		return "", buserr.New(constant.ErrCmdIllegal)
	}
	if err := u.loadPath(&req); err != nil {
		return "", err
	}
	global.LOG.Infof("docker-compose.yml %s create successful, start to docker-compose up", req.Name)

	if req.From == "path" {
		req.Name = path.Base(path.Dir(req.Path))
	}

	dockerLogDir := path.Join(global.CONF.System.TmpDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	logItem := fmt.Sprintf("%s/compose_create_%s_%s.log", dockerLogDir, req.Name, time.Now().Format(constant.DateTimeSlimLayout))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	if len(req.Env) > 0 {
		if err := updateDockerComposeWithEnv(req); err != nil {
			fmt.Printf("failed to update docker-compose.yml with env: %v\n", err)
			return "", err
		}
	}
	go func() {
		defer file.Close()
		cmd := exec.Command("docker-compose", "-f", req.Path, "up", "-d")
		multiWriter := io.MultiWriter(os.Stdout, file)
		cmd.Stdout = multiWriter
		cmd.Stderr = multiWriter
		if err := cmd.Run(); err != nil {
			global.LOG.Errorf("docker-compose up %s failed, err: %v", req.Name, err)
			_, _ = compose.Down(req.Path)
			_, _ = file.WriteString("docker-compose up failed!")
			return
		}
		global.LOG.Infof("docker-compose up %s successful!", req.Name)
		_ = composeRepo.CreateRecord(&model.Compose{Name: req.Name, Path: req.Path})
		_, _ = file.WriteString("docker-compose up successful!")
	}()

	return path.Base(logItem), nil
}

func (u *ContainerService) ComposeOperation(req dto.ComposeOperation) error {
	if cmd.CheckIllegal(req.Path, req.Operation) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	if _, err := os.Stat(req.Path); err != nil {
		return fmt.Errorf("load file with path %s failed, %v", req.Path, err)
	}
	if req.Operation == "up" {
		if stdout, err := compose.Up(req.Path); err != nil {
			return errors.New(string(stdout))
		}
	} else {
		if stdout, err := compose.Operate(req.Path, req.Operation); err != nil {
			return errors.New(string(stdout))
		}
	}
	global.LOG.Infof("docker-compose %s %s successful", req.Operation, req.Name)
	if req.Operation == "down" {
		if req.WithFile {
			_ = composeRepo.DeleteRecord(commonRepo.WithByName(req.Name))
			_ = os.RemoveAll(path.Dir(req.Path))
		} else {
			composeItem, _ := composeRepo.GetRecord(commonRepo.WithByName(req.Name))
			if composeItem.Path == "" {
				upMap := make(map[string]interface{})
				upMap["path"] = req.Path
				_ = composeRepo.UpdateRecord(req.Name, upMap)
			}
		}
	}

	return nil
}

func updateComposeWithEnv(req dto.ComposeUpdate) error {
	var composeItem DockerCompose
	if err := yaml.Unmarshal([]byte(req.Content), &composeItem); err != nil {
		return fmt.Errorf("failed to parse docker-compose content: %v", err)
	}
	for serviceName, service := range composeItem.Services {
		envMap := make(map[string]string)
		for _, env := range req.Env {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
		newEnvVars := []string{}
		if existingEnv, exists := service["environment"].([]interface{}); exists {
			for _, env := range existingEnv {
				envStr := env.(string)
				parts := strings.SplitN(envStr, "=", 2)
				if len(parts) == 2 {
					key := parts[0]
					if value, found := envMap[key]; found {
						newEnvVars = append(newEnvVars, key+"="+value)
						delete(envMap, key)
					} else {
						newEnvVars = append(newEnvVars, envStr)
					}
				}
			}
		}
		for key, value := range envMap {
			newEnvVars = append(newEnvVars, key+"="+value)
		}
		if len(newEnvVars) > 0 {
			service["environment"] = newEnvVars
		} else {
			delete(service, "environment")
		}
		composeItem.Services[serviceName] = service
	}
	if composeItem.Networks != nil {
		for key := range composeItem.Networks {
			composeItem.Networks[key] = map[string]interface{}{}
		}
	}
	newData, err := yaml.Marshal(&composeItem)
	if err != nil {
		return fmt.Errorf("failed to marshal docker-compose.yml: %v", err)
	}
	if err := ioutil.WriteFile(req.Path, newData, 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml to path: %v", err)
	}
	return nil
}

func (u *ContainerService) ComposeUpdate(req dto.ComposeUpdate) error {
	if cmd.CheckIllegal(req.Name, req.Path) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	oldFile, err := os.ReadFile(req.Path)
	if err != nil {
		return fmt.Errorf("load file with path %s failed, %v", req.Path, err)
	}
	if len(req.Env) > 0 {
		if err := updateComposeWithEnv(req); err != nil {
			return fmt.Errorf("failed to update docker-compose with env: %v", err)
		}
	} else {
		file, err := os.OpenFile(req.Path, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			return err
		}
		defer file.Close()
		write := bufio.NewWriter(file)
		_, _ = write.WriteString(req.Content)
		write.Flush()
		global.LOG.Infof("docker-compose.yml %s has been replaced", req.Path)
	}
	if stdout, err := compose.Up(req.Path); err != nil {
		if err := recreateCompose(string(oldFile), req.Path); err != nil {
			return fmt.Errorf("update failed when handle compose up, err: %s, recreate failed: %v", string(stdout), err)
		}
		return fmt.Errorf("update failed when handle compose up, err: %s", string(stdout))
	}

	return nil
}

func (u *ContainerService) loadPath(req *dto.ComposeCreate) error {
	if req.From == "template" || req.From == "edit" {
		dir := fmt.Sprintf("%s/docker/compose/%s", constant.DataDir, req.Name)
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}

		path := fmt.Sprintf("%s/docker-compose.yml", dir)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
		write := bufio.NewWriter(file)
		_, _ = write.WriteString(string(req.File))
		write.Flush()
		req.Path = path
	}
	return nil
}

func recreateCompose(content, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(content)
	write.Flush()

	if stdout, err := compose.Up(path); err != nil {
		return errors.New(string(stdout))
	}
	return nil
}
