package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/net/context"
)

const composeProjectLabel = "com.docker.compose.project"
const composeConfigLabel = "com.docker.compose.project.config_files"
const composeWorkdirLabel = "com.docker.compose.project.working_dir"
const composeCreatedBy = "createdBy"

func (u *ContainerService) PageCompose(req dto.PageInfo) (int64, interface{}, error) {
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
				composeMap[name] = composeItem
			}
		}
	}
	for key, value := range composeMap {
		value.Name = key
		records = append(records, value)
	}
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
		req.From = template.From
		if req.From == "edit" {
			req.File = template.Content
		} else {
			req.Path = template.Path
		}
	}
	if req.From == "edit" {
		dir := fmt.Sprintf("%s/%s", constant.TmpComposeBuildDir, req.Name)
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
	go func() {
		cmd := exec.Command("docker-compose", "-f", req.Path, "up", "-d")
		stdout, err := cmd.CombinedOutput()
		if err != nil {
			global.LOG.Debugf("docker-compose up %s failed, err: %v", req.Name, err)
			return
		}
		global.LOG.Debugf("docker-compose up %s successful, logs: %v", req.Name, string(stdout))
	}()

	return nil
}

func (u *ContainerService) ComposeOperation(req dto.ComposeOperation) error {
	cmd := exec.Command("docker-compose", "-f", req.Path, req.Operation)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	global.LOG.Debugf("docker-compose %s %s successful: logs: %v", req.Operation, req.Path, string(stdout))

	return err
}
