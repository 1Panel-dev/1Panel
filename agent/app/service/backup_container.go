package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/netip"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	dockerUtils "github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type containerBackupMeta struct {
	ContainerName string                 `json:"containerName"`
	ContainerID   string                 `json:"containerID"`
	CreatedAt     string                 `json:"createdAt"`
	Image         string                 `json:"image"`
	HostConfig    *container.HostConfig  `json:"hostConfig,omitempty"`
	Config        *container.Config      `json:"config,omitempty"`
	Mounts        []containerMountBackup `json:"mounts"`
}

type containerMountBackup struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination"`
	Mode        string `json:"mode,omitempty"`
	RW          bool   `json:"rw"`
	Propagation string `json:"propagation,omitempty"`
	BackupPath  string `json:"backupPath,omitempty"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

type containerBackupContext struct {
	containerName string
	backupDir     string
	fileName      string
	secret        string
	filePath      string
	tmpDir        string
	mountRoot     string
	wasRunning    bool
	stopped       bool
	fileOp        files.FileOp
	inspectInfo   container.InspectResponse
	meta          containerBackupMeta
}

type containerRecoverContext struct {
	req                dto.CommonRecover
	targetName         string
	fileOp             files.FileOp
	client             *client.Client
	tmpDir             string
	meta               containerBackupMeta
	inspectInfo        container.InspectResponse
	shouldStart        bool
	createdContainerID string
}

func (u *BackupService) ContainerBackup(req dto.CommonBackup) error {
	timeNow := time.Now().Format(constant.DateTimeSlimLayout) + common.RandStrAndNum(5)
	fileName := req.FileName
	if fileName == "" {
		fileName = fmt.Sprintf("%s_%s.tar.gz", req.Name, timeNow)
	}
	if !strings.HasSuffix(fileName, ".tar.gz") {
		fileName += ".tar.gz"
	}
	itemDir := fmt.Sprintf("container/%s", req.Name)
	backupDir := path.Join(global.Dir.LocalBackupDir, itemDir)
	record := &model.BackupRecord{
		Type:              req.Type,
		Name:              req.Name,
		SourceAccountIDs:  "1",
		DownloadAccountID: 1,
		FileDir:           itemDir,
		FileName:          fileName,
		TaskID:            req.TaskID,
		Status:            constant.StatusWaiting,
		Description:       req.Description,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	if err := handleContainerBackup(req.Name, nil, record.ID, backupDir, fileName, req.TaskID, req.Secret, req.StopBefore); err != nil {
		markBackupFailed(record.ID, err)
		return err
	}
	return nil
}

func (u *BackupService) ContainerRecover(req dto.CommonRecover) error {
	return handleContainerRecover(req, nil)
}

func handleContainerBackup(containerName string, parentTask *task.Task, recordID uint, backupDir, fileName, taskID, secret string, stopBefore bool) error {
	var (
		err        error
		backupTask *task.Task
	)
	backupCtx, err := newContainerBackupContext(containerName, backupDir, fileName, secret)
	if err != nil {
		return err
	}
	backupTask = parentTask
	if backupTask == nil {
		backupTask, err = task.NewTaskWithOps(containerName, task.TaskBackup, task.TaskScopeBackup, taskID, 1)
		if err != nil {
			return err
		}
	}
	if stopBefore {
		backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerBackupStop"), func(t *task.Task) error {
			return stepStopContainerForBackup(backupCtx)
		}, func(t *task.Task) {
			_ = stepStartContainerAfterBackup(backupCtx)
		}, 3, time.Hour)
	}
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerBackupPrepare"), func(t *task.Task) error {
		t.Logf("------------------ %s ------------------", containerName)
		return stepPrepareContainerBackup(backupCtx)
	}, nil, 3, time.Hour)
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerBackupInspect"), func(t *task.Task) error { return stepBackupContainerInspect(backupCtx) }, nil, 3, time.Hour)
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerBackupMounts"), func(t *task.Task) error { return stepBackupContainerMounts(backupCtx) }, nil, 3, time.Hour)
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerBackupMeta"), func(t *task.Task) error { return stepWriteContainerMeta(backupCtx) }, nil, 3, time.Hour)
	backupTask.AddSubTaskWithOps(task.GetTaskName(containerName, task.TaskBackup, task.TaskScopeBackup), func(t *task.Task) error { return stepPackContainerBackup(backupCtx) }, nil, 3, time.Hour)
	if stopBefore {
		backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerBackupStart"), func(t *task.Task) error {
			return stepStartContainerAfterBackup(backupCtx)
		}, nil, 3, time.Hour)
	}
	if parentTask != nil {
		return nil
	}
	go func() {
		defer backupCtx.close()
		if err := backupTask.Execute(); err != nil {
			markBackupFailed(recordID, err)
			return
		}
		backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusSuccess})
	}()
	return nil
}

func handleContainerRecover(req dto.CommonRecover, parentTask *task.Task) error {
	var (
		err         error
		recoverTask *task.Task
		recoverCtx  *containerRecoverContext
	)
	recoverTask = parentTask
	if recoverTask == nil {
		if isImportRecover(req) {
			taskName := i18n.GetMsgByKey("TaskImport") + i18n.GetMsgByKey("Container")
			recoverTask, err = task.NewTask(taskName, task.TaskImport, task.TaskScopeBackup, req.TaskID, 1)
			if err != nil {
				return err
			}
		} else {
			recoverTask, err = task.NewTaskWithOps("container", task.TaskRecover, task.TaskScopeBackup, req.TaskID, 1)
			if err != nil {
				return err
			}
		}
	}

	timeout := loadRecoverTimeout(req.Timeout)
	logName := strings.TrimSpace(req.Name)
	if logName == "" && req.File != "" {
		logName = strings.TrimSuffix(path.Base(req.File), ".tar.gz")
	}
	if logName == "" {
		logName = "container"
	}
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerRecoverPrepare"), func(t *task.Task) error {
		ctx, err := newContainerRecoverContext(req)
		if err != nil {
			return err
		}
		recoverCtx = ctx
		t.Logf("------------------ %s ------------------", logName)
		if err := stepPrepareContainerRecover(recoverCtx); err != nil {
			recoverCtx.close()
			recoverCtx = nil
			return err
		}
		return nil
	}, func(t *task.Task) {
		if recoverCtx != nil {
			recoverCtx.close()
			recoverCtx = nil
		}
	}, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerRecoverExtract"), func(t *task.Task) error { return stepExtractContainerRecover(recoverCtx) }, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerRecoverParse"), func(t *task.Task) error { return stepLoadContainerRecoverData(recoverCtx) }, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerRecoverCreate"), func(t *task.Task) error { return stepRecreateContainer(recoverCtx, t) }, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerRecoverMounts"), func(t *task.Task) error { return stepRestoreContainerMounts(recoverCtx) }, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerRecoverStart"), func(t *task.Task) error { return stepStartRecoveredContainer(recoverCtx) }, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerRecoverCleanup"), func(t *task.Task) error {
		if recoverCtx != nil {
			recoverCtx.close()
			recoverCtx = nil
		}
		return nil
	}, nil, 0, timeout)
	if parentTask != nil {
		return nil
	}
	go func() {
		_ = recoverTask.Execute()
	}()
	return nil
}

func loadRecoverTimeout(timeout int) time.Duration {
	switch timeout {
	case -1:
		return 0
	case 0:
		return 3 * time.Hour
	default:
		return time.Duration(timeout) * time.Second
	}
}

func isImportRecover(req dto.CommonRecover) bool {
	return req.BackupRecordID == 0
}

func newContainerBackupContext(containerName, backupDir, fileName, secret string) (*containerBackupContext, error) {
	dockerClient, err := dockerUtils.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = dockerClient.Close()
	}()
	inspectInfo, err := dockerClient.ContainerInspect(context.Background(), containerName)
	if err != nil {
		return nil, err
	}
	filePath := path.Join(backupDir, fileName)
	tmpDir := path.Join(path.Dir(filePath), strings.TrimSuffix(path.Base(filePath), ".tar.gz"))
	backupCtx := &containerBackupContext{
		containerName: containerName,
		backupDir:     backupDir,
		fileName:      fileName,
		secret:        secret,
		filePath:      filePath,
		tmpDir:        tmpDir,
		mountRoot:     path.Join(tmpDir, "mounts"),
		wasRunning:    inspectInfo.State != nil && inspectInfo.State.Running,
		fileOp:        files.NewFileOp(),
		inspectInfo:   inspectInfo,
		meta: containerBackupMeta{
			ContainerName: containerName,
			ContainerID:   inspectInfo.ID,
			CreatedAt:     time.Now().Format(constant.DateTimeLayout),
			Image:         inspectInfo.Config.Image,
			HostConfig:    inspectInfo.HostConfig,
			Config:        inspectInfo.Config,
			Mounts:        make([]containerMountBackup, 0),
		},
	}
	return backupCtx, nil
}

func newContainerRecoverContext(req dto.CommonRecover) (*containerRecoverContext, error) {
	dockerClient, err := dockerUtils.NewDockerClient()
	if err != nil {
		return nil, err
	}
	tmpDir := path.Join(path.Dir(req.File), strings.TrimSuffix(path.Base(req.File), ".tar.gz"))
	ctx := &containerRecoverContext{
		req:        req,
		targetName: req.Name,
		fileOp:     files.NewFileOp(),
		client:     dockerClient,
		tmpDir:     tmpDir,
		meta: containerBackupMeta{
			Mounts: make([]containerMountBackup, 0),
		},
	}
	return ctx, nil
}

func (c *containerBackupContext) close() {
	if c.tmpDir != "" {
		_ = os.RemoveAll(c.tmpDir)
	}
}

func (c *containerRecoverContext) close() {
	if c.client != nil {
		_ = c.client.Close()
	}
	if c.tmpDir != "" {
		_ = os.RemoveAll(c.tmpDir)
	}
}

func stepPrepareContainerBackup(backupCtx *containerBackupContext) error {
	if err := os.MkdirAll(backupCtx.backupDir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %s failed, err: %v", backupCtx.backupDir, err)
	}
	_ = os.RemoveAll(backupCtx.tmpDir)
	if err := os.MkdirAll(backupCtx.mountRoot, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func stepStopContainerForBackup(backupCtx *containerBackupContext) error {
	if !backupCtx.wasRunning || backupCtx.stopped {
		return nil
	}
	dockerClient, err := dockerUtils.NewDockerClient()
	if err != nil {
		return err
	}
	defer func() {
		_ = dockerClient.Close()
	}()
	if err := dockerClient.ContainerStop(context.Background(), backupCtx.inspectInfo.ID, container.StopOptions{}); err != nil {
		return err
	}
	backupCtx.stopped = true
	return nil
}

func stepStartContainerAfterBackup(backupCtx *containerBackupContext) error {
	if !backupCtx.stopped {
		return nil
	}
	dockerClient, err := dockerUtils.NewDockerClient()
	if err != nil {
		return err
	}
	defer func() {
		_ = dockerClient.Close()
	}()
	if err := dockerClient.ContainerStart(context.Background(), backupCtx.inspectInfo.ID, container.StartOptions{}); err != nil {
		return err
	}
	backupCtx.stopped = false
	return nil
}

func stepBackupContainerInspect(backupCtx *containerBackupContext) error {
	inspectBytes, err := json.MarshalIndent(backupCtx.inspectInfo, "", "  ")
	if err != nil {
		return err
	}
	if err := backupCtx.fileOp.SaveFile(path.Join(backupCtx.tmpDir, "inspect.json"), string(inspectBytes), fs.ModePerm); err != nil {
		return err
	}
	if backupCtx.inspectInfo.NetworkSettings != nil {
		networkBytes, err := json.MarshalIndent(backupCtx.inspectInfo.NetworkSettings, "", "  ")
		if err != nil {
			return err
		}
		if err := backupCtx.fileOp.SaveFile(path.Join(backupCtx.tmpDir, "network.json"), string(networkBytes), fs.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func stepBackupContainerMounts(backupCtx *containerBackupContext) error {
	var (
		dockerClient *client.Client
		clientErr    error
	)
	ensureClient := func() (*client.Client, error) {
		if dockerClient != nil || clientErr != nil {
			return dockerClient, clientErr
		}
		dockerClient, clientErr = dockerUtils.NewDockerClient()
		return dockerClient, clientErr
	}
	defer func() {
		if dockerClient != nil {
			_ = dockerClient.Close()
		}
	}()

	for i, item := range backupCtx.inspectInfo.Mounts {
		mountMeta := containerMountBackup{
			Type:        string(item.Type),
			Name:        item.Name,
			Source:      item.Source,
			Destination: item.Destination,
			Mode:        item.Mode,
			RW:          item.RW,
			Propagation: string(item.Propagation),
			Status:      "skipped",
		}

		mountDirName := fmt.Sprintf("%02d_%s", i, sanitizeContainerMountName(item.Destination))
		mountDir := path.Join(backupCtx.mountRoot, mountDirName)
		mountMeta.BackupPath = path.Join("mounts", mountDirName, "data")

		switch item.Type {
		case mount.TypeBind, mount.TypeVolume:
			if item.Source == "" {
				mountMeta.Message = "empty source"
				backupCtx.meta.Mounts = append(backupCtx.meta.Mounts, mountMeta)
				continue
			}
			sourceInfo, statErr := os.Stat(item.Source)
			if statErr != nil {
				mountMeta.Message = statErr.Error()
				backupCtx.meta.Mounts = append(backupCtx.meta.Mounts, mountMeta)
				continue
			}
			dataDir := path.Join(mountDir, "data")
			if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
				return err
			}
			if sourceInfo.IsDir() {
				if err := backupCtx.fileOp.CopyDirWithNewName(item.Source, dataDir, "."); err != nil {
					return err
				}
			} else {
				if err := backupCtx.fileOp.CopyFile(item.Source, dataDir); err != nil {
					return err
				}
			}

			if item.Type == mount.TypeVolume && item.Name != "" {
				cli, err := ensureClient()
				if err != nil {
					return err
				}
				volumeInfo, volumeErr := cli.VolumeInspect(context.Background(), item.Name)
				if volumeErr == nil {
					volumeBytes, volumeMarshalErr := json.MarshalIndent(volumeInfo, "", "  ")
					if volumeMarshalErr == nil {
						_ = backupCtx.fileOp.SaveFile(path.Join(mountDir, "volume.json"), string(volumeBytes), fs.ModePerm)
					}
				}
			}
			mountMeta.Status = "backed_up"
		default:
			mountMeta.Message = "mount type not supported for data backup"
		}
		backupCtx.meta.Mounts = append(backupCtx.meta.Mounts, mountMeta)
	}
	return nil
}

func stepWriteContainerMeta(backupCtx *containerBackupContext) error {
	metaBytes, err := json.MarshalIndent(backupCtx.meta, "", "  ")
	if err != nil {
		return err
	}
	if err := backupCtx.fileOp.SaveFile(path.Join(backupCtx.tmpDir, "meta.json"), string(metaBytes), fs.ModePerm); err != nil {
		return err
	}
	return nil
}

func stepPackContainerBackup(backupCtx *containerBackupContext) error {
	if err := backupCtx.fileOp.TarGzCompressPro(true, backupCtx.tmpDir, backupCtx.filePath, backupCtx.secret, ""); err != nil {
		return err
	}
	return nil
}

func stepPrepareContainerRecover(recoverCtx *containerRecoverContext) error {
	if !recoverCtx.fileOp.Stat(recoverCtx.req.File) {
		return buserr.WithName("ErrFileNotFound", recoverCtx.req.File)
	}
	_ = os.RemoveAll(recoverCtx.tmpDir)
	return nil
}

func stepExtractContainerRecover(recoverCtx *containerRecoverContext) error {
	return recoverCtx.fileOp.TarGzExtractPro(recoverCtx.req.File, path.Dir(recoverCtx.req.File), recoverCtx.req.Secret)
}

func stepLoadContainerRecoverData(recoverCtx *containerRecoverContext) error {
	if err := loadContainerRecoverMeta(recoverCtx); err != nil {
		return err
	}
	if err := loadContainerRecoverInspect(recoverCtx); err != nil {
		return err
	}
	if recoverCtx.targetName == "" {
		recoverCtx.targetName = strings.TrimPrefix(recoverCtx.inspectInfo.Name, "/")
	}
	if recoverCtx.targetName == "" {
		recoverCtx.targetName = recoverCtx.meta.ContainerName
	}
	if recoverCtx.targetName == "" {
		return fmt.Errorf("container name not found in recover request or backup file")
	}
	if recoverCtx.inspectInfo.Config == nil {
		recoverCtx.inspectInfo.Config = recoverCtx.meta.Config
	}
	if recoverCtx.inspectInfo.HostConfig == nil {
		recoverCtx.inspectInfo.HostConfig = recoverCtx.meta.HostConfig
	}
	if recoverCtx.inspectInfo.Config == nil {
		return fmt.Errorf("container config not found in backup file")
	}
	if recoverCtx.inspectInfo.HostConfig == nil {
		recoverCtx.inspectInfo.HostConfig = &container.HostConfig{}
	}
	recoverCtx.shouldStart = recoverCtx.inspectInfo.State != nil && recoverCtx.inspectInfo.State.Running
	return nil
}

func loadContainerRecoverMeta(recoverCtx *containerRecoverContext) error {
	metaPath := path.Join(recoverCtx.tmpDir, "meta.json")
	if !recoverCtx.fileOp.Stat(metaPath) {
		return nil
	}
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(metaBytes, &recoverCtx.meta); err != nil {
		return fmt.Errorf("unmarshal meta.json failed, err: %v", err)
	}
	return nil
}

func loadContainerRecoverInspect(recoverCtx *containerRecoverContext) error {
	inspectPath := path.Join(recoverCtx.tmpDir, "inspect.json")
	if !recoverCtx.fileOp.Stat(inspectPath) {
		return fmt.Errorf("inspect.json not found in backup file")
	}
	inspectBytes, err := os.ReadFile(inspectPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(inspectBytes, &recoverCtx.inspectInfo); err != nil {
		return fmt.Errorf("unmarshal inspect.json failed, err: %v", err)
	}
	return nil
}

func stepRecreateContainer(recoverCtx *containerRecoverContext, taskItem *task.Task) error {
	ctx := context.Background()
	if err := ensureContainerRecoverNetworks(recoverCtx); err != nil {
		return err
	}
	if err := ensureContainerRecoverVolumes(recoverCtx); err != nil {
		return err
	}

	config := cloneContainerConfig(recoverCtx.inspectInfo.Config)
	hostConfig := cloneContainerHostConfig(recoverCtx.inspectInfo.HostConfig)
	if config.Image == "" {
		config.Image = recoverCtx.meta.Image
	}
	if config.Image == "" {
		return fmt.Errorf("container image not found in backup file")
	}
	if !checkImageExist(recoverCtx.client, config.Image) {
		if err := pullImages(taskItem, recoverCtx.client, config.Image); err != nil {
			return err
		}
	}

	if _, err := recoverCtx.client.ContainerInspect(ctx, recoverCtx.targetName); err == nil {
		if err := recoverCtx.client.ContainerRemove(ctx, recoverCtx.targetName, container.RemoveOptions{Force: true, RemoveVolumes: false}); err != nil {
			return err
		}
	} else if !client.IsErrNotFound(err) {
		return err
	}

	createRes, err := createContainerWithOldNetworks(ctx, recoverCtx.client, config, hostConfig, recoverCtx.inspectInfo.NetworkSettings, recoverCtx.targetName)
	if err != nil {
		return err
	}
	recoverCtx.createdContainerID = createRes.ID
	return nil
}

func removeUnsupportedEndpointStaticIPAM(cli *client.Client, primary *network.NetworkingConfig, extras map[string]*network.EndpointSettings) {
	if primary != nil {
		removeUnsupportedEndpointStaticIPAMFromEndpoints(cli, primary.EndpointsConfig)
	}
	removeUnsupportedEndpointStaticIPAMFromEndpoints(cli, extras)
}

func removeUnsupportedEndpointStaticIPAMFromEndpoints(cli *client.Client, endpoints map[string]*network.EndpointSettings) {
	for netName, endpoint := range endpoints {
		if endpoint == nil || endpoint.IPAMConfig == nil {
			continue
		}
		info, err := cli.NetworkInspect(context.Background(), netName, network.InspectOptions{})
		if err != nil {
			continue
		}
		removeUnsupportedEndpointStaticIP(netName, info, endpoint)
	}
}

func removeUnsupportedEndpointStaticIP(netName string, info network.Inspect, endpoint *network.EndpointSettings) {
	if endpoint == nil || endpoint.IPAMConfig == nil {
		return
	}
	if isDefaultBridgeNetwork(netName, info) {
		endpoint.IPAMConfig = nil
		return
	}

	if endpoint.IPAMConfig.IPv4Address != "" && !networkSupportsStaticIP(info, endpoint.IPAMConfig.IPv4Address, false) {
		endpoint.IPAMConfig.IPv4Address = ""
	}
	if endpoint.IPAMConfig.IPv6Address != "" && !networkSupportsStaticIP(info, endpoint.IPAMConfig.IPv6Address, true) {
		endpoint.IPAMConfig.IPv6Address = ""
	}
	if endpoint.IPAMConfig.IPv4Address == "" && endpoint.IPAMConfig.IPv6Address == "" {
		endpoint.IPAMConfig = nil
	}
}

func isDefaultBridgeNetwork(netName string, info network.Inspect) bool {
	return info.Driver == "bridge" && (netName == "bridge" || info.Name == "bridge")
}

func networkSupportsStaticIP(info network.Inspect, ip string, isIPv6 bool) bool {
	if ip == "" {
		return true
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	if addr.Is6() != isIPv6 {
		return false
	}

	for _, item := range info.IPAM.Config {
		if item.Subnet == "" {
			continue
		}
		subnet, err := netip.ParsePrefix(item.Subnet)
		if err != nil || subnet.Addr().Is6() != isIPv6 {
			continue
		}
		if subnet.Contains(addr) {
			return true
		}
	}
	return false
}

func ensureContainerRecoverNetworks(recoverCtx *containerRecoverContext) error {
	if recoverCtx.inspectInfo.NetworkSettings == nil {
		return nil
	}
	for netName := range recoverCtx.inspectInfo.NetworkSettings.Networks {
		if netName == "" || netName == "bridge" || netName == "host" || netName == "none" {
			continue
		}
		if _, err := recoverCtx.client.NetworkInspect(context.Background(), netName, network.InspectOptions{}); err != nil {
			if !client.IsErrNotFound(err) {
				return err
			}
			if _, err := recoverCtx.client.NetworkCreate(context.Background(), netName, network.CreateOptions{Driver: "bridge"}); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureContainerRecoverVolumes(recoverCtx *containerRecoverContext) error {
	for _, item := range recoverCtx.meta.Mounts {
		if item.Type != string(mount.TypeVolume) || item.Name == "" {
			continue
		}
		if _, err := recoverCtx.client.VolumeInspect(context.Background(), item.Name); err == nil {
			continue
		} else if !client.IsErrNotFound(err) {
			return err
		}
		createOptions := volume.CreateOptions{Name: item.Name}
		if item.BackupPath != "" {
			volumeMetaPath := path.Join(recoverCtx.tmpDir, path.Dir(item.BackupPath), "volume.json")
			if recoverCtx.fileOp.Stat(volumeMetaPath) {
				volumeBytes, readErr := os.ReadFile(volumeMetaPath)
				if readErr != nil {
					return readErr
				}
				var volumeInfo volume.Volume
				if unmarshalErr := json.Unmarshal(volumeBytes, &volumeInfo); unmarshalErr != nil {
					return unmarshalErr
				}
				if volumeInfo.Driver != "" {
					createOptions.Driver = volumeInfo.Driver
				}
				if len(volumeInfo.Options) != 0 {
					createOptions.DriverOpts = volumeInfo.Options
				}
				if len(volumeInfo.Labels) != 0 {
					createOptions.Labels = volumeInfo.Labels
				}
			}
		}
		if _, err := recoverCtx.client.VolumeCreate(context.Background(), createOptions); err != nil {
			return err
		}
	}
	return nil
}

func buildContainerRecoverNetworkConfig(networkSettings *container.NetworkSettings, hostConfig *container.HostConfig) (*network.NetworkingConfig, map[string]*network.EndpointSettings) {
	extraNetworks := make(map[string]*network.EndpointSettings)
	if hostConfig != nil {
		networkMode := string(hostConfig.NetworkMode)
		if networkMode == "host" || networkMode == "none" {
			return nil, extraNetworks
		}
	}
	if networkSettings == nil || len(networkSettings.Networks) == 0 {
		return nil, extraNetworks
	}

	primaryName := ""
	if hostConfig != nil {
		networkMode := string(hostConfig.NetworkMode)
		if networkMode != "" && networkMode != "default" && networkMode != "bridge" {
			if _, ok := networkSettings.Networks[networkMode]; ok {
				primaryName = networkMode
			}
		}
	}
	if primaryName == "" {
		if _, ok := networkSettings.Networks["bridge"]; ok {
			primaryName = "bridge"
		} else {
			names := make([]string, 0, len(networkSettings.Networks))
			for name := range networkSettings.Networks {
				names = append(names, name)
			}
			sort.Strings(names)
			if len(names) > 0 {
				primaryName = names[0]
			}
		}
	}

	config := &network.NetworkingConfig{EndpointsConfig: make(map[string]*network.EndpointSettings)}
	for name, endpoint := range networkSettings.Networks {
		if name == "host" || name == "none" {
			continue
		}
		endpointSetting := &network.EndpointSettings{Aliases: append([]string(nil), endpoint.Aliases...), MacAddress: endpoint.MacAddress}
		if endpoint.IPAMConfig != nil {
			endpointSetting.IPAMConfig = &network.EndpointIPAMConfig{
				IPv4Address: endpoint.IPAMConfig.IPv4Address,
				IPv6Address: endpoint.IPAMConfig.IPv6Address,
			}
		}
		if name == primaryName {
			config.EndpointsConfig[name] = endpointSetting
		} else {
			extraNetworks[name] = endpointSetting
		}
	}
	if len(config.EndpointsConfig) == 0 {
		return nil, extraNetworks
	}
	return config, extraNetworks
}

func cloneContainerConfig(config *container.Config) *container.Config {
	if config == nil {
		return &container.Config{}
	}
	item := *config
	if len(config.Env) != 0 {
		item.Env = append([]string(nil), config.Env...)
	}
	if len(config.Cmd) != 0 {
		item.Cmd = append([]string(nil), config.Cmd...)
	}
	if len(config.Entrypoint) != 0 {
		item.Entrypoint = append([]string(nil), config.Entrypoint...)
	}
	if len(config.Labels) != 0 {
		labels := make(map[string]string, len(config.Labels))
		for key, val := range config.Labels {
			labels[key] = val
		}
		item.Labels = labels
	}
	if len(config.Volumes) != 0 {
		volumes := make(map[string]struct{}, len(config.Volumes))
		for key, val := range config.Volumes {
			volumes[key] = val
		}
		item.Volumes = volumes
	}
	return &item
}

func cloneContainerHostConfig(hostConfig *container.HostConfig) *container.HostConfig {
	if hostConfig == nil {
		return &container.HostConfig{}
	}
	item := *hostConfig
	if len(hostConfig.Binds) != 0 {
		item.Binds = append([]string(nil), hostConfig.Binds...)
	}
	if len(hostConfig.DNS) != 0 {
		item.DNS = append([]string(nil), hostConfig.DNS...)
	}
	if len(hostConfig.ExtraHosts) != 0 {
		item.ExtraHosts = append([]string(nil), hostConfig.ExtraHosts...)
	}
	if len(hostConfig.Mounts) != 0 {
		item.Mounts = append([]mount.Mount(nil), hostConfig.Mounts...)
	}
	return &item
}

func stepRestoreContainerMounts(recoverCtx *containerRecoverContext) error {
	currentContainer := recoverCtx.createdContainerID
	if currentContainer == "" {
		currentContainer = recoverCtx.targetName
	}
	currentInspect, err := recoverCtx.client.ContainerInspect(context.Background(), currentContainer)
	if err != nil {
		return err
	}
	currentMounts := make(map[string]container.MountPoint, len(currentInspect.Mounts))
	for _, item := range currentInspect.Mounts {
		currentMounts[item.Destination] = item
	}

	for _, item := range recoverCtx.meta.Mounts {
		if item.Status != "backed_up" || item.BackupPath == "" || !item.RW {
			continue
		}
		backupPath := path.Join(recoverCtx.tmpDir, item.BackupPath)
		if !recoverCtx.fileOp.Stat(backupPath) {
			continue
		}
		sourcePath := item.Source
		if currentMount, ok := currentMounts[item.Destination]; ok {
			if currentMount.Source != "" {
				sourcePath = currentMount.Source
			}
			if item.Type == string(mount.TypeVolume) && item.Name == "" {
				item.Name = currentMount.Name
			}
		}
		if sourcePath == "" && item.Type == string(mount.TypeVolume) && item.Name != "" {
			volumeInfo, volumeErr := recoverCtx.client.VolumeInspect(context.Background(), item.Name)
			if volumeErr != nil {
				return volumeErr
			}
			sourcePath = volumeInfo.Mountpoint
		}
		if sourcePath == "" {
			continue
		}
		if err := restoreContainerMountData(recoverCtx.fileOp, backupPath, sourcePath); err != nil {
			return err
		}
	}
	return nil
}

func restoreContainerMountData(fileOp files.FileOp, backupPath, sourcePath string) error {
	if sourcePath == "/" {
		return fmt.Errorf("invalid mount source path /")
	}
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		return err
	}
	if len(entries) == 1 && !entries[0].IsDir() && entries[0].Name() == path.Base(sourcePath) {
		if err := os.MkdirAll(path.Dir(sourcePath), os.ModePerm); err != nil {
			return err
		}
		_ = os.RemoveAll(sourcePath)
		if err := fileOp.CopyFile(path.Join(backupPath, entries[0].Name()), path.Dir(sourcePath)); err != nil {
			return err
		}
		return nil
	}

	_ = os.RemoveAll(sourcePath)
	if err := os.MkdirAll(sourcePath, os.ModePerm); err != nil {
		return err
	}
	if err := fileOp.CopyDirWithNewName(backupPath, sourcePath, "."); err != nil {
		return err
	}
	return nil
}

func stepStartRecoveredContainer(recoverCtx *containerRecoverContext) error {
	if !recoverCtx.shouldStart {
		return nil
	}
	containerID := recoverCtx.createdContainerID
	if containerID == "" {
		containerID = recoverCtx.targetName
	}
	return recoverCtx.client.ContainerStart(context.Background(), containerID, container.StartOptions{})
}

func sanitizeContainerMountName(in string) string {
	name := strings.TrimSpace(in)
	name = strings.Trim(name, "/")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	if name == "" {
		return "root"
	}
	return name
}
