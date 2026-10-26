package mcpruntime

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	ImageRepository = "1panel/1panel-mcp-server"
	ContainerName   = "1panel-mcp"
	previousName    = "1panel-mcp-previous"
	SocketDirectory = "/etc/1panel/mcp-runtime"
	CoreSocket      = SocketDirectory + "/core.sock"
	adapterBinary   = "/usr/local/bin/mcp-1panel"
	ownerLabel      = "io.1panel.mcp.managed"
	imageLabel      = "io.1panel.mcp.image"
	portLabel       = "io.1panel.mcp.port"
	versionLabel    = "io.1panel.mcp.version"
	maxCatalogBytes = 2 << 20
	maxLogBytes     = 64 << 10
	contractVersion = 5
	protocol        = "streamable-http"
)

var semanticVersion = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?$`)
var digestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
var ErrUnsupported = errors.New("MCP Docker runtime requires a local rootful Linux Docker daemon without user namespace remapping")

type Request struct {
	Operation string `json:"operation"`
	Image     string `json:"image"`
	Port      int    `json:"port"`
	Enabled   bool   `json:"enabled"`
	Restart   bool   `json:"restart"`
	ImageID   string `json:"imageID,omitempty"`
}

type State struct {
	Installed      bool   `json:"installed"`
	Status         string `json:"status"`
	Ready          bool   `json:"ready"`
	Reason         string `json:"reason,omitempty"`
	AdapterVersion string `json:"adapterVersion,omitempty"`
	Image          string `json:"image,omitempty"`
	ImageID        string `json:"imageID,omitempty"`
}

type Response struct {
	State   *State          `json:"state,omitempty"`
	Catalog json.RawMessage `json:"catalog,omitempty"`
	Logs    string          `json:"logs,omitempty"`
	Error   string          `json:"error,omitempty"`
	ImageID string          `json:"imageID,omitempty"`
}

type DockerClient interface {
	Close() error
	DaemonHost() string
	Info(context.Context) (system.Info, error)
	ImageInspect(context.Context, string, ...client.ImageInspectOption) (image.InspectResponse, error)
	ImagePull(context.Context, string, image.PullOptions) (io.ReadCloser, error)
	ContainerInspect(context.Context, string) (container.InspectResponse, error)
	ContainerCreate(context.Context, *container.Config, *container.HostConfig, *network.NetworkingConfig, *ocispec.Platform, string) (container.CreateResponse, error)
	ContainerStart(context.Context, string, container.StartOptions) error
	ContainerStop(context.Context, string, container.StopOptions) error
	ContainerRemove(context.Context, string, container.RemoveOptions) error
	ContainerRename(context.Context, string, string) error
	ContainerLogs(context.Context, string, container.LogsOptions) (io.ReadCloser, error)
}

type metadata struct {
	Status          string `json:"status"`
	AdapterVersion  string `json:"adapterVersion"`
	ContractVersion int    `json:"contractVersion"`
	Protocol        string `json:"protocol"`
}

type verifiedImage struct {
	metadata metadata
	catalog  []byte
}

type Controller struct {
	newClient      func() (DockerClient, error)
	platform       string
	health         func(context.Context, int) (metadata, error)
	checkDirectory func() error
	mutation       chan struct{}
	probes         chan struct{}
	cacheMu        sync.Mutex
	cache          map[string]verifiedImage
}

func New() *Controller {
	return newController(func() (DockerClient, error) { return docker.NewDockerClient() })
}

func newController(factory func() (DockerClient, error)) *Controller {
	return &Controller{newClient: factory, platform: runtime.GOOS, health: readHealth,
		checkDirectory: checkSocketDirectory, mutation: make(chan struct{}, 1), probes: make(chan struct{}, 8), cache: make(map[string]verifiedImage)}
}

func ValidateImage(value string) error {
	if len(value) > 180 {
		return errors.New("invalid MCP Docker image")
	}
	if strings.HasPrefix(value, ImageRepository+"@") && digestPattern.MatchString(strings.TrimPrefix(value, ImageRepository+"@")) {
		return nil
	}
	if strings.HasPrefix(value, ImageRepository+":") && semanticVersion.MatchString(strings.TrimPrefix(value, ImageRepository+":")) {
		return nil
	}
	return errors.New("MCP Docker image must use the official repository with a fixed semantic version or SHA256 digest")
}

func (c *Controller) Handle(ctx context.Context, req Request) (Response, error) {
	switch req.Operation {
	case "prepare", "catalog", "candidate_catalog", "status", "apply", "install", "rollback", "rollback_absent", "pull", "logs":
	default:
		return Response{}, errors.New("invalid MCP Docker operation")
	}
	if err := ValidateImage(req.Image); err != nil {
		return Response{}, err
	}
	if req.Port < 1024 || req.Port > 65535 {
		return Response{}, errors.New("MCP listen port must be between 1024 and 65535")
	}
	timeout := 60 * time.Second
	if req.Operation == "catalog" || req.Operation == "candidate_catalog" || req.Operation == "logs" || req.Operation == "prepare" {
		timeout = 5 * time.Second
	}
	if req.Operation == "status" {
		timeout = 8 * time.Second
	}
	if req.Operation == "pull" {
		timeout = 4 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if req.Operation == "apply" || req.Operation == "install" || req.Operation == "rollback" || req.Operation == "rollback_absent" || req.Operation == "pull" {
		select {
		case c.mutation <- struct{}{}:
			defer func() { <-c.mutation }()
		case <-ctx.Done():
			return Response{}, ctx.Err()
		}
	}
	cli, err := c.newClient()
	if err != nil {
		return unavailable(req, "Docker daemon is unavailable")
	}
	defer cli.Close()
	if err = c.supported(ctx, cli); err != nil {
		return unavailable(req, err.Error())
	}
	switch req.Operation {
	case "prepare":
	case "catalog", "candidate_catalog":
		var img image.InspectResponse
		img, err = c.resolve(ctx, cli, req.Image)
		if err == nil && req.Operation == "catalog" {
			var current *container.InspectResponse
			current, err = c.owned(ctx, cli, ContainerName)
			if err == nil && current != nil && isRunning(current) && current.Config.Labels[imageLabel] == req.Image && current.Image != img.ID {
				err = errors.New("MCP image reference changed; restart the adapter before reading its catalog")
			}
		}
		if err == nil {
			var verified verifiedImage
			verified, err = c.verify(ctx, cli, img.ID, req.Image)
			if err == nil {
				return Response{Catalog: append(json.RawMessage(nil), verified.catalog...), ImageID: img.ID}, nil
			}
		}
	case "status":
		return c.status(ctx, cli, req), nil
	case "apply":
		err = c.apply(ctx, cli, req)
	case "install":
		err = c.installRuntime(ctx, cli, req)
	case "rollback":
		err = c.rollback(ctx, cli, req)
	case "rollback_absent":
		err = c.rollbackAbsent(ctx, cli, req)
	case "pull":
		err = c.pullImage(ctx, cli, req)
	case "logs":
		var current *container.InspectResponse
		current, err = c.owned(ctx, cli, ContainerName)
		if err == nil && current == nil {
			err = errors.New("MCP container is not installed")
		}
		if err == nil {
			var data []byte
			data, err = readLogs(ctx, cli, current.ID, maxLogBytes, "100")
			if err == nil {
				return Response{Logs: redactLog(string(data))}, nil
			}
		}
	}
	return Response{}, err
}

func unavailable(req Request, reason string) (Response, error) {
	if req.Operation == "status" {
		return Response{State: &State{Status: "unsupported", Image: req.Image, Reason: reason}}, nil
	}
	return Response{}, errors.New(reason)
}

func (c *Controller) supported(ctx context.Context, cli DockerClient) error {
	if c.platform != "linux" || !strings.HasPrefix(cli.DaemonHost(), "unix:///") {
		return ErrUnsupported
	}
	info, err := cli.Info(ctx)
	if err != nil {
		return errors.New("Docker daemon is unavailable")
	}
	if info.OSType != "linux" {
		return ErrUnsupported
	}
	for _, option := range info.SecurityOptions {
		if strings.Contains(option, "rootless") || strings.Contains(option, "userns") {
			return ErrUnsupported
		}
	}
	return nil
}

func (c *Controller) resolve(ctx context.Context, cli DockerClient, ref string) (image.InspectResponse, error) {
	img, err := cli.ImageInspect(ctx, ref)
	if err != nil {
		return image.InspectResponse{}, errors.New("MCP Docker image is unavailable locally; load it or install the selected version")
	}
	if !digestPattern.MatchString(img.ID) || img.Os != "linux" || img.Config == nil || len(img.Config.Volumes) > 0 {
		return image.InspectResponse{}, errors.New("invalid MCP Docker image identity or platform")
	}
	return img, nil
}

func (c *Controller) owned(ctx context.Context, cli DockerClient, name string) (*container.InspectResponse, error) {
	info, err := cli.ContainerInspect(ctx, name)
	if errdefs.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.New("MCP container state is unavailable")
	}
	if info.ContainerJSONBase == nil || info.Config == nil || info.Config.Labels[ownerLabel] != "true" {
		return nil, errors.New("MCP container name is occupied by an unmanaged container")
	}
	return &info, nil
}

func isRunning(info *container.InspectResponse) bool {
	return info != nil && info.State != nil && info.State.Running
}

func (c *Controller) status(ctx context.Context, cli DockerClient, req Request) Response {
	state := &State{Status: "not_installed", Image: req.Image}
	img, err := c.resolve(ctx, cli, req.Image)
	if err == nil {
		state.ImageID = img.ID
	}
	current, inspectErr := c.owned(ctx, cli, ContainerName)
	if inspectErr != nil {
		state.Status = "unhealthy"
		state.Reason = inspectErr.Error()
		return Response{State: state}
	}
	if current == nil {
		if err != nil {
			state.Reason = err.Error()
		} else {
			state.Reason = "MCP image is loaded; install it to create the managed container"
		}
		return Response{State: state}
	}
	state.Installed = true
	state.ImageID = current.Image
	state.AdapterVersion = current.Config.Labels[versionLabel]
	state.Status = "stopped"
	if !isRunning(current) {
		if req.Enabled {
			state.Reason = "MCP container is not running"
		}
		return Response{State: state}
	}
	state.Status = "running"
	if !req.Enabled {
		state.Reason = "MCP access is disabled"
		return Response{State: state}
	}
	if err != nil || current.Image != img.ID {
		state.Status = "unhealthy"
		state.Reason = "MCP container does not match the configured image or listen port"
		return Response{State: state}
	}
	verified, verifyErr := c.verify(ctx, cli, current.Image, req.Image)
	if verifyErr != nil || !matchesRuntimeWithCapabilities(ctx, cli, current, img.ID, req, verified.metadata.AdapterVersion) {
		state.Status = "unhealthy"
		state.Reason = "MCP container configuration or metadata does not match the verified image"
		return Response{State: state}
	}
	state.AdapterVersion = verified.metadata.AdapterVersion
	health, healthErr := c.health(ctx, req.Port)
	if healthErr != nil || health.Status != "ok" || !validMetadata(health) || health.AdapterVersion != state.AdapterVersion {
		state.Status = "unhealthy"
		state.Reason = "MCP adapter health or contract is invalid"
		return Response{State: state}
	}
	state.Ready = true
	return Response{State: state}
}

func (c *Controller) verify(ctx context.Context, cli DockerClient, id, ref string) (verifiedImage, error) {
	c.cacheMu.Lock()
	cached, ok := c.cache[id]
	c.cacheMu.Unlock()
	if !ok {
		select {
		case c.probes <- struct{}{}:
			defer func() { <-c.probes }()
		default:
			return verifiedImage{}, errors.New("MCP image catalog is busy; retry shortly")
		}
		probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		version, err := c.probe(probeCtx, cli, id, "--version", 4096)
		if err != nil || json.Unmarshal(version, &cached.metadata) != nil || !validMetadata(cached.metadata) {
			return verifiedImage{}, errors.New("MCP image version or contract is incompatible with this panel")
		}
		cached.catalog, err = c.probe(probeCtx, cli, id, "--catalog", maxCatalogBytes)
		var manifest struct {
			ContractVersion int               `json:"contractVersion"`
			Tools           []json.RawMessage `json:"tools"`
		}
		if err != nil || json.Unmarshal(cached.catalog, &manifest) != nil || manifest.ContractVersion != contractVersion || len(manifest.Tools) > 256 {
			return verifiedImage{}, errors.New("MCP image catalog is invalid or incompatible")
		}
		c.cacheMu.Lock()
		if len(c.cache) >= 32 {
			c.cache = make(map[string]verifiedImage)
		}
		c.cache[id] = cached
		c.cacheMu.Unlock()
	}
	if strings.HasPrefix(ref, ImageRepository+":") && strings.TrimPrefix(cached.metadata.AdapterVersion, "v") != strings.TrimPrefix(strings.TrimPrefix(ref, ImageRepository+":"), "v") {
		return verifiedImage{}, errors.New("MCP image version does not match its selected tag")
	}
	return cached, nil
}

func validMetadata(value metadata) bool {
	return semanticVersion.MatchString(value.AdapterVersion) && value.ContractVersion == contractVersion && value.Protocol == protocol
}

func safeConfig(id string, args []string) (*container.Config, *container.HostConfig) {
	pids := int64(128)
	config := &container.Config{Image: id, Entrypoint: []string{adapterBinary}, Cmd: args, User: "0:0", Env: []string{}, Labels: map[string]string{ownerLabel: "true"}, NetworkDisabled: true, Healthcheck: &container.HealthConfig{Test: []string{"NONE"}}}
	host := &container.HostConfig{NetworkMode: "none", ReadonlyRootfs: true, CapDrop: []string{"ALL"}, SecurityOpt: []string{"no-new-privileges:true"}, IpcMode: "private",
		Resources: container.Resources{Memory: 256 << 20, MemorySwap: 256 << 20, NanoCPUs: 1_000_000_000, PidsLimit: &pids},
		LogConfig: container.LogConfig{Type: "json-file", Config: map[string]string{"max-size": "3m", "max-file": "1"}}}
	return config, host
}

func (c *Controller) probe(ctx context.Context, cli DockerClient, id, flag string, limit int) (data []byte, err error) {
	config, host := safeConfig(id, []string{flag})
	created, err := cli.ContainerCreate(ctx, config, host, nil, nil, "")
	if err != nil {
		return nil, errors.New("MCP image probe could not be created")
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cleanupErr := cli.ContainerRemove(cleanupCtx, created.ID, container.RemoveOptions{Force: true})
		if cleanupErr != nil && !errdefs.IsNotFound(cleanupErr) {
			err = errors.Join(err, errors.New("MCP image probe cleanup failed"))
		}
	}()
	if err = cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		return nil, errors.New("MCP image probe failed to start")
	}
	for {
		info, inspectErr := cli.ContainerInspect(ctx, created.ID)
		if inspectErr != nil || info.State == nil {
			return nil, errors.New("MCP image probe state is unavailable")
		}
		if info.State.Status == "exited" {
			if info.State.ExitCode != 0 {
				return nil, errors.New("MCP image probe exited unsuccessfully")
			}
			break
		}
		if info.State.Dead {
			return nil, errors.New("MCP image probe exited unsuccessfully")
		}
		select {
		case <-ctx.Done():
			return nil, errors.New("MCP image probe timed out")
		case <-time.After(25 * time.Millisecond):
		}
	}
	return readLogs(ctx, cli, created.ID, limit, "all")
}

type boundedWriter struct {
	bytes.Buffer
	limit int
}

func (b *boundedWriter) Write(value []byte) (int, error) {
	if b.Len()+len(value) > b.limit {
		return 0, errors.New("MCP output exceeds permitted size")
	}
	return b.Buffer.Write(value)
}

func readLogs(ctx context.Context, cli DockerClient, id string, limit int, tail string) ([]byte, error) {
	reader, err := cli.ContainerLogs(ctx, id, container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: tail})
	if err != nil {
		return nil, errors.New("MCP container logs are unavailable")
	}
	defer reader.Close()
	stdout, stderr := &boundedWriter{limit: limit}, &boundedWriter{limit: limit}
	if err = readDockerStreams(reader, stdout, stderr, 2*limit+64<<10); err != nil {
		return nil, errors.New("MCP output exceeds permitted size or is invalid")
	}
	if tail == "all" {
		return stdout.Bytes(), nil
	}
	combined := append(append([]byte(nil), stdout.Bytes()...), stderr.Bytes()...)
	if len(combined) > limit {
		combined = combined[:limit]
	}
	return combined, nil
}

func readDockerStreams(reader io.Reader, stdout, stderr *boundedWriter, maxRaw int) error {
	total := 0
	for {
		var header [8]byte
		_, err := io.ReadFull(reader, header[:])
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		size := int64(binary.BigEndian.Uint32(header[4:]))
		total += len(header)
		if size > int64(maxRaw-total) || header[1] != 0 || header[2] != 0 || header[3] != 0 {
			return errors.New("invalid Docker log frame")
		}
		var output *boundedWriter
		switch header[0] {
		case 1:
			output = stdout
		case 2:
			output = stderr
		default:
			return errors.New("invalid Docker log stream")
		}
		if size > int64(output.limit-output.Len()) {
			return errors.New("MCP output exceeds permitted size")
		}
		if _, err = io.CopyN(output, reader, size); err != nil {
			return err
		}
		total += int(size)
	}
}

func checkSocketDirectory() error {
	info, err := os.Lstat(SocketDirectory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0022 != 0 {
		return errors.New("MCP runtime socket directory is missing or unsafe")
	}
	return nil
}

func readHealth(ctx context.Context, port int) (metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/healthz", port), nil)
	if err != nil {
		return metadata{}, err
	}
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return metadata{}, errors.New("MCP health endpoint is unavailable")
	}
	defer resp.Body.Close()
	var result metadata
	data, readErr := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if resp.StatusCode != 200 || readErr != nil || len(data) > 4096 || json.Unmarshal(data, &result) != nil || result.Status != "ok" || !validMetadata(result) {
		return metadata{}, errors.New("MCP health response is incompatible")
	}
	return result, nil
}

var bearerPattern = regexp.MustCompile(`(?i)bearer\s+[^\s"']+`)
var sensitivePattern = regexp.MustCompile(`(?i)("?(?:password|token|secret|authorization)"?\s*[:=]\s*)("[^"]*"|[^\s,}]+)`)

func redactLog(value string) string {
	return sensitivePattern.ReplaceAllString(bearerPattern.ReplaceAllString(value, "Bearer [REDACTED]"), "${1}[REDACTED]")
}
