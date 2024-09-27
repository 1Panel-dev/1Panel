package service

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
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
	defer client.Close()

	options := container.ListOptions{All: true}
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
	for _, item := range composeCreatedByLocal {
		if err := composeRepo.DeleteRecord(commonRepo.WithByID(item.ID)); err != nil {
			global.LOG.Error(err)
		}
	}
	for key, value := range composeMap {
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
	cmd := exec.Command("docker", "compose", "-f", req.Path, "config")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.New(string(stdout))
	}
	return true, nil
}

func (u *ContainerService) CreateCompose(req dto.ComposeCreate) error {
	if cmd.CheckIllegal(req.Name, req.Path) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	if err := u.loadPath(&req); err != nil {
		return err
	}
	if req.From == "path" {
		req.Name = path.Base(path.Dir(req.Path))
	}
	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskCreate, task.TaskScopeCompose, req.TaskID, 1)
	if err != nil {
		return fmt.Errorf("new task for image build failed, err: %v", err)
	}
	go func() {
		taskItem.AddSubTask(i18n.GetMsgByKey("ComposeCreate"), func(t *task.Task) error {
			cmd := exec.Command("docker-compose", "-f", req.Path, "up", "-d")
			out, err := cmd.CombinedOutput()
			taskItem.Log(i18n.GetWithName("ComposeCreateRes", string(out)))
			if err != nil {
				_, _ = compose.Down(req.Path)
				return err
			}
			_ = composeRepo.CreateRecord(&model.Compose{Name: req.Name})
			return nil
		}, nil)
		_ = taskItem.Execute()
	}()

	return nil
}

func (u *ContainerService) ComposeOperation(req dto.ComposeOperation) error {
	if cmd.CheckIllegal(req.Path, req.Operation) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	if _, err := os.Stat(req.Path); err != nil {
		return fmt.Errorf("load file with path %s failed, %v", req.Path, err)
	}
	if stdout, err := compose.Operate(req.Path, req.Operation); err != nil {
		return errors.New(string(stdout))
	}
	global.LOG.Infof("docker-compose %s %s successful", req.Operation, req.Name)
	if req.Operation == "down" {
		_ = composeRepo.DeleteRecord(commonRepo.WithByName(req.Name))
		if req.WithFile {
			_ = os.RemoveAll(path.Dir(req.Path))
		}
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
	file, err := os.OpenFile(req.Path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.Content)
	write.Flush()

	global.LOG.Infof("docker-compose.yml %s has been replaced, now start to docker-compose restart", req.Path)
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
