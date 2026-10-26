package mcpruntime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/errdefs"
)

func (c *Controller) pullImage(ctx context.Context, cli DockerClient, req Request) error {
	_, err := cli.ImageInspect(ctx, req.Image)
	if err != nil && !errdefs.IsNotFound(err) {
		return errors.New("MCP Docker image inspection failed")
	}
	if errdefs.IsNotFound(err) {
		pullCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()
		reader, pullErr := cli.ImagePull(pullCtx, req.Image, image.PullOptions{})
		if pullErr != nil {
			return errors.New("MCP Docker image download failed")
		}
		defer reader.Close()
		decoder := json.NewDecoder(io.LimitReader(reader, 16<<20))
		for {
			var message struct {
				Error       string `json:"error"`
				ErrorDetail *struct {
					Message string `json:"message"`
				} `json:"errorDetail"`
			}
			if err = decoder.Decode(&message); err == io.EOF {
				break
			} else if err != nil {
				return errors.New("MCP Docker image download was interrupted")
			}
			if message.Error != "" || message.ErrorDetail != nil {
				return errors.New("MCP Docker image download failed")
			}
		}
	}
	img, err := c.resolve(ctx, cli, req.Image)
	if err != nil {
		return err
	}
	_, err = c.verify(ctx, cli, img.ID, req.Image)
	return err
}

func (c *Controller) installRuntime(ctx context.Context, cli DockerClient, req Request) error {
	before, err := c.owned(ctx, cli, ContainerName)
	if err != nil {
		return err
	}
	wasRunning := isRunning(before)
	activate := req
	activate.Enabled = true
	activate.Restart = true
	if err = c.apply(ctx, cli, activate); err != nil {
		return err
	}
	if req.Enabled {
		return nil
	}
	current, err := c.owned(ctx, cli, ContainerName)
	if err == nil && current == nil {
		err = errors.New("installed MCP container is unavailable")
	}
	if err == nil && current != nil {
		err = stopContainer(ctx, cli, current.ID)
	}
	if err == nil {
		return nil
	}
	rollbackCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	var restoreErr error
	if before != nil {
		port, parseErr := strconv.Atoi(before.Config.Labels[portLabel])
		old := Request{Operation: "rollback", Image: before.Config.Labels[imageLabel], ImageID: before.Image, Port: port, Enabled: wasRunning}
		if parseErr != nil || ValidateImage(old.Image) != nil {
			restoreErr = errors.New("old MCP runtime configuration is invalid")
		}
		if restoreErr == nil {
			restoreErr = c.rollback(rollbackCtx, cli, old)
		}
	} else if current != nil {
		restoreErr = cli.ContainerRemove(rollbackCtx, current.ID, container.RemoveOptions{Force: true})
	}
	return errors.Join(errors.New("installed MCP container could not be left stopped"), restoreErr)
}

func runtimeConfig(id string, req Request, version string) (*container.Config, *container.HostConfig) {
	config, host := safeConfig(id, []string{"--mode=panel", "--listen=127.0.0.1:" + strconv.Itoa(req.Port), "--panel-socket=" + CoreSocket})
	config.NetworkDisabled = false
	config.Labels[imageLabel] = req.Image
	config.Labels[portLabel] = strconv.Itoa(req.Port)
	config.Labels[versionLabel] = version
	host.NetworkMode = "host"
	host.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyUnlessStopped}
	host.Mounts = []mount.Mount{{Type: mount.TypeBind, Source: SocketDirectory, Target: SocketDirectory, ReadOnly: true, BindOptions: &mount.BindOptions{Propagation: mount.PropagationRPrivate}}}
	host.LogConfig.Config = map[string]string{"max-size": "1m", "max-file": "3"}
	return config, host
}

func matchesRuntime(info *container.InspectResponse, id string, req Request, version string) bool {
	if info == nil || info.ContainerJSONBase == nil || info.Config == nil || info.HostConfig == nil || info.Image != id {
		return false
	}
	wantConfig, wantHost := runtimeConfig(id, req, version)
	cfg, host := info.Config, info.HostConfig
	if cfg.Labels[versionLabel] != version || cfg.Labels[imageLabel] != req.Image || cfg.Labels[portLabel] != strconv.Itoa(req.Port) || cfg.User != wantConfig.User || cfg.Tty || !reflect.DeepEqual(cfg.Entrypoint, wantConfig.Entrypoint) || !reflect.DeepEqual(cfg.Cmd, wantConfig.Cmd) {
		return false
	}
	if host.NetworkMode != "host" || host.Privileged || !host.ReadonlyRootfs || !reflect.DeepEqual(host.CapDrop, wantHost.CapDrop) || len(host.CapAdd) != 0 || !reflect.DeepEqual(host.SecurityOpt, wantHost.SecurityOpt) || len(host.Binds) != 0 || len(host.VolumesFrom) != 0 || len(host.Devices) != 0 || len(host.DeviceRequests) != 0 || len(host.Mounts) != 1 {
		return false
	}
	m := host.Mounts[0]
	if m.Type != mount.TypeBind || m.Source != SocketDirectory || m.Target != SocketDirectory || !m.ReadOnly {
		return false
	}
	if host.PidMode != "" || host.UsernsMode != "" || host.UTSMode != "" || (host.IpcMode != "" && host.IpcMode != "private") || host.PublishAllPorts || len(host.PortBindings) != 0 || host.Resources.Memory != wantHost.Memory || host.Resources.MemorySwap != wantHost.MemorySwap || host.Resources.NanoCPUs != wantHost.NanoCPUs || host.Resources.PidsLimit == nil || *host.Resources.PidsLimit != *wantHost.PidsLimit {
		return false
	}
	return host.RestartPolicy.Name == container.RestartPolicyUnlessStopped
}

func matchesRuntimeWithCapabilities(ctx context.Context, cli DockerClient, info *container.InspectResponse, id string, req Request, version string) bool {
	if matchesRuntime(info, id, req, version) {
		return true
	}
	if info == nil || info.ContainerJSONBase == nil || info.HostConfig == nil || info.HostConfig.MemorySwap != -1 {
		return false
	}
	adjusted := *info
	base := *info.ContainerJSONBase
	host := *info.HostConfig
	_, wantHost := runtimeConfig(id, req, version)
	host.MemorySwap = wantHost.MemorySwap
	base.HostConfig = &host
	adjusted.ContainerJSONBase = &base
	if !matchesRuntime(&adjusted, id, req, version) {
		return false
	}
	daemon, err := cli.Info(ctx)
	return err == nil && !daemon.SwapLimit
}

func stopContainer(ctx context.Context, cli DockerClient, id string) error {
	seconds := 10
	if err := cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &seconds}); err != nil && !errdefs.IsNotModified(err) {
		return errors.New("MCP container could not be stopped")
	}
	return nil
}

func (c *Controller) apply(ctx context.Context, cli DockerClient, req Request) error {
	current, err := c.owned(ctx, cli, ContainerName)
	if err != nil {
		return err
	}
	if !req.Enabled {
		if current == nil || !isRunning(current) {
			return nil
		}
		return stopContainer(ctx, cli, current.ID)
	}
	if err = c.checkDirectory(); err != nil {
		return err
	}
	img, err := c.resolve(ctx, cli, req.Image)
	if err != nil {
		return err
	}
	if !digestPattern.MatchString(req.ImageID) || img.ID != req.ImageID {
		return errors.New("MCP image identity changed after catalog validation; retry the operation")
	}
	verified, err := c.verify(ctx, cli, img.ID, req.Image)
	if err != nil {
		return err
	}
	if matchesRuntimeWithCapabilities(ctx, cli, current, img.ID, req, verified.metadata.AdapterVersion) {
		wasRunning := isRunning(current)
		if wasRunning && !req.Restart {
			return c.waitHealthy(ctx, cli, current.ID, req, verified.metadata.AdapterVersion)
		}
		if wasRunning {
			if err = stopContainer(ctx, cli, current.ID); err != nil {
				return err
			}
		}
		if err = cli.ContainerStart(ctx, current.ID, container.StartOptions{}); err == nil {
			err = c.waitHealthy(ctx, cli, current.ID, req, verified.metadata.AdapterVersion)
		}
		if err == nil {
			return nil
		}
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		stopErr := stopContainer(rollbackCtx, cli, current.ID)
		if wasRunning {
			stopErr = errors.Join(stopErr, cli.ContainerStart(rollbackCtx, current.ID, container.StartOptions{}))
		}
		return errors.Join(errors.New("MCP container failed to start or become healthy"), stopErr)
	}
	previous, err := c.owned(ctx, cli, previousName)
	if err != nil {
		return err
	}
	if isRunning(previous) {
		return errors.New("MCP previous container is unexpectedly running; resolve its state before upgrading")
	}
	if matchesRuntimeWithCapabilities(ctx, cli, previous, img.ID, req, verified.metadata.AdapterVersion) {
		return c.restorePrevious(ctx, cli, current, previous, req, verified.metadata.AdapterVersion)
	}
	if previous != nil {
		if err = cli.ContainerRemove(ctx, previous.ID, container.RemoveOptions{}); err != nil {
			return errors.New("MCP previous container could not be removed")
		}
	}
	wasRunning := isRunning(current)
	if current != nil {
		if wasRunning {
			if err = stopContainer(ctx, cli, current.ID); err != nil {
				return err
			}
		}
		if err = cli.ContainerRename(ctx, current.ID, previousName); err != nil {
			if wasRunning {
				rollbackCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				err = errors.Join(err, cli.ContainerStart(rollbackCtx, current.ID, container.StartOptions{}))
			}
			return errors.Join(errors.New("MCP old container could not be retained for rollback"), err)
		}
	}
	config, host := runtimeConfig(img.ID, req, verified.metadata.AdapterVersion)
	created, err := cli.ContainerCreate(ctx, config, host, nil, nil, ContainerName)
	if err == nil {
		err = cli.ContainerStart(ctx, created.ID, container.StartOptions{})
	}
	if err == nil {
		err = c.waitHealthy(ctx, cli, created.ID, req, verified.metadata.AdapterVersion)
	}
	if err == nil {
		return nil
	}
	rollbackCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	var rollbackErr error
	if created.ID != "" {
		rollbackErr = cli.ContainerRemove(rollbackCtx, created.ID, container.RemoveOptions{Force: true})
	}
	if current != nil && rollbackErr == nil {
		rollbackErr = cli.ContainerRename(rollbackCtx, current.ID, ContainerName)
		if rollbackErr == nil && wasRunning {
			rollbackErr = cli.ContainerStart(rollbackCtx, current.ID, container.StartOptions{})
		}
	}
	return errors.Join(errors.New("MCP replacement failed to start or become healthy; previous container restored where possible"), rollbackErr)
}

func (c *Controller) restorePrevious(ctx context.Context, cli DockerClient, current, previous *container.InspectResponse, req Request, version string) error {
	wasRunning := isRunning(current)
	if wasRunning {
		if err := stopContainer(ctx, cli, current.ID); err != nil {
			return err
		}
	}
	temporary := ContainerName + "-rollback-" + strings.TrimPrefix(previous.ID, "sha256:")[:min(12, len(strings.TrimPrefix(previous.ID, "sha256:")))]
	if current != nil {
		if err := cli.ContainerRename(ctx, current.ID, temporary); err != nil {
			rollbackCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if wasRunning {
				err = errors.Join(err, cli.ContainerStart(rollbackCtx, current.ID, container.StartOptions{}))
			}
			return errors.Join(errors.New("MCP current container could not be retained for rollback"), err)
		}
	}
	if err := cli.ContainerRename(ctx, previous.ID, ContainerName); err != nil {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var restoreErr error
		if current != nil {
			restoreErr = cli.ContainerRename(rollbackCtx, current.ID, ContainerName)
			if restoreErr == nil && wasRunning {
				restoreErr = cli.ContainerStart(rollbackCtx, current.ID, container.StartOptions{})
			}
		}
		return errors.Join(errors.New("MCP previous container could not be restored"), restoreErr)
	}
	var err error
	if req.Enabled {
		err = cli.ContainerStart(ctx, previous.ID, container.StartOptions{})
		if err == nil {
			err = c.waitHealthy(ctx, cli, previous.ID, req, version)
		}
	}
	if err == nil {
		if current != nil {
			err = cli.ContainerRename(ctx, current.ID, previousName)
		}
		if err == nil {
			return nil
		}
	}
	rollbackCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	rollbackErr := stopContainer(rollbackCtx, cli, previous.ID)
	rollbackErr = errors.Join(rollbackErr, cli.ContainerRename(rollbackCtx, previous.ID, previousName))
	if current != nil && rollbackErr == nil {
		rollbackErr = cli.ContainerRename(rollbackCtx, current.ID, ContainerName)
		if rollbackErr == nil && wasRunning {
			rollbackErr = cli.ContainerStart(rollbackCtx, current.ID, container.StartOptions{})
		}
	}
	return errors.Join(errors.New("MCP previous container failed its health check"), rollbackErr)
}

func (c *Controller) rollback(ctx context.Context, cli DockerClient, req Request) error {
	if (req.ImageID != "" && !digestPattern.MatchString(req.ImageID)) || (req.Enabled && req.ImageID == "") {
		return errors.New("invalid MCP rollback image identity")
	}
	current, err := c.owned(ctx, cli, ContainerName)
	if err != nil {
		return err
	}
	previous, err := c.owned(ctx, cli, previousName)
	if err != nil {
		return err
	}
	var target *container.InspectResponse
	for _, candidate := range []*container.InspectResponse{current, previous} {
		if candidate != nil && candidate.Image == req.ImageID && candidate.Config.Labels[imageLabel] == req.Image && candidate.Config.Labels[portLabel] == strconv.Itoa(req.Port) {
			target = candidate
			break
		}
	}
	if target == nil {
		if !req.Enabled {
			if isRunning(current) {
				return stopContainer(ctx, cli, current.ID)
			}
			return nil
		}
		return errors.New("the previous MCP runtime instance is unavailable for rollback")
	}
	verified, err := c.verify(ctx, cli, target.Image, req.Image)
	if err != nil {
		return err
	}
	if !matchesRuntimeWithCapabilities(ctx, cli, target, req.ImageID, req, verified.metadata.AdapterVersion) {
		return errors.New("the previous MCP runtime configuration is unsafe or incompatible")
	}
	if req.Enabled {
		if err = c.checkDirectory(); err != nil {
			return err
		}
	}
	if target == previous {
		if isRunning(previous) {
			return errors.New("MCP previous container is unexpectedly running")
		}
		return c.restorePrevious(ctx, cli, current, previous, req, verified.metadata.AdapterVersion)
	}
	if !req.Enabled {
		if isRunning(current) {
			return stopContainer(ctx, cli, current.ID)
		}
		return nil
	}
	if !isRunning(current) {
		if err = cli.ContainerStart(ctx, current.ID, container.StartOptions{}); err != nil {
			return errors.New("MCP rollback instance failed to start")
		}
	}
	return c.waitHealthy(ctx, cli, current.ID, req, verified.metadata.AdapterVersion)
}

// rollbackAbsent removes only the fixed owned instance created by a failed
// first installation. It never resolves a tag, pulls, or creates a runtime.
func (c *Controller) rollbackAbsent(ctx context.Context, cli DockerClient, req Request) error {
	if req.Enabled || !digestPattern.MatchString(req.ImageID) {
		return errors.New("invalid MCP absent-runtime rollback identity")
	}
	current, err := c.owned(ctx, cli, ContainerName)
	if err != nil || current == nil {
		return err
	}
	if current.Image != req.ImageID || current.Config.Labels[imageLabel] != req.Image || current.Config.Labels[portLabel] != strconv.Itoa(req.Port) {
		return errors.New("MCP runtime changed after installation; absent-runtime rollback was rejected")
	}
	if _, err = c.resolve(ctx, cli, current.Image); err != nil {
		return err
	}
	verified, err := c.verify(ctx, cli, current.Image, req.Image)
	if err != nil {
		return err
	}
	if !matchesRuntimeWithCapabilities(ctx, cli, current, req.ImageID, req, verified.metadata.AdapterVersion) {
		return errors.New("MCP installed runtime configuration is unsafe or incompatible")
	}
	if err = cli.ContainerRemove(ctx, current.ID, container.RemoveOptions{Force: true}); err != nil && !errdefs.IsNotFound(err) {
		return errors.New("MCP failed installation could not be removed")
	}
	return nil
}

func (c *Controller) waitHealthy(ctx context.Context, cli DockerClient, id string, req Request, version string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	for {
		if ctx.Err() != nil {
			return errors.New("MCP container did not become healthy before the timeout")
		}
		info, err := cli.ContainerInspect(ctx, id)
		if err == nil && isRunning(&info) {
			if info.Config == nil || info.Config.Labels[ownerLabel] != "true" || !matchesRuntimeWithCapabilities(ctx, cli, &info, req.ImageID, req, version) {
				return errors.New("MCP container configuration or metadata does not match the verified image")
			}
			health, healthErr := c.health(ctx, req.Port)
			if healthErr == nil && ctx.Err() == nil && validMetadata(health) && health.Status == "ok" && health.AdapterVersion == version {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return errors.New("MCP container did not become healthy before the timeout")
		case <-time.After(250 * time.Millisecond):
		}
	}
}
