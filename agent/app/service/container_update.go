package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func (u *ContainerService) ContainerUpdate(req dto.ContainerOperate) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskUpdate, task.TaskScopeContainer, req.TaskID, 1)
	if err != nil {
		_ = client.Close()
		global.LOG.Errorf("new task for create container failed, err: %v", err)
		return err
	}
	go func() {
		defer client.Close()
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

		taskItem.AddSubTaskWithOps(task.GetTaskName(req.Name, task.TaskUpdate, task.TaskScopeContainer), func(t *task.Task) error {
			t.LogStart(i18n.GetWithName("ContainerAcquireLock", req.Name))
			unlock := containerOperationLock.lock(req.Name)
			defer unlock()
			t.LogWithStatus(i18n.GetWithName("ContainerAcquireLock", req.Name), nil)

			oldContainer, err := client.ContainerInspect(ctx, req.Name)
			if err != nil {
				taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerLoadInfo"), err)
				return err
			}
			config, hostConf, networkConf, err := loadConfigInfo(false, req, &oldContainer)
			taskItem.LogWithStatus(i18n.GetMsgByKey("ContainerLoadInfo"), err)
			if err != nil {
				return err
			}
			normalizeContainerEndpointSettings(ctx, client, networkConf, nil)

			cleanupErr, err := switchContainer(ctx, client, req.Name, oldContainer, func() (container.CreateResponse, error) {
				return createContainerWithDynamicIPFallback(func() (container.CreateResponse, error) {
					return client.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
				}, networkConf.EndpointsConfig, oldContainer.NetworkSettings)
			}, newContainerSwitchTaskLogger(t))
			if err != nil {
				return fmt.Errorf("update container failed, err: %v", err)
			}
			if cleanupErr != nil {
				taskItem.Log(i18n.GetWithNameAndErr("ContainerCleanupWarning", containerSwitchBackupName(oldContainer.ID), cleanupErr))
			}
			return nil
		}, nil, 0, 0)

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
	ctx := context.Background()
	taskItem, err := task.NewTaskWithOps(req.Image, task.TaskUpgrade, task.TaskScopeImage, req.TaskID, 1)
	if err != nil {
		_ = client.Close()
		global.LOG.Errorf("new task for create container failed, err: %v", err)
		return err
	}
	go func() {
		defer client.Close()
		taskItem.AddSubTask(i18n.GetWithName("ContainerImagePull", req.Image), func(t *task.Task) error {
			taskItem.LogStart(i18n.GetWithName("ContainerImagePull", req.Image))
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
		var upgradeErrors []error
		for _, item := range req.Names {
			item := item
			taskItem.AddSubTaskWithIgnoreErr(i18n.GetWithName("ContainerUpgradeItem", item), func(t *task.Task) error {
				t.Logf("----------------- %s -----------------", item)
				t.LogStart(i18n.GetWithName("ContainerAcquireLock", item))
				unlock := containerOperationLock.lock(item)
				defer unlock()
				t.LogWithStatus(i18n.GetWithName("ContainerAcquireLock", item), nil)

				oldContainer, inspectErr := client.ContainerInspect(ctx, item)
				t.LogWithStatus(i18n.GetWithName("ContainerLoadInfo", item), inspectErr)
				if inspectErr != nil {
					err := fmt.Errorf("reload container %s failed: %w", item, inspectErr)
					upgradeErrors = append(upgradeErrors, err)
					return err
				}
				config := cloneContainerConfig(oldContainer.Config)
				config.Image = req.Image
				hostConf := cloneContainerHostConfig(oldContainer.HostConfig)
				preserveContainerVolumeMounts(hostConf, oldContainer.Mounts)
				cleanupErr, err := switchContainer(ctx, client, item, oldContainer, func() (container.CreateResponse, error) {
					return createContainerWithOldNetworks(ctx, client, config, hostConf, oldContainer.NetworkSettings, item)
				}, newContainerSwitchTaskLogger(t))
				if err != nil {
					upgradeErr := fmt.Errorf("upgrade container %s failed: %w", item, err)
					upgradeErrors = append(upgradeErrors, upgradeErr)
					return upgradeErr
				}
				if cleanupErr != nil {
					t.Log(i18n.GetWithNameAndErr("ContainerCleanupWarning", containerSwitchBackupName(oldContainer.ID), cleanupErr))
				}
				return nil
			})
		}
		taskItem.AddSubTask(i18n.GetMsgByKey("ContainerUpgradeSummary"), func(t *task.Task) error {
			return errors.Join(upgradeErrors...)
		}, nil)
		if err := taskItem.Execute(); err != nil {
			global.LOG.Error(err.Error())
		}
	}()

	return nil
}

type containerSwitchClient interface {
	ContainerStop(context.Context, string, container.StopOptions) error
	ContainerRename(context.Context, string, string) error
	ContainerStart(context.Context, string, container.StartOptions) error
	ContainerRemove(context.Context, string, container.RemoveOptions) error
	ContainerInspect(context.Context, string) (container.InspectResponse, error)
	NetworkConnect(context.Context, string, string, *network.EndpointSettings) error
	NetworkDisconnect(context.Context, string, string, bool) error
}

type containerOperationMutex struct {
	mutex sync.Mutex
	locks map[string]*containerOperationLockEntry
}

var containerOperationLock containerOperationMutex

type containerOperationLockEntry struct {
	mutex      sync.Mutex
	references int
}

func (l *containerOperationMutex) lock(names ...string) func() {
	nameSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name != "" {
			nameSet[name] = struct{}{}
		}
	}
	orderedNames := make([]string, 0, len(nameSet))
	for name := range nameSet {
		orderedNames = append(orderedNames, name)
	}
	sort.Strings(orderedNames)
	if len(orderedNames) == 0 {
		return func() {}
	}

	l.mutex.Lock()
	if l.locks == nil {
		l.locks = make(map[string]*containerOperationLockEntry)
	}
	entries := make([]*containerOperationLockEntry, 0, len(orderedNames))
	for _, name := range orderedNames {
		entry := l.locks[name]
		if entry == nil {
			entry = &containerOperationLockEntry{}
			l.locks[name] = entry
		}
		entry.references++
		entries = append(entries, entry)
	}
	l.mutex.Unlock()

	for _, entry := range entries {
		entry.mutex.Lock()
	}
	return func() {
		for index := len(entries) - 1; index >= 0; index-- {
			entries[index].mutex.Unlock()
		}
		l.mutex.Lock()
		defer l.mutex.Unlock()
		for index, name := range orderedNames {
			entry := entries[index]
			entry.references--
			if entry.references == 0 {
				delete(l.locks, name)
			}
		}
	}
}

type containerNetworkAttachment struct {
	name      string
	endpoint  *network.EndpointSettings
	isDynamic bool
}

type containerSwitchLogFunc func(messageKey, containerName string, err error)

func newContainerSwitchTaskLogger(t *task.Task) containerSwitchLogFunc {
	return func(messageKey, containerName string, err error) {
		t.LogWithStatus(i18n.GetWithName(messageKey, containerName), err)
	}
}

func logContainerSwitchStep(logger containerSwitchLogFunc, messageKey, containerName string, err error) {
	if logger != nil {
		logger(messageKey, containerName, err)
	}
}

// switchContainer keeps the stopped original container as the rollback target until the replacement starts.
func switchContainer(
	ctx context.Context,
	cli containerSwitchClient,
	name string,
	oldContainer container.InspectResponse,
	createNew func() (container.CreateResponse, error),
	logger containerSwitchLogFunc,
) (cleanupErr error, err error) {
	if oldContainer.ID == "" {
		return nil, fmt.Errorf("original container ID is empty")
	}
	wasRunning := oldContainer.State != nil && oldContainer.State.Running
	if wasRunning && oldContainer.HostConfig != nil && oldContainer.HostConfig.AutoRemove {
		return nil, fmt.Errorf("cannot safely replace container %s with auto-remove enabled", name)
	}

	if wasRunning {
		if err := cli.ContainerStop(ctx, oldContainer.ID, container.StopOptions{}); err != nil {
			logContainerSwitchStep(logger, "ContainerStopOld", name, err)
			current, inspectErr := cli.ContainerInspect(ctx, oldContainer.ID)
			if inspectErr == nil && current.State != nil && !current.State.Running {
				restartErr := restartOriginalContainer(ctx, cli, oldContainer.ID)
				logContainerSwitchStep(logger, "ContainerRollbackRestartOld", name, restartErr)
				return nil, errors.Join(fmt.Errorf("stop original container failed: %w", err), restartErr)
			}
			return nil, fmt.Errorf("stop original container failed: %w", err)
		}
		logContainerSwitchStep(logger, "ContainerStopOld", name, nil)
	}

	backupName := containerSwitchBackupName(oldContainer.ID)
	if err := cli.ContainerRename(ctx, oldContainer.ID, backupName); err != nil {
		logContainerSwitchStep(logger, "ContainerRenameOld", name, err)
		if wasRunning {
			restartErr := restartOriginalContainer(ctx, cli, oldContainer.ID)
			logContainerSwitchStep(logger, "ContainerRollbackRestartOld", name, restartErr)
			return nil, errors.Join(fmt.Errorf("rename original container failed: %w", err), restartErr)
		}
		return nil, fmt.Errorf("rename original container failed: %w", err)
	}
	logContainerSwitchStep(logger, "ContainerRenameOld", name, nil)
	disconnectedNetworks, disconnectErr := disconnectOriginalContainerNetworks(ctx, cli, oldContainer)
	logContainerSwitchStep(logger, "ContainerDisconnectOld", backupName, disconnectErr)
	if disconnectErr != nil {
		rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, "", disconnectedNetworks, logger)
		return nil, errors.Join(disconnectErr, rollbackErr)
	}

	created, createErr := createNew()
	logContainerSwitchStep(logger, "ContainerCreateReplacement", name, createErr)
	if createErr != nil {
		rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, created.ID, disconnectedNetworks, logger)
		return nil, errors.Join(createErr, rollbackErr)
	}
	if err := cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		logContainerSwitchStep(logger, "ContainerStartReplacement", name, err)
		rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, created.ID, disconnectedNetworks, logger)
		return nil, errors.Join(fmt.Errorf("start new container failed: %w", err), rollbackErr)
	}
	logContainerSwitchStep(logger, "ContainerStartReplacement", name, nil)
	if wasRunning {
		if err := waitContainerReady(ctx, cli, created.ID); err != nil {
			logContainerSwitchStep(logger, "ContainerWaitReplacement", name, err)
			rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, created.ID, disconnectedNetworks, logger)
			return nil, errors.Join(fmt.Errorf("new container readiness check failed: %w", err), rollbackErr)
		}
		logContainerSwitchStep(logger, "ContainerWaitReplacement", name, nil)
	}

	cleanupErr = cli.ContainerRemove(ctx, oldContainer.ID, container.RemoveOptions{Force: true, RemoveVolumes: false})
	logContainerSwitchStep(logger, "ContainerRemoveOld", backupName, cleanupErr)
	return cleanupErr, nil
}

const (
	containerStartStabilization = 10 * time.Second
	containerStartPollInterval  = time.Second
	containerHealthCheckMinWait = 30 * time.Second
	containerHealthCheckMaxWait = 10 * time.Minute
)

func waitContainerReady(ctx context.Context, cli containerSwitchClient, containerID string) error {
	info, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	if err := checkContainerRunningState(info); err != nil {
		return err
	}
	if info.State.Health == nil {
		return waitContainerStable(ctx, cli, containerID, info)
	}

	timeout := containerHealthCheckTimeout(info.Config)
	deadline := time.NewTimer(timeout)
	ticker := time.NewTicker(time.Second)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		if info.State.Restarting || info.RestartCount != 0 {
			return fmt.Errorf("container restarted %d times during startup", info.RestartCount)
		}
		if info.State.Health == nil {
			return fmt.Errorf("container health status is unavailable")
		}
		switch info.State.Health.Status {
		case container.Healthy:
			return nil
		case container.Unhealthy:
			return fmt.Errorf("container health status is unhealthy")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("container health check timed out after %s", timeout)
		case <-ticker.C:
			info, err = cli.ContainerInspect(ctx, containerID)
			if err != nil {
				return err
			}
			if err := checkContainerRunningState(info); err != nil {
				return err
			}
		}
	}
}

func waitContainerStable(ctx context.Context, cli containerSwitchClient, containerID string, initial container.InspectResponse) error {
	startedAt := initial.State.StartedAt
	if err := checkContainerStableState(initial, startedAt); err != nil {
		return err
	}
	deadline := time.NewTimer(containerStartStabilization)
	ticker := time.NewTicker(containerStartPollInterval)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			info, err := cli.ContainerInspect(ctx, containerID)
			if err != nil {
				return err
			}
			return checkContainerStableState(info, startedAt)
		case <-ticker.C:
			info, err := cli.ContainerInspect(ctx, containerID)
			if err != nil {
				return err
			}
			if err := checkContainerStableState(info, startedAt); err != nil {
				return err
			}
		}
	}
}

func checkContainerStableState(info container.InspectResponse, startedAt string) error {
	if err := checkContainerRunningState(info); err != nil {
		return err
	}
	if info.State.Restarting || info.RestartCount != 0 {
		return fmt.Errorf("container restarted %d times during startup", info.RestartCount)
	}
	if startedAt != "" && info.State.StartedAt != startedAt {
		return fmt.Errorf("container start time changed during startup")
	}
	return nil
}

func checkContainerRunningState(info container.InspectResponse) error {
	if info.State == nil {
		return fmt.Errorf("container state is unavailable")
	}
	if !info.State.Running {
		return fmt.Errorf("container exited with code %d: %s", info.State.ExitCode, info.State.Error)
	}
	return nil
}

func containerHealthCheckTimeout(config *container.Config) time.Duration {
	if config == nil || config.Healthcheck == nil {
		return containerHealthCheckMinWait
	}
	health := config.Healthcheck
	interval := health.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	checkTimeout := health.Timeout
	if checkTimeout <= 0 {
		checkTimeout = 30 * time.Second
	}
	retries := health.Retries
	if retries <= 0 {
		retries = 3
	}
	timeout := health.StartPeriod + time.Duration(retries)*(interval+checkTimeout)
	if timeout < containerHealthCheckMinWait {
		return containerHealthCheckMinWait
	}
	if timeout > containerHealthCheckMaxWait {
		return containerHealthCheckMaxWait
	}
	return timeout
}

func preserveContainerVolumeMounts(hostConfig *container.HostConfig, oldMounts []container.MountPoint) {
	if hostConfig == nil {
		return
	}
	for _, oldMount := range oldMounts {
		if oldMount.Type != mount.TypeVolume || oldMount.Name == "" || oldMount.Destination == "" {
			continue
		}
		preserved := false
		for index := range hostConfig.Mounts {
			if hostConfig.Mounts[index].Target == oldMount.Destination {
				hostConfig.Mounts[index].Type = mount.TypeVolume
				hostConfig.Mounts[index].Source = oldMount.Name
				preserved = true
			}
		}
		for index, raw := range hostConfig.Binds {
			destination, mode := containerBindDestinationAndMode(raw)
			if destination != oldMount.Destination {
				continue
			}
			hostConfig.Binds[index] = oldMount.Name + ":" + oldMount.Destination
			if mode != "" {
				hostConfig.Binds[index] += ":" + mode
			}
			preserved = true
		}
		if !preserved {
			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:     mount.TypeVolume,
				Source:   oldMount.Name,
				Target:   oldMount.Destination,
				ReadOnly: !oldMount.RW,
			})
		}
	}
}

func containerBindDestinationAndMode(raw string) (string, string) {
	parts := strings.SplitN(raw, ":", 3)
	switch len(parts) {
	case 1:
		return parts[0], ""
	case 2:
		return parts[1], ""
	default:
		return parts[1], parts[2]
	}
}

func containerSwitchBackupName(containerID string) string {
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}
	return "1panel-backup-" + containerID
}

func restartOriginalContainer(ctx context.Context, cli containerSwitchClient, containerID string) error {
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("restart original container failed: %w", err)
	}
	return nil
}

func disconnectOriginalContainerNetworks(ctx context.Context, cli containerSwitchClient, oldContainer container.InspectResponse) ([]containerNetworkAttachment, error) {
	primary, extras := buildContainerRecoverNetworkConfig(oldContainer.NetworkSettings, oldContainer.HostConfig)
	endpoints := make(map[string]*network.EndpointSettings, len(extras)+1)
	if primary != nil {
		for name, endpoint := range primary.EndpointsConfig {
			if name != "bridge" && endpoint != nil && endpoint.IPAMConfig != nil {
				endpoints[name] = endpoint
			}
		}
	}
	for name, endpoint := range extras {
		if name != "bridge" && endpoint != nil && endpoint.IPAMConfig != nil {
			endpoints[name] = endpoint
		}
	}
	names := make([]string, 0, len(endpoints))
	for name := range endpoints {
		names = append(names, name)
	}
	sort.Strings(names)

	disconnected := make([]containerNetworkAttachment, 0, len(names))
	for _, name := range names {
		if err := cli.NetworkDisconnect(ctx, name, oldContainer.ID, true); err != nil {
			return disconnected, fmt.Errorf("disconnect original container from network %s failed: %w", name, err)
		}
		disconnected = append(disconnected, containerNetworkAttachment{
			name:      name,
			endpoint:  endpoints[name],
			isDynamic: isDynamicContainerNetwork(oldContainer.NetworkSettings, name),
		})
	}
	return disconnected, nil
}

func reconnectOriginalContainerNetworks(ctx context.Context, cli containerSwitchClient, containerID string, attachments []containerNetworkAttachment) error {
	var reconnectErr error
	for _, attachment := range attachments {
		err := cli.NetworkConnect(ctx, attachment.name, containerID, attachment.endpoint)
		if err != nil && attachment.isDynamic && strings.Contains(err.Error(), unsupportedUserSpecifiedIPAddress) {
			attachment.endpoint.IPAMConfig = nil
			err = cli.NetworkConnect(ctx, attachment.name, containerID, attachment.endpoint)
		}
		if err != nil {
			reconnectErr = errors.Join(reconnectErr, fmt.Errorf("reconnect original container to network %s failed: %w", attachment.name, err))
		}
	}
	return reconnectErr
}

func restoreOriginalContainer(ctx context.Context, cli containerSwitchClient, oldContainerID, originalName string, wasRunning bool, newContainer string, disconnectedNetworks []containerNetworkAttachment, logger containerSwitchLogFunc) error {
	var rollbackErr error
	backupName := containerSwitchBackupName(oldContainerID)
	if newContainer != "" {
		removeErr := cli.ContainerRemove(ctx, newContainer, container.RemoveOptions{Force: true, RemoveVolumes: true})
		if client.IsErrNotFound(removeErr) {
			removeErr = nil
		}
		logContainerSwitchStep(logger, "ContainerRollbackRemoveReplacement", originalName, removeErr)
		if removeErr != nil {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf("remove failed replacement container failed: %w", removeErr))
		}
	}
	renameErr := cli.ContainerRename(ctx, oldContainerID, originalName)
	logContainerSwitchStep(logger, "ContainerRollbackRenameOld", backupName, renameErr)
	if renameErr != nil {
		rollbackErr = errors.Join(rollbackErr, fmt.Errorf("restore original container name failed: %w", renameErr))
	}
	currentName := originalName
	if renameErr != nil {
		currentName = backupName
	}
	reconnectErr := reconnectOriginalContainerNetworks(ctx, cli, oldContainerID, disconnectedNetworks)
	logContainerSwitchStep(logger, "ContainerRollbackReconnectOld", currentName, reconnectErr)
	rollbackErr = errors.Join(rollbackErr, reconnectErr)
	if wasRunning {
		restartErr := restartOriginalContainer(ctx, cli, oldContainerID)
		logContainerSwitchStep(logger, "ContainerRollbackRestartOld", currentName, restartErr)
		rollbackErr = errors.Join(rollbackErr, restartErr)
	}
	return rollbackErr
}

func createContainerWithOldNetworks(ctx context.Context, client *client.Client, config *container.Config, hostConf *container.HostConfig, networkSettings *container.NetworkSettings, name string) (container.CreateResponse, error) {
	networkConf, extraNetworks := buildContainerRecoverNetworkConfig(networkSettings, hostConf)
	normalizeContainerEndpointSettings(ctx, client, networkConf, extraNetworks)
	var primaryEndpoints map[string]*network.EndpointSettings
	if networkConf != nil {
		primaryEndpoints = networkConf.EndpointsConfig
	}

	created, err := createContainerWithDynamicIPFallback(func() (container.CreateResponse, error) {
		return client.ContainerCreate(ctx, config, hostConf, networkConf, nil, name)
	}, primaryEndpoints, networkSettings)
	if err != nil {
		return created, err
	}

	extraNames := make([]string, 0, len(extraNetworks))
	for item := range extraNetworks {
		extraNames = append(extraNames, item)
	}
	sort.Strings(extraNames)
	for _, item := range extraNames {
		err := client.NetworkConnect(ctx, item, created.ID, extraNetworks[item])
		if clearUnsupportedDynamicEndpointIPAM(err, map[string]*network.EndpointSettings{item: extraNetworks[item]}, networkSettings) {
			err = client.NetworkConnect(ctx, item, created.ID, extraNetworks[item])
		}
		if err != nil {
			_ = client.ContainerRemove(ctx, created.ID, container.RemoveOptions{Force: true})
			return created, err
		}
	}
	return created, nil
}

func createContainerWithDynamicIPFallback(
	create func() (container.CreateResponse, error),
	endpoints map[string]*network.EndpointSettings,
	networkSettings *container.NetworkSettings,
) (container.CreateResponse, error) {
	for {
		created, err := create()
		if err == nil || created.ID != "" || !clearUnsupportedDynamicEndpointIPAM(err, endpoints, networkSettings) {
			return created, err
		}
	}
}
