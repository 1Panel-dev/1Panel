package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
	"github.com/docker/docker/pkg/stdcopy"
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
			if err := normalizeContainerEndpointSettings(ctx, client, networkConf, nil); err != nil {
				return err
			}

			cleanupErr, err := switchContainer(ctx, client, req.Name, oldContainer, func() (container.CreateResponse, error) {
				return client.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
			}, config.Tty, t)
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

const containerUpgradeConcurrency = 2

// Shared by all upgrade requests handled by this agent.
var containerUpgradeSlots = make(chan struct{}, containerUpgradeConcurrency)

func (u *ContainerService) ContainerUpgrade(req dto.ContainerUpgrade) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	taskItem, err := task.NewTaskWithOps(req.Image, task.TaskUpgrade, task.TaskScopeImage, req.TaskID, 1)
	if err != nil {
		_ = client.Close()
		global.LOG.Errorf("new task for create container failed, err: %v", err)
		return err
	}
	taskItem.AddSubTask(i18n.GetWithName("ContainerImagePull", req.Image), func(t *task.Task) error {
		ctx, cancel := context.WithTimeout(context.Background(), containerUpgradePullTimeout)
		defer cancel()
		taskItem.LogStart(i18n.GetWithName("ContainerImagePull", req.Image))
		if req.ForcePull || !checkImageExistContext(ctx, client, req.Image) {
			if err := pullImagesContext(ctx, taskItem, client, req.Image); err != nil {
				return fmt.Errorf("pull image %s failed: %w", req.Image, err)
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
			ctx, cancel := context.WithTimeout(context.Background(), containerUpgradeReplaceTimeout)
			defer cancel()
			unlock, lockErr := containerOperationLock.lockContext(ctx, item)
			if lockErr != nil {
				upgradeErrors = append(upgradeErrors, fmt.Errorf("wait for container %s lock: %w", item, lockErr))
				return lockErr
			}
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
			if config.Labels == nil {
				config.Labels = make(map[string]string)
			}
			config.Labels[containerUpgradeOwnerLabel] = oldContainer.ID
			hostConf := cloneContainerHostConfig(oldContainer.HostConfig)
			preserveContainerVolumeMounts(hostConf, oldContainer.Mounts)
			networkConf, extraNetworks := buildContainerRecoverNetworkConfig(oldContainer.NetworkSettings, hostConf)
			if err := normalizeContainerEndpointSettings(ctx, client, networkConf, extraNetworks); err != nil {
				upgradeErr := fmt.Errorf("prepare networks for container %s failed: %w", item, err)
				upgradeErrors = append(upgradeErrors, upgradeErr)
				return upgradeErr
			}
			cleanupErr, err := switchContainerWithRecovery(ctx, client, item, oldContainer, func() (container.CreateResponse, error) {
				return createContainerWithNetworks(ctx, client, config, hostConf, networkConf, extraNetworks, item)
			}, config.Tty, t, true)
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
	// Stage contexts own the deadlines, including a separate recovery budget.
	for _, subTask := range taskItem.SubTasks {
		subTask.Timeout = 0
	}
	if err := taskItem.ExecuteQueued(containerUpgradeSlots, "ContainerUpgradeWaiting", func() { _ = client.Close() }); err != nil {
		_ = client.Close()
		return err
	}
	return nil
}

type containerSwitchClient interface {
	ContainerStop(context.Context, string, container.StopOptions) error
	ContainerRename(context.Context, string, string) error
	ContainerStart(context.Context, string, container.StartOptions) error
	ContainerRemove(context.Context, string, container.RemoveOptions) error
	ContainerInspect(context.Context, string) (container.InspectResponse, error)
	ContainerLogs(context.Context, string, container.LogsOptions) (io.ReadCloser, error)
	NetworkConnect(context.Context, string, string, *network.EndpointSettings) error
	NetworkDisconnect(context.Context, string, string, bool) error
}

type containerInspectClient interface {
	ContainerInspect(context.Context, string) (container.InspectResponse, error)
}

type containerOperationMutex struct {
	mutex sync.Mutex
	locks map[string]*containerOperationLockEntry
}

var containerOperationLock containerOperationMutex

type containerOperationLockEntry struct {
	token      chan struct{}
	references int
}

func (l *containerOperationMutex) lock(names ...string) func() {
	unlock, _ := l.lockContext(context.Background(), names...)
	return unlock
}

func (l *containerOperationMutex) lockContext(ctx context.Context, names ...string) (func(), error) {
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
		return func() {}, nil
	}

	l.mutex.Lock()
	if l.locks == nil {
		l.locks = make(map[string]*containerOperationLockEntry)
	}
	entries := make([]*containerOperationLockEntry, 0, len(orderedNames))
	for _, name := range orderedNames {
		entry := l.locks[name]
		if entry == nil {
			entry = &containerOperationLockEntry{token: make(chan struct{}, 1)}
			l.locks[name] = entry
		}
		entry.references++
		entries = append(entries, entry)
	}
	l.mutex.Unlock()

	acquired := 0
	release := func() {
		for index := acquired - 1; index >= 0; index-- {
			<-entries[index].token
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
	for _, entry := range entries {
		select {
		case entry.token <- struct{}{}:
			acquired++
		case <-ctx.Done():
			release()
			return nil, ctx.Err()
		}
	}
	if err := ctx.Err(); err != nil {
		release()
		return nil, err
	}
	return release, nil
}

type containerNetworkAttachment struct {
	name     string
	endpoint *network.EndpointSettings
}

type containerSwitchLogger interface {
	LogWithStatus(string, error)
	Log(string)
}

func logContainerSwitchStep(logger containerSwitchLogger, messageKey, containerName string, err error) {
	if logger != nil {
		logger.LogWithStatus(i18n.GetWithName(messageKey, containerName), err)
	}
}

// switchContainer keeps the stopped original container as the rollback target until the replacement starts.
func switchContainer(
	ctx context.Context,
	cli containerSwitchClient,
	name string,
	oldContainer container.InspectResponse,
	createNew func() (container.CreateResponse, error),
	tty bool,
	logger containerSwitchLogger,
) (cleanupErr error, err error) {
	return switchContainerWithRecovery(ctx, cli, name, oldContainer, createNew, tty, logger, false)
}


const (
	containerUpgradePullTimeout     = 30 * time.Minute
	containerUpgradeReplaceTimeout  = 15 * time.Minute
	containerUpgradeRecoveryTimeout = 5 * time.Minute
	containerUpgradeOwnerLabel      = "io.1panel.upgrade.original-id"
)

// Reconcile against Docker's actual state: a failed request may already have
// changed names, created a container, or disconnected a network on the daemon.
func recoverContainerUpgrade(ctx context.Context, cli containerSwitchClient, name string, old container.InspectResponse, newID string, logger containerSwitchLogger) error {
	original, err := cli.ContainerInspect(ctx, old.ID)
	if err != nil {
		return fmt.Errorf("inspect original %s: %w", old.ID, err)
	}
	if original.State == nil {
		return fmt.Errorf("original container state unavailable")
	}

	replacement, inspectErr := cli.ContainerInspect(ctx, name)
	if inspectErr != nil && !client.IsErrNotFound(inspectErr) {
		return fmt.Errorf("inspect replacement by name: %w", inspectErr)
	}
	if inspectErr == nil && replacement.ID != old.ID {
		if replacement.Config == nil || replacement.Config.Labels[containerUpgradeOwnerLabel] != old.ID {
			return fmt.Errorf("container %s belongs to another operation; refusing to remove it", name)
		}
		if newID != "" && replacement.ID != newID {
			return fmt.Errorf("replacement ID changed; refusing to remove it")
		}
		newID = replacement.ID
	}
	if newID != "" {
		replacement, inspectErr = cli.ContainerInspect(ctx, newID)
		if inspectErr != nil && !client.IsErrNotFound(inspectErr) {
			return fmt.Errorf("inspect replacement %s: %w", newID, inspectErr)
		}
		if inspectErr == nil {
			if replacement.ID == old.ID || replacement.Config == nil || replacement.Config.Labels[containerUpgradeOwnerLabel] != old.ID {
				return fmt.Errorf("replacement identity could not be confirmed")
			}
			err = cli.ContainerRemove(ctx, newID, container.RemoveOptions{Force: true, RemoveVolumes: false})
			logContainerSwitchStep(logger, "ContainerRollbackRemoveReplacement", name, err)
			if err != nil && !client.IsErrNotFound(err) {
				return fmt.Errorf("remove replacement: %w", err)
			}
		}
	}
	if strings.TrimPrefix(original.Name, "/") != name {
		err = cli.ContainerRename(ctx, old.ID, name)
		logContainerSwitchStep(logger, "ContainerRollbackRenameOld", name, err)
		if err != nil {
			return fmt.Errorf("restore original name: %w", err)
		}
	}
	// Reload after mutations; connect only networks which are actually missing.
	original, err = cli.ContainerInspect(ctx, old.ID)
	if err != nil {
		return err
	}
	primary, extras := buildContainerRecoverNetworkConfig(old.NetworkSettings, old.HostConfig)
	wanted := make(map[string]*network.EndpointSettings)
	if primary != nil {
		for n, endpoint := range primary.EndpointsConfig {
			if n != "bridge" {
				wanted[n] = endpoint
			}
		}
	}
	for n, endpoint := range extras {
		if n != "bridge" {
			wanted[n] = endpoint
		}
	}
	var reconnectErr error
	for n, endpoint := range wanted {
		if original.NetworkSettings != nil && original.NetworkSettings.Networks[n] != nil {
			continue
		}
		err = cli.NetworkConnect(ctx, n, old.ID, endpoint)
		logContainerSwitchStep(logger, "ContainerRollbackReconnectOld", n, err)
		if err != nil {
			reconnectErr = errors.Join(reconnectErr, fmt.Errorf("reconnect network %s: %w", n, err))
		}
	}
	if reconnectErr != nil {
		return reconnectErr
	}
	if old.State != nil && old.State.Running {
		if original.State == nil {
			return fmt.Errorf("original state unavailable after recovery")
		}
		if !original.State.Running {
			err = restartOriginalContainer(ctx, cli, old.ID)
			logContainerSwitchStep(logger, "ContainerRollbackRestartOld", name, err)
			if err != nil {
				return err
			}
		}
	}
	verified, err := cli.ContainerInspect(ctx, old.ID)
	if err != nil {
		return fmt.Errorf("verify recovery: %w", err)
	}
	if strings.TrimPrefix(verified.Name, "/") != name {
		return fmt.Errorf("original name was not restored")
	}
	if old.State != nil && old.State.Running && (verified.State == nil || !verified.State.Running) {
		return fmt.Errorf("original container is not running after recovery")
	}
	for n := range wanted {
		if verified.NetworkSettings == nil || verified.NetworkSettings.Networks[n] == nil {
			return fmt.Errorf("network %s was not restored", n)
		}
	}
	return nil
}

func switchContainerWithRecovery(ctx context.Context, cli containerSwitchClient, name string,
	oldContainer container.InspectResponse, createNew func() (container.CreateResponse, error),
	tty bool, logger containerSwitchLogger, boundedRecovery bool,
) (cleanupErr error, err error) {
	if oldContainer.ID == "" {
		return nil, fmt.Errorf("original container ID is empty")
	}
	wasRunning := oldContainer.State != nil && oldContainer.State.Running
	if wasRunning && oldContainer.HostConfig != nil && oldContainer.HostConfig.AutoRemove {
		return nil, fmt.Errorf("cannot safely replace container %s with auto-remove enabled", name)
	}

	recoverUpgrade := func(newID string, cause error) error {
		recoveryCtx, cancel := context.WithTimeout(context.Background(), containerUpgradeRecoveryTimeout)
		defer cancel()
		if newID != "" {
			logContainerStartupLogs(recoveryCtx, cli, newID, name, tty, logger)
		}
		recoveryErr := recoverContainerUpgrade(recoveryCtx, cli, name, oldContainer, newID, logger)
		if ctx.Err() != nil {
			recoveryErr = errors.Join(recoveryErr, fmt.Errorf("Docker request timed out or was canceled; daemon-side completion cannot be confirmed: %w", ctx.Err()))
		}
		if recoveryErr != nil {
			message := fmt.Sprintf("Recovery unconfirmed for container %s (original ID %s, backup %s): %v", name, oldContainer.ID, containerSwitchBackupName(oldContainer.ID), recoveryErr)
			if logger != nil {
				logger.Log(message)
			}
			return errors.Join(cause, errors.New(message))
		}
		return cause
	}

	if wasRunning {
		if err := cli.ContainerStop(ctx, oldContainer.ID, container.StopOptions{}); err != nil {
			logContainerSwitchStep(logger, "ContainerStopOld", name, err)
			if boundedRecovery {
				return nil, recoverUpgrade("", fmt.Errorf("stop original container failed: %w", err))
			}
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
		if boundedRecovery {
			return nil, recoverUpgrade("", fmt.Errorf("rename original container failed: %w", err))
		}
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
		if boundedRecovery {
			return nil, recoverUpgrade("", disconnectErr)
		}
		rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, "", disconnectedNetworks, logger)
		return nil, errors.Join(disconnectErr, rollbackErr)
	}

	created, createErr := createNew()
	logContainerSwitchStep(logger, "ContainerCreateReplacement", name, createErr)
	if createErr != nil {
		if boundedRecovery {
			return nil, recoverUpgrade(created.ID, createErr)
		}
		rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, created.ID, disconnectedNetworks, logger)
		return nil, errors.Join(createErr, rollbackErr)
	}
	if err := cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		logContainerSwitchStep(logger, "ContainerStartReplacement", name, err)
		if boundedRecovery {
			return nil, recoverUpgrade(created.ID, fmt.Errorf("start new container failed: %w", err))
		}
		logContainerStartupLogs(ctx, cli, created.ID, name, tty, logger)
		rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, created.ID, disconnectedNetworks, logger)
		return nil, errors.Join(fmt.Errorf("start new container failed: %w", err), rollbackErr)
	}
	logContainerSwitchStep(logger, "ContainerStartReplacement", name, nil)
	if wasRunning {
		if err := waitContainerReady(ctx, cli, created.ID); err != nil {
			logContainerSwitchStep(logger, "ContainerWaitReplacement", name, err)
			if boundedRecovery {
				return nil, recoverUpgrade(created.ID, fmt.Errorf("new container readiness check failed: %w", err))
			}
			logContainerStartupLogs(ctx, cli, created.ID, name, tty, logger)
			rollbackErr := restoreOriginalContainer(ctx, cli, oldContainer.ID, name, wasRunning, created.ID, disconnectedNetworks, logger)
			return nil, errors.Join(fmt.Errorf("new container readiness check failed: %w", err), rollbackErr)
		}
		logContainerSwitchStep(logger, "ContainerWaitReplacement", name, nil)
	}

	if boundedRecovery {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), containerUpgradeRecoveryTimeout)
		defer cancel()
		ctx = cleanupCtx
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
	containerDiagnosticLogTail  = "200"
)

func logContainerStartupLogs(ctx context.Context, cli containerSwitchClient, containerID, name string, tty bool, logger containerSwitchLogger) {
	if logger == nil {
		return
	}
	logger.Log(fmt.Sprintf("========== %s ==========", i18n.GetWithName("ContainerStartupDiagnostic", name)))
	diagnosticCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reader, err := cli.ContainerLogs(diagnosticCtx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       containerDiagnosticLogTail,
	})
	if err != nil {
		logger.Log(i18n.GetWithNameAndErr("ContainerDiagnosticLogsFailed", name, err))
		return
	}
	defer reader.Close()

	var output bytes.Buffer
	if tty {
		_, err = io.Copy(&output, reader)
	} else {
		_, err = stdcopy.StdCopy(&output, &output, reader)
	}
	if err != nil {
		logger.Log(i18n.GetWithNameAndErr("ContainerDiagnosticLogsFailed", name, err))
		return
	}
	logs := strings.TrimSpace(output.String())
	logger.Log(fmt.Sprintf("---------- %s ----------", i18n.GetMsgByKey("ContainerRecentLogs")))
	if logs == "" {
		logger.Log(i18n.GetMsgByKey("ContainerDiagnosticLogsEmpty"))
		return
	}
	logger.Log(logs)
}

func waitContainerReady(ctx context.Context, cli containerInspectClient, containerID string) error {
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

	initialRestartCount := info.RestartCount
	timeout := containerHealthCheckTimeout(info.Config)
	deadline := time.NewTimer(timeout)
	ticker := time.NewTicker(time.Second)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		if info.State.Restarting || info.RestartCount != initialRestartCount {
			return fmt.Errorf("container restart count changed from %d to %d during startup", initialRestartCount, info.RestartCount)
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

func waitContainerStable(ctx context.Context, cli containerInspectClient, containerID string, initial container.InspectResponse) error {
	startedAt := initial.State.StartedAt
	restartCount := initial.RestartCount
	if err := checkContainerStableState(initial, startedAt, restartCount); err != nil {
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
			return checkContainerStableState(info, startedAt, restartCount)
		case <-ticker.C:
			info, err := cli.ContainerInspect(ctx, containerID)
			if err != nil {
				return err
			}
			if err := checkContainerStableState(info, startedAt, restartCount); err != nil {
				return err
			}
		}
	}
}

func checkContainerStableState(info container.InspectResponse, startedAt string, restartCount int) error {
	if err := checkContainerRunningState(info); err != nil {
		return err
	}
	if info.State.Restarting || info.RestartCount != restartCount {
		return fmt.Errorf("container restart count changed from %d to %d during startup", restartCount, info.RestartCount)
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
			if name != "bridge" && endpoint != nil {
				endpoints[name] = endpoint
			}
		}
	}
	for name, endpoint := range extras {
		if name != "bridge" && endpoint != nil {
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
			name:     name,
			endpoint: endpoints[name],
		})
	}
	return disconnected, nil
}

func reconnectOriginalContainerNetworks(ctx context.Context, cli containerSwitchClient, containerID string, attachments []containerNetworkAttachment) error {
	var reconnectErr error
	for _, attachment := range attachments {
		err := cli.NetworkConnect(ctx, attachment.name, containerID, attachment.endpoint)
		if err != nil {
			reconnectErr = errors.Join(reconnectErr, fmt.Errorf("reconnect original container to network %s failed: %w", attachment.name, err))
		}
	}
	return reconnectErr
}

func restoreOriginalContainer(ctx context.Context, cli containerSwitchClient, oldContainerID, originalName string, wasRunning bool, newContainer string, disconnectedNetworks []containerNetworkAttachment, logger containerSwitchLogger) error {
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
	if wasRunning && reconnectErr == nil {
		restartErr := restartOriginalContainer(ctx, cli, oldContainerID)
		logContainerSwitchStep(logger, "ContainerRollbackRestartOld", currentName, restartErr)
		rollbackErr = errors.Join(rollbackErr, restartErr)
	}
	return rollbackErr
}

func createContainerWithNetworks(ctx context.Context, client *client.Client, config *container.Config, hostConf *container.HostConfig, networkConf *network.NetworkingConfig, extraNetworks map[string]*network.EndpointSettings, name string) (container.CreateResponse, error) {
	created, err := client.ContainerCreate(ctx, config, hostConf, networkConf, nil, name)
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
		if err != nil {
			_ = client.ContainerRemove(ctx, created.ID, container.RemoveOptions{Force: true})
			return created, err
		}
	}
	return created, nil
}
