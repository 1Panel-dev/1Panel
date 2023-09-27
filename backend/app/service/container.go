package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/websocket"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type ContainerService struct{}

type IContainerService interface {
	Page(req dto.PageContainer) (int64, interface{}, error)
	List() ([]string, error)
	PageNetwork(req dto.SearchWithPage) (int64, interface{}, error)
	ListNetwork() ([]dto.Options, error)
	PageVolume(req dto.SearchWithPage) (int64, interface{}, error)
	ListVolume() ([]dto.Options, error)
	PageCompose(req dto.SearchWithPage) (int64, interface{}, error)
	CreateCompose(req dto.ComposeCreate) (string, error)
	ComposeOperation(req dto.ComposeOperation) error
	ContainerCreate(req dto.ContainerOperate) error
	ContainerUpdate(req dto.ContainerOperate) error
	ContainerUpgrade(req dto.ContainerUpgrade) error
	ContainerInfo(req dto.OperationWithName) (*dto.ContainerOperate, error)
	ContainerListStats() ([]dto.ContainerListStats, error)
	LoadResouceLimit() (*dto.ResourceLimit, error)
	ContainerLogClean(req dto.OperationWithName) error
	ContainerOperation(req dto.ContainerOperation) error
	ContainerLogs(wsConn *websocket.Conn, container, since, tail string, follow bool) error
	ContainerStats(id string) (*dto.ContainerStats, error)
	Inspect(req dto.InspectReq) (string, error)
	DeleteNetwork(req dto.BatchDelete) error
	CreateNetwork(req dto.NetworkCreate) error
	DeleteVolume(req dto.BatchDelete) error
	CreateVolume(req dto.VolumeCreate) error
	TestCompose(req dto.ComposeCreate) (bool, error)
	ComposeUpdate(req dto.ComposeUpdate) error
	Prune(req dto.ContainerPrune) (dto.ContainerPruneReport, error)

	LoadContainerLogs(req dto.OperationWithNameAndType) string
	ComposeLogs(wsConn *websocket.Conn, composePath, since, tail string, follow bool) error
}

func NewIContainerService() IContainerService {
	return &ContainerService{}
}

func (u *ContainerService) Page(req dto.PageContainer) (int64, interface{}, error) {
	var (
		records []types.Container
		list    []types.Container
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	options := types.ContainerListOptions{All: true}
	if len(req.Filters) != 0 {
		options.Filters = filters.NewArgs()
		options.Filters.Add("label", req.Filters)
	}
	list, err = client.ContainerList(context.Background(), options)
	if err != nil {
		return 0, nil, err
	}
	if len(req.Name) != 0 {
		length, count := len(list), 0
		for count < length {
			if !strings.Contains(list[count].Names[0][1:], req.Name) {
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
	case "state":
		sort.Slice(list, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return list[i].State < list[j].State
			}
			return list[i].State > list[j].State
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
		records = make([]types.Container, 0)
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

		ports := simplifyPort(item.Ports)
		backDatas[i] = dto.ContainerInfo{
			ContainerID:   item.ID,
			CreateTime:    time.Unix(item.Created, 0).Format("2006-01-02 15:04:05"),
			Name:          item.Names[0][1:],
			ImageId:       strings.Split(item.ImageID, ":")[1],
			ImageName:     item.Image,
			State:         item.State,
			RunTime:       item.Status,
			Ports:         ports,
			IsFromApp:     IsFromApp,
			IsFromCompose: IsFromCompose,
		}
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

func (u *ContainerService) List() ([]string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	var datas []string
	for _, container := range containers {
		for _, name := range container.Names {
			if len(name) != 0 {
				datas = append(datas, strings.TrimPrefix(name, "/"))
			}
		}
	}

	return datas, nil
}

func (u *ContainerService) ContainerListStats() ([]dto.ContainerListStats, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	list, err := client.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	var datas []dto.ContainerListStats
	var wg sync.WaitGroup
	wg.Add(len(list))
	for i := 0; i < len(list); i++ {
		go func(item types.Container) {
			datas = append(datas, loadCpuAndMem(client, item.ID))
			wg.Done()
		}(list[i])
	}
	wg.Wait()
	return datas, nil
}

func (u *ContainerService) Inspect(req dto.InspectReq) (string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	var inspectInfo interface{}
	switch req.Type {
	case "container":
		inspectInfo, err = client.ContainerInspect(context.Background(), req.ID)
	case "image":
		inspectInfo, _, err = client.ImageInspectWithRaw(context.Background(), req.ID)
	case "network":
		inspectInfo, err = client.NetworkInspect(context.TODO(), req.ID, types.NetworkInspectOptions{})
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
		rep, err := client.VolumesPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.VolumesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	}
	return report, nil
}

func (u *ContainerService) LoadResouceLimit() (*dto.ResourceLimit, error) {
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
		Memory: int(memoryInfo.Total),
	}
	return &data, nil
}

func (u *ContainerService) ContainerCreate(req dto.ContainerOperate) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	newContainer, _ := client.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil {
		return buserr.New(constant.ErrContainerName)
	}

	var config container.Config
	var hostConf container.HostConfig
	var networkConf network.NetworkingConfig
	if err := loadConfigInfo(req, &config, &hostConf, &networkConf); err != nil {
		return err
	}

	global.LOG.Infof("new container info %s has been made, now start to create", req.Name)

	if !checkImageExist(client, req.Image) || req.ForcePull {
		if err := pullImages(ctx, client, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			global.LOG.Errorf("force pull image %s failed, err: %v", req.Image, err)
		}
	}
	container, err := client.ContainerCreate(ctx, &config, &hostConf, &networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		_ = client.ContainerRemove(ctx, req.Name, types.ContainerRemoveOptions{RemoveVolumes: true, Force: true})
		return err
	}
	global.LOG.Infof("create container %s successful! now check if the container is started and delete the container information if it is not.", req.Name)
	if err := client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		_ = client.ContainerRemove(ctx, req.Name, types.ContainerRemoveOptions{RemoveVolumes: true, Force: true})
		return fmt.Errorf("create successful but start failed, err: %v", err)
	}
	return nil
}

func (u *ContainerService) ContainerInfo(req dto.OperationWithName) (*dto.ContainerOperate, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
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
	data.Cmd = oldContainer.Config.Cmd
	data.Entrypoint = oldContainer.Config.Entrypoint
	data.Env = oldContainer.Config.Env
	data.CPUShares = oldContainer.HostConfig.CPUShares
	for key, val := range oldContainer.Config.Labels {
		data.Labels = append(data.Labels, fmt.Sprintf("%s=%s", key, val))
	}
	for key, val := range oldContainer.HostConfig.PortBindings {
		var itemPort dto.PortHelper
		if !strings.Contains(string(key), "/") {
			continue
		}
		itemPort.ContainerPort = strings.Split(string(key), "/")[0]
		itemPort.Protocol = strings.Split(string(key), "/")[1]
		for _, binds := range val {
			itemPort.HostIP = binds.HostIP
			itemPort.HostPort = binds.HostPort
			data.ExposedPorts = append(data.ExposedPorts, itemPort)
		}
	}
	data.AutoRemove = oldContainer.HostConfig.AutoRemove
	data.PublishAllPorts = oldContainer.HostConfig.PublishAllPorts
	data.RestartPolicy = oldContainer.HostConfig.RestartPolicy.Name
	if oldContainer.HostConfig.NanoCPUs != 0 {
		data.NanoCPUs = float64(oldContainer.HostConfig.NanoCPUs) / 1000000000
	}
	if oldContainer.HostConfig.Memory != 0 {
		data.Memory = float64(oldContainer.HostConfig.Memory) / 1024 / 1024
	}
	data.Volumes = loadVolumeBinds(oldContainer.HostConfig.Binds)

	return &data, nil
}

func (u *ContainerService) ContainerUpdate(req dto.ContainerOperate) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	newContainer, _ := client.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil && newContainer.ID != req.ContainerID {
		return buserr.New(constant.ErrContainerName)
	}

	oldContainer, err := client.ContainerInspect(ctx, req.ContainerID)
	if err != nil {
		return err
	}
	if !checkImageExist(client, req.Image) || req.ForcePull {
		if err := pullImages(ctx, client, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			global.LOG.Errorf("force pull image %s failed, err: %v", req.Image, err)
		}
	}

	if err := client.ContainerRemove(ctx, req.ContainerID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return err
	}

	config := oldContainer.Config
	hostConf := oldContainer.HostConfig
	var networkConf network.NetworkingConfig
	if err := loadConfigInfo(req, config, hostConf, &networkConf); err != nil {
		reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
		return err
	}

	global.LOG.Infof("new container info %s has been update, now start to recreate", req.Name)
	container, err := client.ContainerCreate(ctx, config, hostConf, &networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
		return fmt.Errorf("update contianer failed, err: %v", err)
	}
	global.LOG.Infof("update container %s successful! now check if the container is started.", req.Name)
	if err := client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("update successful but start failed, err: %v", err)
	}

	return nil
}

func (u *ContainerService) ContainerUpgrade(req dto.ContainerUpgrade) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	oldContainer, err := client.ContainerInspect(ctx, req.Name)
	if err != nil {
		return err
	}
	if !checkImageExist(client, req.Image) || req.ForcePull {
		if err := pullImages(ctx, client, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			global.LOG.Errorf("force pull image %s failed, err: %v", req.Image, err)
		}
	}
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
	if err := client.ContainerRemove(ctx, req.Name, types.ContainerRemoveOptions{Force: true}); err != nil {
		return err
	}

	global.LOG.Infof("new container info %s has been update, now start to recreate", req.Name)
	container, err := client.ContainerCreate(ctx, config, hostConf, &networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
		return fmt.Errorf("upgrade contianer failed, err: %v", err)
	}
	global.LOG.Infof("upgrade container %s successful! now check if the container is started.", req.Name)
	if err := client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("upgrade successful but start failed, err: %v", err)
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
	global.LOG.Infof("start container %s operation %s", req.Name, req.Operation)
	switch req.Operation {
	case constant.ContainerOpStart:
		err = client.ContainerStart(ctx, req.Name, types.ContainerStartOptions{})
	case constant.ContainerOpStop:
		err = client.ContainerStop(ctx, req.Name, container.StopOptions{})
	case constant.ContainerOpRestart:
		err = client.ContainerRestart(ctx, req.Name, container.StopOptions{})
	case constant.ContainerOpKill:
		err = client.ContainerKill(ctx, req.Name, "SIGKILL")
	case constant.ContainerOpPause:
		err = client.ContainerPause(ctx, req.Name)
	case constant.ContainerOpUnpause:
		err = client.ContainerUnpause(ctx, req.Name)
	case constant.ContainerOpRename:
		newContainer, _ := client.ContainerInspect(ctx, req.NewName)
		if newContainer.ContainerJSONBase != nil {
			return buserr.New(constant.ErrContainerName)
		}
		err = client.ContainerRename(ctx, req.Name, req.NewName)
	case constant.ContainerOpRemove:
		err = client.ContainerRemove(ctx, req.Name, types.ContainerRemoveOptions{RemoveVolumes: true, Force: true})
	}
	return err
}

func (u *ContainerService) ContainerLogClean(req dto.OperationWithName) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	containerItem, err := client.ContainerInspect(ctx, req.Name)
	if err != nil {
		return err
	}
	if err := client.ContainerStop(ctx, containerItem.ID, container.StopOptions{}); err != nil {
		return err
	}
	file, err := os.OpenFile(containerItem.LogPath, os.O_RDWR|os.O_CREATE, 0666)
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

	if err := client.ContainerStart(ctx, containerItem.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

func (u *ContainerService) ContainerLogs(wsConn *websocket.Conn, container, since, tail string, follow bool) error {
	if cmd.CheckIllegal(container, since, tail) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	command := fmt.Sprintf("docker logs %s", container)
	if tail != "0" {
		command += " --tail " + tail
	}
	if since != "all" {
		command += " --since " + since
	}
	if follow {
		command += " -f"
	}
	command += " 2>&1"
	cmd := exec.Command("bash", "-c", command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	buffer := make([]byte, 1024)
	for {
		n, err := stdout.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			global.LOG.Errorf("read bytes from container log failed, err: %v", err)
			continue
		}
		if err = wsConn.WriteMessage(websocket.TextMessage, buffer[:n]); err != nil {
			global.LOG.Errorf("send message with container log to ws failed, err: %v", err)
			break
		}
	}
	return nil
}

func (u *ContainerService) ContainerStats(id string) (*dto.ContainerStats, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
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
	var stats *types.StatsJSON
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
	data.Memory = data.Memory - data.Cache
	data.NetworkRX, data.NetworkTX = calculateNetwork(stats.Networks)
	data.ShotTime = stats.Read
	return &data, nil
}

func (u *ContainerService) LoadContainerLogs(req dto.OperationWithNameAndType) string {
	filePath := ""
	switch req.Type {
	case "image-pull", "image-push", "image-build", "compose-create":
		filePath = path.Join(global.CONF.System.TmpDir, fmt.Sprintf("docker_logs/%s", req.Name))
	case "compose-detail":
		client, err := docker.NewDockerClient()
		if err != nil {
			return ""
		}
		options := types.ContainerListOptions{All: true}
		options.Filters = filters.NewArgs()
		options.Filters.Add("label", fmt.Sprintf("%s=%s", composeProjectLabel, req.Name))
		containers, err := client.ContainerList(context.Background(), options)
		if err != nil {
			return ""
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
		if req.Type == "compose-create" {
			filePath = path.Join(path.Dir(filePath), "compose.log")
		}
	}
	if _, err := os.Stat(filePath); err != nil {
		return ""
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(content)
}

func stringsToMap(list []string) map[string]string {
	var lableMap = make(map[string]string)
	for _, label := range list {
		if strings.Contains(label, "=") {
			sps := strings.SplitN(label, "=", 2)
			lableMap[sps[0]] = sps[1]
		}
	}
	return lableMap
}

func calculateCPUPercentUnix(stats *types.StatsJSON) float64 {
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
func calculateMemPercentUnix(memStats types.MemoryStats) float64 {
	memPercent := 0.0
	memUsage := float64(memStats.Usage - memStats.Stats["cache"])
	memLimit := float64(memStats.Limit)
	if memUsage > 0.0 && memLimit > 0.0 {
		memPercent = (memUsage / memLimit) * 100.0
	}
	return memPercent
}
func calculateBlockIO(blkio types.BlkioStats) (blkRead float64, blkWrite float64) {
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
func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes) / 1024
		tx += float64(v.TxBytes) / 1024
	}
	return rx, tx
}

func checkImageExist(client *client.Client, image string) bool {
	images, err := client.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		fmt.Println(err)
		return false
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == image || tag == image+":latest" {
				return true
			}
		}
	}
	return false
}

func pullImages(ctx context.Context, client *client.Client, image string) error {
	out, err := client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(io.Discard, out)
	if err != nil {
		return err
	}
	return nil
}

func loadCpuAndMem(client *client.Client, container string) dto.ContainerListStats {
	data := dto.ContainerListStats{
		ContainerID: container,
	}
	res, err := client.ContainerStats(context.Background(), container, false)
	if err != nil {
		return data
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		return data
	}
	res.Body.Close()
	var stats *types.StatsJSON
	if err := json.Unmarshal(body, &stats); err != nil {
		return data
	}

	data.CPUTotalUsage = stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	data.SystemUsage = stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
	data.CPUPercent = calculateCPUPercentUnix(stats)
	data.PercpuUsage = len(stats.CPUStats.CPUUsage.PercpuUsage)

	data.MemroyCache = stats.MemoryStats.Stats["cache"]
	data.MemoryUsage = stats.MemoryStats.Usage
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
				return portMap, buserr.New(constant.ErrPortRules)
			}
			hostStart, _ := strconv.Atoi(strings.Split(port.HostPort, "-")[0])
			hostEnd, _ := strconv.Atoi(strings.Split(port.HostPort, "-")[1])
			containerStart, _ := strconv.Atoi(strings.Split(port.ContainerPort, "-")[0])
			containerEnd, _ := strconv.Atoi(strings.Split(port.ContainerPort, "-")[1])
			if (hostEnd-hostStart) <= 0 || (containerEnd-containerStart) <= 0 {
				return portMap, buserr.New(constant.ErrPortRules)
			}
			if (containerEnd - containerStart) != (hostEnd - hostStart) {
				return portMap, buserr.New(constant.ErrPortRules)
			}
			for i := 0; i <= hostEnd-hostStart; i++ {
				bindItem := nat.PortBinding{HostPort: strconv.Itoa(hostStart + i), HostIP: port.HostIP}
				portMap[nat.Port(fmt.Sprintf("%d/%s", containerStart+i, port.Protocol))] = []nat.PortBinding{bindItem}
			}
			for i := hostStart; i <= hostEnd; i++ {
				if common.ScanPort(i) {
					return portMap, buserr.WithDetail(constant.ErrPortInUsed, i, nil)
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
				return portMap, buserr.WithDetail(constant.ErrPortInUsed, portItem, nil)
			}
			bindItem := nat.PortBinding{HostPort: strconv.Itoa(portItem), HostIP: port.HostIP}
			portMap[nat.Port(fmt.Sprintf("%s/%s", port.ContainerPort, port.Protocol))] = []nat.PortBinding{bindItem}
		}
	}
	return portMap, nil
}

func loadConfigInfo(req dto.ContainerOperate, config *container.Config, hostConf *container.HostConfig, networkConf *network.NetworkingConfig) error {
	portMap, err := checkPortStats(req.ExposedPorts)
	if err != nil {
		return err
	}
	exposeds := make(nat.PortSet)
	for port := range portMap {
		exposeds[port] = struct{}{}
	}
	config.Image = req.Image
	config.Cmd = req.Cmd
	config.Entrypoint = req.Entrypoint
	config.Env = req.Env
	config.Labels = stringsToMap(req.Labels)
	config.ExposedPorts = exposeds

	if len(req.Network) != 0 {
		networkConf.EndpointsConfig = map[string]*network.EndpointSettings{req.Network: {}}
	} else {
		networkConf = &network.NetworkingConfig{}
	}

	hostConf.AutoRemove = req.AutoRemove
	hostConf.CPUShares = req.CPUShares
	hostConf.PublishAllPorts = req.PublishAllPorts
	hostConf.RestartPolicy = container.RestartPolicy{Name: req.RestartPolicy}
	if req.RestartPolicy == "on-failure" {
		hostConf.RestartPolicy.MaximumRetryCount = 5
	}
	hostConf.NanoCPUs = int64(req.NanoCPUs * 1000000000)
	hostConf.Memory = int64(req.Memory * 1024 * 1024)
	hostConf.PortBindings = portMap
	hostConf.Binds = []string{}
	config.Volumes = make(map[string]struct{})
	for _, volume := range req.Volumes {
		config.Volumes[volume.ContainerDir] = struct{}{}
		hostConf.Binds = append(hostConf.Binds, fmt.Sprintf("%s:%s:%s", volume.SourceDir, volume.ContainerDir, volume.Mode))
	}
	return nil
}

func reCreateAfterUpdate(name string, client *client.Client, config *container.Config, hostConf *container.HostConfig, networkConf *types.NetworkSettings) {
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
	if err := client.ContainerStart(ctx, oldContainer.ID, types.ContainerStartOptions{}); err != nil {
		global.LOG.Errorf("restart after container update failed, err: %v", err)
	}
}

func loadVolumeBinds(binds []string) []dto.VolumeHelper {
	var datas []dto.VolumeHelper
	for _, bind := range binds {
		parts := strings.Split(bind, ":")
		var volumeItem dto.VolumeHelper
		if len(parts) > 3 {
			continue
		}
		volumeItem.SourceDir = parts[0]
		if len(parts) == 1 {
			volumeItem.ContainerDir = parts[0]
			volumeItem.Mode = "rw"
		}
		if len(parts) == 2 {
			switch parts[1] {
			case "r", "ro":
				volumeItem.ContainerDir = parts[0]
				volumeItem.Mode = "ro"
			case "rw":
				volumeItem.ContainerDir = parts[0]
				volumeItem.Mode = "rw"
			default:
				volumeItem.ContainerDir = parts[1]
				volumeItem.Mode = "rw"
			}
		}
		if len(parts) == 3 {
			volumeItem.ContainerDir = parts[1]
			if parts[2] == "r" {
				volumeItem.Mode = "ro"
			} else {
				volumeItem.Mode = parts[2]
			}
		}
		datas = append(datas, volumeItem)
	}
	return datas
}

func simplifyPort(ports []types.Port) []string {
	var datas []string
	if len(ports) == 0 {
		return datas
	}
	if len(ports) == 1 {
		itemPortStr := fmt.Sprintf("%s:%v/%s", ports[0].IP, ports[0].PrivatePort, ports[0].Type)
		if ports[0].PublicPort != 0 {
			itemPortStr = fmt.Sprintf("%s:%v->%v/%s", ports[0].IP, ports[0].PublicPort, ports[0].PrivatePort, ports[0].Type)
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
				datas = append(datas, itemPortStr)
			} else {
				itemPortStr := fmt.Sprintf("%s:%v-%v/%s", start.IP, start.PrivatePort, ports[i-1].PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v-%v->%v-%v/%s", start.IP, start.PublicPort, ports[i-1].PublicPort, start.PrivatePort, ports[i-1].PrivatePort, start.Type)
				}
				datas = append(datas, itemPortStr)
			}
			start = ports[i]
		} else if i == len(ports)-1 {
			if ports[i].PrivatePort == start.PrivatePort {
				itemPortStr := fmt.Sprintf("%s:%v/%s", start.IP, start.PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v->%v/%s", start.IP, start.PublicPort, start.PrivatePort, start.Type)
				}
				datas = append(datas, itemPortStr)
			} else {
				itemPortStr := fmt.Sprintf("%s:%v-%v/%s", start.IP, start.PrivatePort, ports[i].PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v-%v->%v-%v/%s", start.IP, start.PublicPort, ports[i].PublicPort, start.PrivatePort, ports[i].PrivatePort, start.Type)
				}
				datas = append(datas, itemPortStr)
			}
		}
	}
	return datas
}
