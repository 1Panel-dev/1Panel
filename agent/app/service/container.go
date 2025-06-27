package service

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/gin-gonic/gin"

	"github.com/pkg/errors"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type ContainerService struct{}

type IContainerService interface {
	Page(req dto.PageContainer) (int64, interface{}, error)
	List() []string
	LoadStatus() (dto.ContainerStatus, error)
	PageNetwork(req dto.SearchWithPage) (int64, interface{}, error)
	ListNetwork() ([]dto.Options, error)
	PageVolume(req dto.SearchWithPage) (int64, interface{}, error)
	ListVolume() ([]dto.Options, error)
	PageCompose(req dto.SearchWithPage) (int64, interface{}, error)
	CreateCompose(req dto.ComposeCreate) error
	ComposeOperation(req dto.ComposeOperation) error
	ContainerCreate(req dto.ContainerOperate) error
	ContainerCreateByCommand(req dto.ContainerCreateByCommand) error
	ContainerUpdate(req dto.ContainerOperate) error
	ContainerUpgrade(req dto.ContainerUpgrade) error
	ContainerInfo(req dto.OperationWithName) (*dto.ContainerOperate, error)
	ContainerListStats() ([]dto.ContainerListStats, error)
	LoadResourceLimit() (*dto.ResourceLimit, error)
	ContainerRename(req dto.ContainerRename) error
	ContainerCommit(req dto.ContainerCommit) error
	ContainerLogClean(req dto.OperationWithName) error
	ContainerOperation(req dto.ContainerOperation) error
	DownloadContainerLogs(containerType, container, since, tail string, c *gin.Context) error
	ContainerStats(id string) (*dto.ContainerStats, error)
	Inspect(req dto.InspectReq) (string, error)
	DeleteNetwork(req dto.BatchDelete) error
	CreateNetwork(req dto.NetworkCreate) error
	DeleteVolume(req dto.BatchDelete) error
	CreateVolume(req dto.VolumeCreate) error
	TestCompose(req dto.ComposeCreate) (bool, error)
	ComposeUpdate(req dto.ComposeUpdate) error
	Prune(req dto.ContainerPrune) (dto.ContainerPruneReport, error)

	LoadUsers(req dto.OperationWithName) ([]string, error)

	StreamLogs(ctx *gin.Context, params dto.StreamLog)
}

func NewIContainerService() IContainerService {
	return &ContainerService{}
}

func (u *ContainerService) Page(req dto.PageContainer) (int64, interface{}, error) {
	var (
		records []container.Summary
		list    []container.Summary
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	defer client.Close()
	options := container.ListOptions{
		All: true,
	}
	if len(req.Filters) != 0 {
		options.Filters = filters.NewArgs()
		options.Filters.Add("label", req.Filters)
	}
	containers, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return 0, nil, err
	}
	if req.ExcludeAppStore {
		for _, item := range containers {
			if created, ok := item.Labels[composeCreatedBy]; ok && created == "Apps" {
				continue
			}
			list = append(list, item)
		}
	} else {
		list = containers
	}

	if len(req.Name) != 0 {
		length, count := len(list), 0
		for count < length {
			if !strings.Contains(list[count].Names[0][1:], req.Name) && !strings.Contains(list[count].Image, req.Name) {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	if req.State != "all" {
		length, count := len(list), 0
		for count < length {
			if list[count].State != req.State {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	switch req.OrderBy {
	case "name":
		sort.Slice(list, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return list[i].Names[0][1:] < list[j].Names[0][1:]
			}
			return list[i].Names[0][1:] > list[j].Names[0][1:]
		})
	default:
		sort.Slice(list, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return list[i].Created < list[j].Created
			}
			return list[i].Created > list[j].Created
		})
	}

	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]container.Summary, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list[start:end]
	}

	backDatas := make([]dto.ContainerInfo, len(records))
	for i := 0; i < len(records); i++ {
		item := records[i]
		IsFromCompose := false
		if _, ok := item.Labels[composeProjectLabel]; ok {
			IsFromCompose = true
		}
		IsFromApp := false
		if created, ok := item.Labels[composeCreatedBy]; ok && created == "Apps" {
			IsFromApp = true
		}

		exposePorts := transPortToStr(records[i].Ports)
		info := dto.ContainerInfo{
			ContainerID:   item.ID,
			CreateTime:    time.Unix(item.Created, 0).Format(constant.DateTimeLayout),
			Name:          item.Names[0][1:],
			ImageId:       strings.Split(item.ImageID, ":")[1],
			ImageName:     item.Image,
			State:         item.State,
			RunTime:       item.Status,
			Ports:         exposePorts,
			IsFromApp:     IsFromApp,
			IsFromCompose: IsFromCompose,
		}
		install, _ := appInstallRepo.GetFirst(appInstallRepo.WithContainerName(info.Name))
		if install.ID > 0 {
			info.AppInstallName = install.Name
			info.AppName = install.App.Name
			websites, _ := websiteRepo.GetBy(websiteRepo.WithAppInstallId(install.ID))
			for _, website := range websites {
				info.Websites = append(info.Websites, website.PrimaryDomain)
			}
		}
		backDatas[i] = info
		if item.NetworkSettings != nil && len(item.NetworkSettings.Networks) > 0 {
			networks := make([]string, 0, len(item.NetworkSettings.Networks))
			for key := range item.NetworkSettings.Networks {
				networks = append(networks, item.NetworkSettings.Networks[key].IPAddress)
			}
			sort.Strings(networks)
			backDatas[i].Network = networks
		}
	}

	return int64(total), backDatas, nil
}

func (u *ContainerService) List() []string {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("load docker client for contianer list failed, err: %v", err)
		return nil
	}
	defer client.Close()
	containers, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		global.LOG.Errorf("load contianer list failed, err: %v", err)
		return nil
	}
	var datas []string
	for _, container := range containers {
		for _, name := range container.Names {
			if len(name) != 0 {
				datas = append(datas, strings.TrimPrefix(name, "/"))
			}
		}
	}

	return datas
}

func (u *ContainerService) LoadStatus() (dto.ContainerStatus, error) {
	var data dto.ContainerStatus
	client, err := docker.NewDockerClient()
	if err != nil {
		return data, err
	}
	defer client.Close()
	c := context.Background()
	images, _ := client.ImageList(c, image.ListOptions{})
	for _, image := range images {
		data.ImageSize += uint64(image.Size)
	}
	data.ImageCount = len(images)
	repo, _ := imageRepoRepo.List()
	data.RepoCount = len(repo)
	templates, _ := composeRepo.List()
	data.ComposeTemplateCount = len(templates)
	networks, _ := client.NetworkList(c, network.ListOptions{})
	data.NetworkCount = len(networks)
	volumes, _ := client.VolumeList(c, volume.ListOptions{})
	data.VolumeCount = len(volumes.Volumes)
	data.ComposeCount = loadComposeCount(client)
	containers, err := client.ContainerList(c, container.ListOptions{All: true})
	if err != nil {
		return data, err
	}
	data.All = uint(len(containers))
	for _, item := range containers {
		switch item.State {
		case "created":
			data.Created++
		case "running":
			data.Running++
		case "paused":
			data.Paused++
		case "restarting":
			data.Restarting++
		case "dead":
			data.Dead++
		case "exited":
			data.Exited++
		case "removing":
			data.Removing++
		}
	}
	data.ContainerCount = int(data.All)
	return data, nil
}
func (u *ContainerService) ContainerListStats() ([]dto.ContainerListStats, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	list, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	var datas []dto.ContainerListStats
	var wg sync.WaitGroup
	wg.Add(len(list))
	for i := 0; i < len(list); i++ {
		go func(item container.Summary) {
			datas = append(datas, loadCpuAndMem(client, item.ID))
			wg.Done()
		}(list[i])
	}
	wg.Wait()
	return datas, nil
}

func (u *ContainerService) ContainerCreateByCommand(req dto.ContainerCreateByCommand) error {
	if cmd.CheckIllegal(req.Command) {
		return buserr.New("ErrCmdIllegal")
	}
	if !strings.HasPrefix(strings.TrimSpace(req.Command), "docker run ") {
		return errors.New("error command format")
	}
	containerName := ""
	commands := strings.Split(req.Command, " ")
	for index, val := range commands {
		if val == "--name" && len(commands) > index+1 {
			containerName = commands[index+1]
		}
	}
	if !strings.Contains(req.Command, " -d ") {
		req.Command = strings.ReplaceAll(req.Command, "docker run", "docker run -d")
	}
	if len(containerName) == 0 {
		containerName = fmt.Sprintf("1Panel-%s-%s", common.RandStr(5), common.RandStrAndNum(4))
		req.Command += fmt.Sprintf(" --name %s", containerName)
	}
	taskItem, err := task.NewTaskWithOps(containerName, task.TaskCreate, task.TaskScopeContainer, req.TaskID, 1)
	if err != nil {
		global.LOG.Errorf("new task for create container failed, err: %v", err)
		return err
	}
	go func() {
		taskItem.AddSubTask(i18n.GetWithName("ContainerCreate", containerName), func(t *task.Task) error {
			cmdMgr := cmd.NewCommandMgr(cmd.WithTask(*taskItem), cmd.WithTimeout(5*time.Minute))
			return cmdMgr.RunBashC(req.Command)
		}, nil)
		_ = taskItem.Execute()
	}()
	return nil
}

func (u *ContainerService) Inspect(req dto.InspectReq) (string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	var inspectInfo interface{}
	switch req.Type {
	case "container":
		inspectInfo, err = client.ContainerInspect(context.Background(), req.ID)
	case "compose":
		filePath := ""
		cli, err := docker.NewDockerClient()
		if err != nil {
			return "", err
		}
		defer cli.Close()
		options := container.ListOptions{All: true}
		options.Filters = filters.NewArgs()
		options.Filters.Add("label", fmt.Sprintf("%s=%s", composeProjectLabel, req.ID))
		containers, err := cli.ContainerList(context.Background(), options)
		if err != nil {
			return "", err
		}
		for _, container := range containers {
			config := container.Labels[composeConfigLabel]
			workdir := container.Labels[composeWorkdirLabel]
			if len(config) != 0 && len(workdir) != 0 && strings.Contains(config, workdir) {
				filePath = config
				break
			} else {
				filePath = workdir
				break
			}
		}
		if len(containers) == 0 {
			composeItem, _ := composeRepo.GetRecord(repo.WithByName(req.ID))
			filePath = composeItem.Path
		}
		if _, err := os.Stat(filePath); err != nil {
			return "", err
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	case "image":
		inspectInfo, _, err = client.ImageInspectWithRaw(context.Background(), req.ID)
	case "network":
		inspectInfo, err = client.NetworkInspect(context.TODO(), req.ID, network.InspectOptions{})
	case "volume":
		inspectInfo, err = client.VolumeInspect(context.TODO(), req.ID)
	}
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(inspectInfo)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (u *ContainerService) Prune(req dto.ContainerPrune) (dto.ContainerPruneReport, error) {
	report := dto.ContainerPruneReport{}
	client, err := docker.NewDockerClient()
	if err != nil {
		return report, err
	}
	defer client.Close()
	pruneFilters := filters.NewArgs()
	if req.WithTagAll {
		pruneFilters.Add("dangling", "false")
		if req.PruneType != "image" {
			pruneFilters.Add("until", "24h")
		}
	}
	switch req.PruneType {
	case "container":
		rep, err := client.ContainersPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.ContainersDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "image":
		rep, err := client.ImagesPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.ImagesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "network":
		rep, err := client.NetworksPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.NetworksDeleted)
	case "volume":
		versions, err := client.ServerVersion(context.Background())
		if err != nil {
			return report, err
		}
		if common.ComparePanelVersion(versions.APIVersion, "1.42") {
			pruneFilters.Add("all", "true")
		}
		rep, err := client.VolumesPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.VolumesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "buildcache":
		opts := build.CachePruneOptions{}
		opts.All = true
		rep, err := client.BuildCachePrune(context.Background(), opts)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.CachesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	}
	return report, nil
}

func (u *ContainerService) LoadResourceLimit() (*dto.ResourceLimit, error) {
	cpuCounts, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("load cpu limit failed, err: %v", err)
	}
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("load memory limit failed, err: %v", err)
	}

	data := dto.ResourceLimit{
		CPU:    cpuCounts,
		Memory: memoryInfo.Total,
	}
	return &data, nil
}

func (u *ContainerService) ContainerCreate(req dto.ContainerOperate) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	newContainer, _ := client.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil {
		return buserr.New("ErrContainerName")
	}

	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskCreate, task.TaskScopeContainer, req.TaskID, 1)
	if err != nil {
		global.LOG.Errorf("new task for create container failed, err: %v", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		taskItem.AddSubTask(i18n.GetWithName("ContainerImagePull", req.Image), func(t *task.Task) error {
			if !checkImageExist(client, req.Image) || req.ForcePull {
				if err := pullImages(taskItem, client, req.Image); err != nil {
					if !req.ForcePull {
						return err
					}
				}
			}
			return nil
		}, nil)

		taskItem.AddSubTask(i18n.GetMsgByKey("ContainerImageCheck"), func(t *task.Task) error {
			imageInfo, _, err := client.ImageInspectWithRaw(ctx, req.Image)
			if err != nil {
				return err
			}
			if len(req.Entrypoint) == 0 {
				req.Entrypoint = imageInfo.Config.Entrypoint
			}
			if len(req.Cmd) == 0 {
				req.Cmd = imageInfo.Config.Cmd
			}
			return nil
		}, nil)

		taskItem.AddSubTask(i18n.GetWithName("ContainerCreate", req.Name), func(t *task.Task) error {
			config, hostConf, networkConf, err := loadConfigInfo(true, req, nil)
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerLoadInfo"), err)
			if err != nil {
				return err
			}
			con, err := client.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
			if err != nil {
				taskItem.Log(i18n.GetMsgByKey("ContainerCreateFailed"))
				_ = client.ContainerRemove(ctx, req.Name, container.RemoveOptions{RemoveVolumes: true, Force: true})
				return err
			}
			err = client.ContainerStart(ctx, con.ID, container.StartOptions{})
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerStartCheck"), err)
			if err != nil {
				taskItem.Log(i18n.GetMsgByKey("ContainerCreateFailed"))
				_ = client.ContainerRemove(ctx, req.Name, container.RemoveOptions{RemoveVolumes: true, Force: true})
				return fmt.Errorf("create successful but start failed, err: %v", err)
			}
			return nil
		}, nil)

		if err := taskItem.Execute(); err != nil {
			global.LOG.Error(err.Error())
		}
	}()
	wg.Wait()
	return nil
}

func (u *ContainerService) ContainerInfo(req dto.OperationWithName) (*dto.ContainerOperate, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	ctx := context.Background()
	oldContainer, err := client.ContainerInspect(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	var data dto.ContainerOperate
	data.ContainerID = oldContainer.ID
	data.Name = strings.ReplaceAll(oldContainer.Name, "/", "")
	data.Image = oldContainer.Config.Image
	if oldContainer.NetworkSettings != nil {
		for network := range oldContainer.NetworkSettings.Networks {
			data.Network = network
			break
		}
	}

	exposePorts, _ := loadPortByInspect(oldContainer.ID, client)
	data.ExposedPorts = loadContainerPortForInfo(exposePorts)
	networkSettings := oldContainer.NetworkSettings
	bridgeNetworkSettings := networkSettings.Networks[data.Network]
	if bridgeNetworkSettings.IPAMConfig != nil {
		data.MacAddr = bridgeNetworkSettings.MacAddress
		if data.Network != "bridge" {
			data.Ipv4 = bridgeNetworkSettings.IPAMConfig.IPv4Address
			data.Ipv6 = bridgeNetworkSettings.IPAMConfig.IPv6Address
		}
	} else {
		if data.Network != "bridge" {
			data.Ipv4 = bridgeNetworkSettings.IPAddress
		}
	}
	data.Hostname = oldContainer.Config.Hostname
	data.DNS = oldContainer.HostConfig.DNS
	data.DomainName = oldContainer.Config.Domainname

	data.Cmd = oldContainer.Config.Cmd
	data.WorkingDir = oldContainer.Config.WorkingDir
	data.User = oldContainer.Config.User
	data.OpenStdin = oldContainer.Config.OpenStdin
	data.Tty = oldContainer.Config.Tty
	data.Entrypoint = oldContainer.Config.Entrypoint
	data.Env = oldContainer.Config.Env
	data.CPUShares = oldContainer.HostConfig.CPUShares
	for key, val := range oldContainer.Config.Labels {
		data.Labels = append(data.Labels, fmt.Sprintf("%s=%s", key, val))
	}
	data.AutoRemove = oldContainer.HostConfig.AutoRemove
	data.Privileged = oldContainer.HostConfig.Privileged
	data.PublishAllPorts = oldContainer.HostConfig.PublishAllPorts
	data.RestartPolicy = string(oldContainer.HostConfig.RestartPolicy.Name)
	if oldContainer.HostConfig.NanoCPUs != 0 {
		data.NanoCPUs = float64(oldContainer.HostConfig.NanoCPUs) / 1000000000
	}
	if oldContainer.HostConfig.Memory != 0 {
		data.Memory = float64(oldContainer.HostConfig.Memory) / 1024 / 1024
	}
	data.Volumes = loadVolumeBinds(oldContainer.Mounts)

	return &data, nil
}

func (u *ContainerService) ContainerUpdate(req dto.ContainerOperate) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	newContainer, _ := client.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil && newContainer.ID != req.ContainerID {
		return buserr.New("ErrContainerName")
	}
	oldContainer, err := client.ContainerInspect(ctx, req.ContainerID)
	if err != nil {
		return err
	}

	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskUpdate, task.TaskScopeContainer, req.TaskID, 1)
	if err != nil {
		global.LOG.Errorf("new task for create container failed, err: %v", err)
		return err
	}
	go func() {
		taskItem.AddSubTask(i18n.GetWithName("ContainerImagePull", req.Image), func(t *task.Task) error {
			if !checkImageExist(client, req.Image) || req.ForcePull {
				if err := pullImages(taskItem, client, req.Image); err != nil {
					if !req.ForcePull {
						return err
					}
					return fmt.Errorf("pull image %s failed, err: %v", req.Image, err)
				}
			}
			return nil
		}, nil)

		taskItem.AddSubTask(i18n.GetWithName("ContainerCreate", req.Name), func(t *task.Task) error {
			err := client.ContainerRemove(ctx, req.ContainerID, container.RemoveOptions{Force: true})
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerRemoveOld"), err)
			if err != nil {
				return err
			}

			config, hostConf, networkConf, err := loadConfigInfo(false, req, &oldContainer)
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerLoadInfo"), err)
			if err != nil {
				taskItem.Log(i18n.GetMsgByKey("ContainerRecreate"))
				reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
				return err
			}

			con, err := client.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
			if err != nil {
				taskItem.Log(i18n.GetMsgByKey("ContainerRecreate"))
				reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
				return fmt.Errorf("update container failed, err: %v", err)
			}
			err = client.ContainerStart(ctx, con.ID, container.StartOptions{})
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerStartCheck"), err)
			if err != nil {
				return fmt.Errorf("update successful but start failed, err: %v", err)
			}
			return nil
		}, nil)

		if err := taskItem.Execute(); err != nil {
			global.LOG.Error(err.Error())
		}
	}()

	return nil
}

func (u *ContainerService) ContainerUpgrade(req dto.ContainerUpgrade) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	oldContainer, err := client.ContainerInspect(ctx, req.Name)
	if err != nil {
		return err
	}
	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskUpgrade, task.TaskScopeContainer, req.TaskID, 1)
	if err != nil {
		global.LOG.Errorf("new task for create container failed, err: %v", err)
		return err
	}
	go func() {
		taskItem.AddSubTask(i18n.GetWithName("ContainerImagePull", req.Image), func(t *task.Task) error {
			if !checkImageExist(client, req.Image) || req.ForcePull {
				if err := pullImages(taskItem, client, req.Image); err != nil {
					if !req.ForcePull {
						return err
					}
					return fmt.Errorf("pull image %s failed, err: %v", req.Image, err)
				}
			}
			return nil
		}, nil)

		taskItem.AddSubTask(i18n.GetWithName("ContainerCreate", req.Name), func(t *task.Task) error {
			config := oldContainer.Config
			config.Image = req.Image
			hostConf := oldContainer.HostConfig
			var networkConf network.NetworkingConfig
			if oldContainer.NetworkSettings != nil {
				for networkKey := range oldContainer.NetworkSettings.Networks {
					networkConf.EndpointsConfig = map[string]*network.EndpointSettings{networkKey: {}}
					break
				}
			}
			err := client.ContainerRemove(ctx, req.Name, container.RemoveOptions{Force: true})
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerRemoveOld"), err)
			if err != nil {
				return err
			}

			con, err := client.ContainerCreate(ctx, config, hostConf, &networkConf, &v1.Platform{}, req.Name)
			if err != nil {
				taskItem.Log(i18n.GetMsgByKey("ContainerRecreate"))
				reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
				return fmt.Errorf("upgrade container failed, err: %v", err)
			}
			err = client.ContainerStart(ctx, con.ID, container.StartOptions{})
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerStartCheck"), err)
			if err != nil {
				return fmt.Errorf("upgrade successful but start failed, err: %v", err)
			}
			return nil
		}, nil)

		if err := taskItem.Execute(); err != nil {
			global.LOG.Error(err.Error())
		}
	}()

	return nil
}

func (u *ContainerService) ContainerRename(req dto.ContainerRename) error {
	ctx := context.Background()
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

	newContainer, _ := client.ContainerInspect(ctx, req.NewName)
	if newContainer.ContainerJSONBase != nil {
		return buserr.New("ErrContainerName")
	}
	return client.ContainerRename(ctx, req.Name, req.NewName)
}

func (u *ContainerService) ContainerCommit(req dto.ContainerCommit) error {
	ctx := context.Background()
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	options := container.CommitOptions{
		Reference: req.NewImageName,
		Comment:   req.Comment,
		Author:    req.Author,
		Changes:   nil,
		Pause:     req.Pause,
		Config:    nil,
	}
	_, err = client.ContainerCommit(ctx, req.ContainerId, options)
	if err != nil {
		return fmt.Errorf("failed to commit container, err: %v", err)
	}
	return nil
}

func (u *ContainerService) ContainerOperation(req dto.ContainerOperation) error {
	var err error
	ctx := context.Background()
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	for _, item := range req.Names {
		global.LOG.Infof("start container %s operation %s", item, req.Operation)
		switch req.Operation {
		case constant.ContainerOpStart:
			err = client.ContainerStart(ctx, item, container.StartOptions{})
		case constant.ContainerOpStop:
			err = client.ContainerStop(ctx, item, container.StopOptions{})
		case constant.ContainerOpRestart:
			err = client.ContainerRestart(ctx, item, container.StopOptions{})
		case constant.ContainerOpKill:
			err = client.ContainerKill(ctx, item, "SIGKILL")
		case constant.ContainerOpPause:
			err = client.ContainerPause(ctx, item)
		case constant.ContainerOpUnpause:
			err = client.ContainerUnpause(ctx, item)
		case constant.ContainerOpRemove:
			err = client.ContainerRemove(ctx, item, container.RemoveOptions{RemoveVolumes: true, Force: true})
		}
	}
	return err
}

func (u *ContainerService) ContainerLogClean(req dto.OperationWithName) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	containerItem, err := client.ContainerInspect(ctx, req.Name)
	if err != nil {
		return err
	}
	if err := client.ContainerStop(ctx, containerItem.ID, container.StopOptions{}); err != nil {
		return err
	}
	file, err := os.OpenFile(containerItem.LogPath, os.O_RDWR|os.O_CREATE, constant.FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = file.Truncate(0); err != nil {
		return err
	}
	_, _ = file.Seek(0, 0)

	files, _ := filepath.Glob(fmt.Sprintf("%s.*", containerItem.LogPath))
	for _, file := range files {
		_ = os.Remove(file)
	}

	if err := client.ContainerStart(ctx, containerItem.ID, container.StartOptions{}); err != nil {
		return err
	}
	return nil
}

func (u *ContainerService) StreamLogs(ctx *gin.Context, params dto.StreamLog) {
	messageChan := make(chan string, 1024)
	errorChan := make(chan error, 1)
	doneChan := make(chan struct{})

	go func() {
		<-ctx.Request.Context().Done()
		close(doneChan)
	}()
	go collectLogs(doneChan, params, messageChan, errorChan)
	ctx.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				return false
			}
			_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
			if err != nil {
				return false
			}
			return true
		case err := <-errorChan:
			if err != nil {
				_, _ = fmt.Fprintf(w, "event: error\ndata: %v\n\n", err.Error())
			}
			return false
		}
	})
}

func collectLogs(done <-chan struct{}, params dto.StreamLog, messageChan chan<- string, errorChan chan<- error) {
	defer close(messageChan)
	defer close(errorChan)
	var cmdArgs []string
	if params.Type == "compose" {
		cmdArgs = []string{"compose", "-f", params.Compose}
	}
	cmdArgs = append(cmdArgs, "logs")
	if params.Follow {
		cmdArgs = append(cmdArgs, "-f")
	}
	if params.Tail != "0" {
		cmdArgs = append(cmdArgs, "--tail", params.Tail)
	}
	if params.Since != "all" {
		cmdArgs = append(cmdArgs, "--since", params.Since)
	}
	if params.Container != "" {
		cmdArgs = append(cmdArgs, params.Container)
	}

	dockerCmd := exec.Command("docker", cmdArgs...)
	dockerCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := dockerCmd.StdoutPipe()
	if err != nil {
		errorChan <- fmt.Errorf("failed to get stdout pipe: %v", err)
		return
	}

	dockerCmd.Stderr = dockerCmd.Stdout

	if err = dockerCmd.Start(); err != nil {
		errorChan <- fmt.Errorf("failed to start docker logs command: %v", err)
		return
	}

	defer func() {
		if dockerCmd.Process != nil {
			if pgid, err := syscall.Getpgid(dockerCmd.Process.Pid); err == nil {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
			_ = dockerCmd.Process.Kill()
			_ = dockerCmd.Wait()
		}
	}()

	reader := bufio.NewReader(stdout)

	processKilled := false
	go func() {
		<-done
		if !processKilled && dockerCmd.Process != nil {
			processKilled = true
			if pgid, err := syscall.Getpgid(dockerCmd.Process.Pid); err == nil {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
			_ = dockerCmd.Process.Kill()
		}
	}()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) > 0 {
					line = strings.TrimSuffix(line, "\n")
					select {
					case messageChan <- line:
					case <-done:
						return
					}
				}
				break
			}
			errorChan <- fmt.Errorf("reader error: %v", err)
			return
		}
		line = strings.TrimSuffix(line, "\n")
		select {
		case messageChan <- line:
		case <-done:
			return
		}
	}
	_ = dockerCmd.Wait()
}

func (u *ContainerService) DownloadContainerLogs(containerType, container, since, tail string, c *gin.Context) error {
	if cmd.CheckIllegal(container, since, tail) {
		return buserr.New("ErrCmdIllegal")
	}
	commandArg := []string{"logs", container}
	if containerType == "compose" {
		commandArg = []string{"compose", "-f", container, "logs"}
	}
	if tail != "0" {
		commandArg = append(commandArg, "--tail")
		commandArg = append(commandArg, tail)
	}
	if since != "all" {
		commandArg = append(commandArg, "--since")
		commandArg = append(commandArg, since)
	}

	cmd := exec.Command("docker", commandArg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		return err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		return err
	}

	tempFile, err := os.CreateTemp("", "cmd_output_*.txt")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			global.LOG.Errorf("os.Remove() failed: %v", err)
		}
	}()
	errCh := make(chan error)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if _, err := tempFile.WriteString(line + "\n"); err != nil {
				errCh <- err
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case err := <-errCh:
		if err != nil {
			global.LOG.Errorf("Error: %v", err)
		}
	case <-time.After(3 * time.Second):
		global.LOG.Errorf("Timeout reached")
	}
	info, _ := tempFile.Stat()

	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(info.Name()))
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), tempFile)
	return nil
}

func (u *ContainerService) ContainerStats(id string) (*dto.ContainerStats, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	res, err := client.ContainerStats(context.TODO(), id, false)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		return nil, err
	}
	res.Body.Close()
	var stats *container.StatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}
	var data dto.ContainerStats
	data.CPUPercent = calculateCPUPercentUnix(stats)
	data.IORead, data.IOWrite = calculateBlockIO(stats.BlkioStats)
	data.Memory = float64(stats.MemoryStats.Usage) / 1024 / 1024
	if cache, ok := stats.MemoryStats.Stats["cache"]; ok {
		data.Cache = float64(cache) / 1024 / 1024
	}
	data.NetworkRX, data.NetworkTX = calculateNetwork(stats.Networks)
	data.ShotTime = stats.Read
	return &data, nil
}

func (u *ContainerService) LoadUsers(req dto.OperationWithName) ([]string, error) {
	std, err := cmd.RunDefaultWithStdoutBashCf("docker exec %s cat /etc/passwd", req.Name)
	if err != nil {
		return nil, err
	}
	var users []string
	lines := strings.Split(string(std), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			users = append(users, strings.Split(line, ":")[0])
		}
	}
	return users, nil
}

func stringsToMap(list []string) map[string]string {
	var labelMap = make(map[string]string)
	for _, label := range list {
		if strings.Contains(label, "=") {
			sps := strings.SplitN(label, "=", 2)
			labelMap[sps[0]] = sps[1]
		}
	}
	return labelMap
}

func calculateCPUPercentUnix(stats *container.StatsResponse) float64 {
	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * 100.0
		if len(stats.CPUStats.CPUUsage.PercpuUsage) != 0 {
			cpuPercent = cpuPercent * float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
	}
	return cpuPercent
}
func calculateMemPercentUnix(memStats container.MemoryStats) float64 {
	memPercent := 0.0
	memUsage := calculateMemUsageUnixNoCache(memStats)
	memLimit := float64(memStats.Limit)
	if memUsage > 0.0 && memLimit > 0.0 {
		memPercent = (float64(memUsage) / float64(memLimit)) * 100.0
	}
	return memPercent
}
func calculateMemUsageUnixNoCache(mem container.MemoryStats) uint64 {
	if v, isCgroup1 := mem.Stats["total_inactive_file"]; isCgroup1 && v < mem.Usage {
		return mem.Usage - v
	}
	if v := mem.Stats["inactive_file"]; v < mem.Usage {
		return mem.Usage - v
	}
	return mem.Usage
}
func calculateBlockIO(blkio container.BlkioStats) (blkRead float64, blkWrite float64) {
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = (blkRead + float64(bioEntry.Value)) / 1024 / 1024
		case "write":
			blkWrite = (blkWrite + float64(bioEntry.Value)) / 1024 / 1024
		}
	}
	return
}
func calculateNetwork(network map[string]container.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes) / 1024
		tx += float64(v.TxBytes) / 1024
	}
	return rx, tx
}

func checkImageExist(client *client.Client, imageItem string) bool {
	if client == nil {
		var err error
		client, err = docker.NewDockerClient()
		if err != nil {
			return false
		}
	}
	images, err := client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return false
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageItem || tag == imageItem+":latest" {
				return true
			}
		}
	}
	return false
}

func pullImages(task *task.Task, client *client.Client, imageName string) error {
	dockerCli := docker.NewClientWithExist(client)
	options := image.PullOptions{}
	repos, _ := imageRepoRepo.List()
	if len(repos) != 0 {
		for _, repo := range repos {
			if strings.HasPrefix(imageName, repo.DownloadUrl) && repo.Auth {
				authConfig := registry.AuthConfig{
					Username: repo.Username,
					Password: repo.Password,
				}
				encodedJSON, err := json.Marshal(authConfig)
				if err != nil {
					return err
				}
				authStr := base64.URLEncoding.EncodeToString(encodedJSON)
				options.RegistryAuth = authStr
			}
		}
	} else {
		hasAuth, authStr := loadAuthInfo(imageName)
		if hasAuth {
			options.RegistryAuth = authStr
		}
	}
	return dockerCli.PullImageWithProcessAndOptions(task, imageName, options)
}

func loadCpuAndMem(client *client.Client, containerItem string) dto.ContainerListStats {
	data := dto.ContainerListStats{
		ContainerID: containerItem,
	}
	res, err := client.ContainerStats(context.Background(), containerItem, false)
	if err != nil {
		return data
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return data
	}
	var stats *container.StatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return data
	}

	data.CPUTotalUsage = stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	data.SystemUsage = stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
	data.CPUPercent = calculateCPUPercentUnix(stats)
	data.PercpuUsage = len(stats.CPUStats.CPUUsage.PercpuUsage)

	data.MemoryCache = stats.MemoryStats.Stats["cache"]
	data.MemoryUsage = calculateMemUsageUnixNoCache(stats.MemoryStats)
	data.MemoryLimit = stats.MemoryStats.Limit

	data.MemoryPercent = calculateMemPercentUnix(stats.MemoryStats)
	return data
}

func checkPortStats(ports []dto.PortHelper) (nat.PortMap, error) {
	portMap := make(nat.PortMap)
	if len(ports) == 0 {
		return portMap, nil
	}
	for _, port := range ports {
		if strings.Contains(port.ContainerPort, "-") {
			if !strings.Contains(port.HostPort, "-") {
				return portMap, buserr.New("ErrPortRules")
			}
			hostStart, _ := strconv.Atoi(strings.Split(port.HostPort, "-")[0])
			hostEnd, _ := strconv.Atoi(strings.Split(port.HostPort, "-")[1])
			containerStart, _ := strconv.Atoi(strings.Split(port.ContainerPort, "-")[0])
			containerEnd, _ := strconv.Atoi(strings.Split(port.ContainerPort, "-")[1])
			if (hostEnd-hostStart) <= 0 || (containerEnd-containerStart) <= 0 {
				return portMap, buserr.New("ErrPortRules")
			}
			if (containerEnd - containerStart) != (hostEnd - hostStart) {
				return portMap, buserr.New("ErrPortRules")
			}
			for i := 0; i <= hostEnd-hostStart; i++ {
				bindItem := nat.PortBinding{HostPort: strconv.Itoa(hostStart + i), HostIP: port.HostIP}
				portMap[nat.Port(fmt.Sprintf("%d/%s", containerStart+i, port.Protocol))] = []nat.PortBinding{bindItem}
			}
			for i := hostStart; i <= hostEnd; i++ {
				if common.ScanPort(i) {
					return portMap, buserr.WithDetail("ErrPortInUsed", i, nil)
				}
			}
		} else {
			portItem := 0
			if strings.Contains(port.HostPort, "-") {
				portItem, _ = strconv.Atoi(strings.Split(port.HostPort, "-")[0])
			} else {
				portItem, _ = strconv.Atoi(port.HostPort)
			}
			if common.ScanPort(portItem) {
				return portMap, buserr.WithDetail("ErrPortInUsed", portItem, nil)
			}
			bindItem := nat.PortBinding{HostPort: strconv.Itoa(portItem), HostIP: port.HostIP}
			portMap[nat.Port(fmt.Sprintf("%s/%s", port.ContainerPort, port.Protocol))] = []nat.PortBinding{bindItem}
		}
	}
	return portMap, nil
}

func loadConfigInfo(isCreate bool, req dto.ContainerOperate, oldContainer *container.InspectResponse) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	var config container.Config
	var hostConf container.HostConfig
	if !isCreate {
		config = *oldContainer.Config
		hostConf = *oldContainer.HostConfig
	}
	var networkConf network.NetworkingConfig

	portMap, err := checkPortStats(req.ExposedPorts)
	if err != nil {
		return nil, nil, nil, err
	}
	exposed := make(nat.PortSet)
	for port := range portMap {
		exposed[port] = struct{}{}
	}
	config.Image = req.Image
	config.Cmd = req.Cmd
	config.Entrypoint = req.Entrypoint
	config.Env = req.Env
	config.Labels = stringsToMap(req.Labels)
	config.ExposedPorts = exposed
	config.OpenStdin = req.OpenStdin
	config.Tty = req.Tty
	config.Hostname = req.Hostname
	config.Domainname = req.DomainName
	config.User = req.User
	config.WorkingDir = req.WorkingDir

	if len(req.Network) != 0 {
		switch req.Network {
		case "host", "none", "bridge":
			hostConf.NetworkMode = container.NetworkMode(req.Network)
		}
		if req.Ipv4 != "" || req.Ipv6 != "" {
			networkConf.EndpointsConfig = map[string]*network.EndpointSettings{req.Network: {
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: req.Ipv4,
					IPv6Address: req.Ipv6,
				}, MacAddress: req.MacAddr}}
		} else {
			networkConf.EndpointsConfig = map[string]*network.EndpointSettings{req.Network: {}}
		}
	} else {
		if req.Ipv4 != "" || req.Ipv6 != "" {
			return nil, nil, nil, fmt.Errorf("please set up the network")
		}
		networkConf = network.NetworkingConfig{}
	}

	hostConf.Privileged = req.Privileged
	hostConf.AutoRemove = req.AutoRemove
	hostConf.CPUShares = req.CPUShares
	hostConf.PublishAllPorts = req.PublishAllPorts
	hostConf.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(req.RestartPolicy)}
	if req.RestartPolicy == "on-failure" {
		hostConf.RestartPolicy.MaximumRetryCount = 5
	}
	hostConf.NanoCPUs = int64(req.NanoCPUs * 1000000000)
	hostConf.Memory = int64(req.Memory * 1024 * 1024)
	hostConf.MemorySwap = 0
	hostConf.PortBindings = portMap
	hostConf.Binds = []string{}
	hostConf.Mounts = []mount.Mount{}
	hostConf.DNS = req.DNS
	config.Volumes = make(map[string]struct{})
	for _, volume := range req.Volumes {
		if volume.Type == "volume" {
			hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
				Type:     mount.Type(volume.Type),
				Source:   volume.SourceDir,
				Target:   volume.ContainerDir,
				ReadOnly: volume.Mode == "ro",
			})
			config.Volumes[volume.ContainerDir] = struct{}{}
		} else {
			hostConf.Binds = append(hostConf.Binds, fmt.Sprintf("%s:%s:%s", volume.SourceDir, volume.ContainerDir, volume.Mode))
		}
	}
	return &config, &hostConf, &networkConf, nil
}

func reCreateAfterUpdate(name string, client *client.Client, config *container.Config, hostConf *container.HostConfig, networkConf *container.NetworkSettings) {
	ctx := context.Background()

	var oldNetworkConf network.NetworkingConfig
	if networkConf != nil {
		for networkKey := range networkConf.Networks {
			oldNetworkConf.EndpointsConfig = map[string]*network.EndpointSettings{networkKey: {}}
			break
		}
	}

	oldContainer, err := client.ContainerCreate(ctx, config, hostConf, &oldNetworkConf, &v1.Platform{}, name)
	if err != nil {
		global.LOG.Errorf("recreate after container update failed, err: %v", err)
		return
	}
	if err := client.ContainerStart(ctx, oldContainer.ID, container.StartOptions{}); err != nil {
		global.LOG.Errorf("restart after container update failed, err: %v", err)
	}
	global.LOG.Info("recreate after container update successful")
}

func loadVolumeBinds(binds []container.MountPoint) []dto.VolumeHelper {
	var datas []dto.VolumeHelper
	for _, bind := range binds {
		var volumeItem dto.VolumeHelper
		volumeItem.Type = string(bind.Type)
		if bind.Type == "volume" {
			volumeItem.SourceDir = bind.Name
		} else {
			volumeItem.SourceDir = bind.Source
		}
		volumeItem.ContainerDir = bind.Destination
		volumeItem.Mode = "ro"
		if bind.RW {
			volumeItem.Mode = "rw"
		}
		datas = append(datas, volumeItem)
	}
	return datas
}

func loadPortByInspect(id string, client *client.Client) ([]container.Port, error) {
	containerItem, err := client.ContainerInspect(context.Background(), id)
	if err != nil {
		return nil, err
	}
	var itemPorts []container.Port
	for key, val := range containerItem.ContainerJSONBase.HostConfig.PortBindings {
		if !strings.Contains(string(key), "/") {
			continue
		}
		item := strings.Split(string(key), "/")
		itemPort, _ := strconv.ParseUint(item[0], 10, 16)

		for _, itemVal := range val {
			publicPort, _ := strconv.ParseUint(itemVal.HostPort, 10, 16)
			itemPorts = append(itemPorts, container.Port{PrivatePort: uint16(itemPort), Type: item[1], PublicPort: uint16(publicPort), IP: itemVal.HostIP})
		}
	}
	return itemPorts, nil
}
func transPortToStr(ports []container.Port) []string {
	var (
		ipv4Ports []container.Port
		ipv6Ports []container.Port
	)
	for _, port := range ports {
		if strings.Contains(port.IP, ":") {
			ipv6Ports = append(ipv6Ports, port)
		} else {
			ipv4Ports = append(ipv4Ports, port)
		}
	}
	list1 := simplifyPort(ipv4Ports)
	list2 := simplifyPort(ipv6Ports)
	return append(list1, list2...)
}
func simplifyPort(ports []container.Port) []string {
	var datas []string
	if len(ports) == 0 {
		return datas
	}
	if len(ports) == 1 {
		ip := ""
		if len(ports[0].IP) != 0 {
			ip = ports[0].IP + ":"
		}
		itemPortStr := fmt.Sprintf("%s%v/%s", ip, ports[0].PrivatePort, ports[0].Type)
		if ports[0].PublicPort != 0 {
			itemPortStr = fmt.Sprintf("%s%v->%v/%s", ip, ports[0].PublicPort, ports[0].PrivatePort, ports[0].Type)
		}
		datas = append(datas, itemPortStr)
		return datas
	}

	sort.Slice(ports, func(i, j int) bool {
		return ports[i].PrivatePort < ports[j].PrivatePort
	})
	start := ports[0]

	for i := 1; i < len(ports); i++ {
		if ports[i].PrivatePort != ports[i-1].PrivatePort+1 || ports[i].IP != ports[i-1].IP || ports[i].PublicPort != ports[i-1].PublicPort+1 || ports[i].Type != ports[i-1].Type {
			if ports[i-1].PrivatePort == start.PrivatePort {
				itemPortStr := fmt.Sprintf("%s:%v/%s", start.IP, start.PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v->%v/%s", start.IP, start.PublicPort, start.PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			} else {
				itemPortStr := fmt.Sprintf("%s:%v-%v/%s", start.IP, start.PrivatePort, ports[i-1].PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v-%v->%v-%v/%s", start.IP, start.PublicPort, ports[i-1].PublicPort, start.PrivatePort, ports[i-1].PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			}
			start = ports[i]
		}
		if i == len(ports)-1 {
			if ports[i].PrivatePort == start.PrivatePort {
				itemPortStr := fmt.Sprintf("%s:%v/%s", start.IP, start.PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v->%v/%s", start.IP, start.PublicPort, start.PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			} else {
				itemPortStr := fmt.Sprintf("%s:%v-%v/%s", start.IP, start.PrivatePort, ports[i].PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v-%v->%v-%v/%s", start.IP, start.PublicPort, ports[i].PublicPort, start.PrivatePort, ports[i].PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			}
		}
	}
	return datas
}

func loadComposeCount(client *client.Client) int {
	options := container.ListOptions{All: true}
	options.Filters = filters.NewArgs()
	options.Filters.Add("label", composeProjectLabel)
	list, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return 0
	}
	composeCreatedByLocal, _ := composeRepo.ListRecord()
	composeMap := make(map[string]struct{})
	for _, container := range list {
		if name, ok := container.Labels[composeProjectLabel]; ok {
			if _, has := composeMap[name]; !has {
				composeMap[name] = struct{}{}
			}
		}
	}
	for _, compose := range composeCreatedByLocal {
		if _, has := composeMap[compose.Name]; !has {
			composeMap[compose.Name] = struct{}{}
		}
	}

	return len(composeMap)
}
func loadContainerPortForInfo(itemPorts []container.Port) []dto.PortHelper {
	var exposedPorts []dto.PortHelper
	samePortMap := make(map[string]dto.PortHelper)
	ports := transPortToStr(itemPorts)
	for _, item := range ports {
		itemStr := strings.Split(item, "->")
		if len(itemStr) < 2 {
			continue
		}
		var itemPort dto.PortHelper
		lastIndex := strings.LastIndex(itemStr[0], ":")
		if lastIndex == -1 {
			itemPort.HostPort = itemStr[0]
		} else {
			itemPort.HostIP = itemStr[0][0:lastIndex]
			itemPort.HostPort = itemStr[0][lastIndex+1:]
		}
		itemContainer := strings.Split(itemStr[1], "/")
		if len(itemContainer) != 2 {
			continue
		}
		itemPort.ContainerPort = itemContainer[0]
		itemPort.Protocol = itemContainer[1]
		keyItem := fmt.Sprintf("%s->%s/%s", itemPort.HostPort, itemPort.ContainerPort, itemPort.Protocol)
		if val, ok := samePortMap[keyItem]; ok {
			val.HostIP = ""
			samePortMap[keyItem] = val
		} else {
			samePortMap[keyItem] = itemPort
		}
	}
	for _, val := range samePortMap {
		exposedPorts = append(exposedPorts, val)
	}
	return exposedPorts
}
