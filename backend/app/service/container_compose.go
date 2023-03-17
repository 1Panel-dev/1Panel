package service

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/net/context"
)

const composeProjectLabel = "com.docker.compose.project"
const composeConfigLabel = "com.docker.compose.project.config_files"
const composeWorkdirLabel = "com.docker.compose.project.working_dir"
const composeCreatedBy = "createdBy"

func (u *ContainerService) PageCompose(req dto.SearchWithPage) (int64, interface{}, error) {
	var (
		records   []dto.ComposeInfo
		BackDatas []dto.ComposeInfo
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}

	options := types.ContainerListOptions{All: true}
	options.Filters = filters.NewArgs()
	options.Filters.Add("label", composeProjectLabel)

	list, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return 0, nil, err
	}

	composeCreatedByLocal, _ := composeRepo.ListRecord()
	composeMap := make(map[string]dto.ComposeInfo)
	for _, container := range list {
		if name, ok := container.Labels[composeProjectLabel]; ok {
			containerItem := dto.ComposeContainer{
				ContainerID: container.ID,
				Name:        container.Names[0][1:],
				State:       container.State,
				CreateTime:  time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
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
					CreatedAt:       time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
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
	for _, item := range composeCreatedByLocal {
		if err := composeRepo.DeleteRecord(commonRepo.WithByName(item.Name)); err != nil {
			global.LOG.Error(err)
		}
	}
	for key, value := range composeMap {
		value.Name = key
		records = append(records, value)
	}
	if len(req.Info) != 0 {
		lenth, count := len(records), 0
		for count < lenth {
			if !strings.Contains(records[count].Name, req.Info) {
				records = append(records[:count], records[(count+1):]...)
				lenth--
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

func (u *ContainerService) CreateCompose(req dto.ComposeCreate) error {
	if req.From == "template" {
		template, err := composeRepo.Get(commonRepo.WithByID(req.Template))
		if err != nil {
			return err
		}
		req.From = "edit"
		req.File = template.Content
	}
	if req.From == "edit" {
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
	global.LOG.Infof("docker-compose.yml %s create successful, start to docker-compose up", req.Name)

	if req.From == "path" {
		req.Name = path.Base(strings.ReplaceAll(req.Path, "/docker-compose.yml", ""))
	}
	if stdout, err := compose.Up(req.Path); err != nil {
		_, _ = compose.Down(req.Path)
		return errors.New(stdout)
	}
	_ = composeRepo.CreateRecord(&model.Compose{Name: req.Name})

	return nil
}

func (u *ContainerService) ComposeOperation(req dto.ComposeOperation) error {
	if _, err := os.Stat(req.Path); err != nil {
		return fmt.Errorf("load file with path %s failed, %v", req.Path, err)
	}
	if stdout, err := compose.Operate(req.Path, req.Operation); err != nil {
		return errors.New(string(stdout))
	}
	global.LOG.Infof("docker-compose %s %s successful", req.Operation, req.Name)
	if req.Operation == "down" {
		_ = composeRepo.DeleteRecord(commonRepo.WithByName(req.Name))
		_ = os.RemoveAll(strings.ReplaceAll(req.Path, "/docker-compose.yml", ""))
	}

	return nil
}

func (u *ContainerService) ComposeUpdate(req dto.ComposeUpdate) error {
	if _, err := os.Stat(req.Path); err != nil {
		return fmt.Errorf("load file with path %s failed, %v", req.Path, err)
	}
	file, err := os.OpenFile(req.Path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.Content)
	write.Flush()

	global.LOG.Infof("docker-compose.yml %s has been replaced, now start to docker-compose restart", req.Path)
	if stdout, err := compose.Down(req.Path); err != nil {
		return errors.New(string(stdout))
	}
	if stdout, err := compose.Up(req.Path); err != nil {
		return errors.New(string(stdout))
	}

	return nil
}
