package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"

	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
)

const dockerNftablesMinVersion = "29.0.0"

type DockerService struct{}

type IDockerService interface {
	UpdateConf(req dto.SettingUpdate, withRestart bool) error
	UpdateFirewallBackend(backend string) error
	UpdateLogOption(req dto.LogOption) error
	UpdateIpv6Option(req dto.Ipv6Option) error
	UpdateConfByFile(info dto.DaemonJsonUpdateByFile) error
	LoadDockerStatus() *dto.DockerStatus
	LoadDockerConf() (*dto.DaemonJsonConf, error)
	OperateDocker(req dto.DockerOperation) error
}

func loadDockerEngineVersion(ctx context.Context) string {
	client, err := docker.NewDockerClient()
	if err == nil {
		defer client.Close()
		if version, versionErr := client.ServerVersion(ctx); versionErr == nil && version.Version != "" {
			return version.Version
		}
	}
	if !cmd.Which("dockerd") {
		return ""
	}
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("dockerd", "--version")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(stdout)
}

func dockerNftablesSupported(version string) bool {
	return version != "" && common.CompareAppVersion(version, dockerNftablesMinVersion)
}

func applyDockerFirewallBackendConfig(daemonMap map[string]interface{}, backend, version string) error {
	switch backend {
	case constant.FirewallProviderNftables:
		if !dockerNftablesSupported(version) {
			return fmt.Errorf("Docker Engine %s or later is required for the nftables firewall backend", dockerNftablesMinVersion)
		}
		daemonMap["experimental"] = true
		daemonMap["firewall-backend"] = constant.FirewallProviderNftables
	case constant.FirewallProviderIptables:
		if dockerNftablesSupported(version) {
			daemonMap["firewall-backend"] = constant.FirewallProviderIptables
		} else {
			delete(daemonMap, "firewall-backend")
		}
	default:
		return fmt.Errorf("unsupported Docker firewall backend %q", backend)
	}
	return nil
}

func (u *DockerService) UpdateFirewallBackend(backend string) error {
	version := loadDockerEngineVersion(context.Background())
	if backend == constant.FirewallProviderNftables && !dockerNftablesSupported(version) {
		return fmt.Errorf("Docker Engine %s or later is required for the nftables firewall backend", dockerNftablesMinVersion)
	}
	if backend == constant.FirewallProviderNftables {
		if err := dockerfirewall.CheckIPv4Forwarding(); err != nil {
			return err
		}
	}

	original, readErr := os.ReadFile(constant.DaemonJsonPath)
	existed := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}
	daemonMap := make(map[string]interface{})
	if len(bytes.TrimSpace(original)) > 0 {
		if err := json.Unmarshal(original, &daemonMap); err != nil {
			return fmt.Errorf("failed to parse Docker configuration: %w", err)
		}
	}
	if err := applyDockerFirewallBackendConfig(daemonMap, backend, version); err != nil {
		return err
	}
	updated, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if existed && bytes.Equal(bytes.TrimSpace(original), bytes.TrimSpace(updated)) {
		return nil
	}
	if err := os.MkdirAll(path.Dir(constant.DaemonJsonPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, updated, 0640); err != nil {
		return err
	}
	restore := func() error {
		if existed {
			return os.WriteFile(constant.DaemonJsonPath, original, 0640)
		}
		err := os.Remove(constant.DaemonJsonPath)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := validateDockerConfig(); err != nil {
		if restoreErr := restore(); restoreErr != nil {
			return fmt.Errorf("%v; failed to restore Docker configuration: %w", err, restoreErr)
		}
		return err
	}
	if err := controller.HandleRestart("docker"); err != nil {
		cause := fmt.Errorf("failed to restart Docker: %w", err)
		if restoreErr := restore(); restoreErr != nil {
			return fmt.Errorf("%v; failed to restore Docker configuration: %w", cause, restoreErr)
		}
		if restoreRestartErr := controller.HandleRestart("docker"); restoreRestartErr != nil {
			return fmt.Errorf("%v; the previous configuration was restored but Docker could not be restarted: %w", cause, restoreRestartErr)
		}
		return cause
	}
	return nil
}

func NewIDockerService() IDockerService {
	return &DockerService{}
}

type daemonJsonItem struct {
	Status       string    `json:"status"`
	Mirrors      []string  `json:"registry-mirrors"`
	Registries   []string  `json:"insecure-registries"`
	LiveRestore  bool      `json:"live-restore"`
	Ipv6         bool      `json:"ipv6"`
	FixedCidrV6  string    `json:"fixed-cidr-v6"`
	Ip6Tables    bool      `json:"ip6tables"`
	Experimental bool      `json:"experimental"`
	IPTables     bool      `json:"iptables"`
	ExecOpts     []string  `json:"exec-opts"`
	LogOption    logOption `json:"log-opts"`
}
type logOption struct {
	LogMaxSize string `json:"max-size"`
	LogMaxFile string `json:"max-file"`
}

func (u *DockerService) LoadDockerStatus() *dto.DockerStatus {
	ctx := context.Background()
	var data dto.DockerStatus
	if !cmd.Which("docker") {
		data.IsExist = false
		return &data
	}
	data.IsExist = true
	data.IsActive = true
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("load docker client failed, err: %v", err)
		data.IsActive = false
		return &data
	}
	defer client.Close()
	if _, err := client.Ping(ctx); err != nil {
		global.LOG.Errorf("ping docker client failed, err: %v", err)
		data.IsActive = false
	}

	return &data
}

func readDockerDaemonConfig() (map[string]interface{}, error) {
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return nil, err
	}
	daemonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &daemonMap); err != nil {
		return nil, err
	}
	return daemonMap, nil
}

func (u *DockerService) LoadDockerConf() (*dto.DaemonJsonConf, error) {
	ctx := context.Background()
	var data dto.DaemonJsonConf
	data.IPTables = true
	data.Version = "-"
	client, err := docker.NewDockerClient()
	if err != nil {
		return &data, err
	}
	defer client.Close()
	data.NVIDIASupported = client.DaemonHost() == "unix:///var/run/docker.sock"
	itemVersion, err := client.ServerVersion(ctx)
	if err == nil {
		data.Version = itemVersion.Version
	}
	data.IsSwarm = false
	stdout2, _ := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("docker", "info")
	if strings.Contains(stdout2, "Swarm: active") {
		data.IsSwarm = true
	}
	daemonMap, err := readDockerDaemonConfig()
	if err != nil {
		return &data, nil
	}
	if runtimes, ok := daemonMap["runtimes"].(map[string]interface{}); ok {
		_, data.NVIDIA = runtimes["nvidia"]
	}
	var conf daemonJsonItem
	arr, err := json.Marshal(daemonMap)
	if err != nil {
		return &data, err
	}
	if err := json.Unmarshal(arr, &conf); err != nil {
		return &data, err
	}
	if _, ok := daemonMap["iptables"]; !ok {
		conf.IPTables = true
	}
	data.CgroupDriver = "cgroupfs"
	for _, opt := range conf.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			data.CgroupDriver = strings.ReplaceAll(opt, "native.cgroupdriver=", "")
			break
		}
	}
	data.Ipv6 = conf.Ipv6
	data.FixedCidrV6 = conf.FixedCidrV6
	data.Ip6Tables = conf.Ip6Tables
	data.Experimental = conf.Experimental
	data.LogMaxSize = conf.LogOption.LogMaxSize
	data.LogMaxFile = conf.LogOption.LogMaxFile
	data.Mirrors = conf.Mirrors
	data.Registries = conf.Registries
	data.IPTables = conf.IPTables
	data.LiveRestore = conf.LiveRestore
	return &data, nil
}

func (u *DockerService) UpdateConf(req dto.SettingUpdate, withRestart bool) error {
	if req.Key == "NVIDIA" {
		return updateDockerNVIDIA(req.Value, withRestart)
	}
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)
	switch req.Key {
	case "Registries":
		req.Value = strings.TrimSuffix(req.Value, ",")
		if len(req.Value) == 0 {
			delete(daemonMap, "insecure-registries")
		} else {
			daemonMap["insecure-registries"] = strings.Split(req.Value, ",")
		}
	case "Mirrors":
		req.Value = strings.TrimSuffix(req.Value, ",")
		if len(req.Value) == 0 {
			delete(daemonMap, "registry-mirrors")
		} else {
			daemonMap["registry-mirrors"] = strings.Split(req.Value, ",")
		}
	case "Ipv6":
		if req.Value == "disable" {
			delete(daemonMap, "ipv6")
			delete(daemonMap, "fixed-cidr-v6")
			delete(daemonMap, "ip6tables")
			backend, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
			if !strings.EqualFold(strings.TrimSpace(backend), constant.FirewallProviderNftables) {
				delete(daemonMap, "experimental")
			}
		}
	case "LogOption":
		if req.Value == "disable" {
			delete(daemonMap, "log-opts")
		}
	case "LiveRestore":
		if req.Value == "disable" {
			delete(daemonMap, "live-restore")
		} else {
			daemonMap["live-restore"] = true
		}
	case "IPtables":
		if req.Value == "enable" {
			delete(daemonMap, "iptables")
		} else {
			daemonMap["iptables"] = false
		}
	case "Driver":
		if opts, ok := daemonMap["exec-opts"]; ok {
			if optsValue, isArray := opts.([]interface{}); isArray {
				for i := 0; i < len(optsValue); i++ {
					if opt, isStr := optsValue[i].(string); isStr {
						if strings.HasPrefix(opt, "native.cgroupdriver=") {
							optsValue[i] = "native.cgroupdriver=" + req.Value
							break
						}
					}
				}
			}
		} else {
			if req.Value == "systemd" {
				daemonMap["exec-opts"] = []string{"native.cgroupdriver=systemd"}
			}
		}
	case "http-proxy", "https-proxy":
		delete(daemonMap, "proxies")
		if len(req.Value) > 0 {
			proxies := map[string]interface{}{
				req.Key: req.Value,
			}
			daemonMap["proxies"] = proxies
		}
	case "socks5-proxy", "close-proxy":
		delete(daemonMap, "proxies")
		if len(req.Value) > 0 {
			proxies := map[string]interface{}{
				"http-proxy":  req.Value,
				"https-proxy": req.Value,
			}
			daemonMap["proxies"] = proxies
		}
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if string(newJson) == string(file) {
		return nil
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}
	if err := validateDockerConfig(); err != nil {
		return err
	}

	if withRestart {
		if err := controller.HandleRestart("docker"); err != nil {
			return fmt.Errorf("failed to restart Docker: %v", err)
		}
	}
	return nil
}
func createIfNotExistDaemonJsonFile() error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		var daemonFile *os.File
		daemonFile, err = os.Create(constant.DaemonJsonPath)
		if err != nil {
			return err
		}
		defer daemonFile.Close()
	}
	return nil
}

func (u *DockerService) UpdateLogOption(req dto.LogOption) error {
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	changeLogOption(daemonMap, req.LogMaxFile, req.LogMaxSize)
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	if err := validateDockerConfig(); err != nil {
		return err
	}

	if err := controller.HandleRestart("docker"); err != nil {
		return fmt.Errorf("failed to restart Docker: %v", err)
	}
	return nil
}

func (u *DockerService) UpdateIpv6Option(req dto.Ipv6Option) error {
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}

	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	daemonMap["ipv6"] = true
	daemonMap["fixed-cidr-v6"] = req.FixedCidrV6
	if req.Ip6Tables {
		daemonMap["ip6tables"] = req.Ip6Tables
	}
	if req.Experimental {
		daemonMap["experimental"] = req.Experimental
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	if err := validateDockerConfig(); err != nil {
		return err
	}

	if err := controller.HandleRestart("docker"); err != nil {
		return fmt.Errorf("failed to restart Docker: %v", err)
	}
	return nil
}

func (u *DockerService) UpdateConfByFile(req dto.DaemonJsonUpdateByFile) error {
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(constant.DaemonJsonPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	if err := validateDockerConfig(); err != nil {
		return err
	}

	if err := controller.HandleRestart("docker"); err != nil {
		return fmt.Errorf("failed to restart Docker: %v", err)
	}
	return nil
}

func (u *DockerService) OperateDocker(req dto.DockerOperation) error {
	service := "docker"
	if req.Operation == "stop" {
		isSocketActive, _ := controller.CheckExist("docker.socket")
		if isSocketActive {
			if err := controller.HandleStop("docker.socket"); err != nil {
				global.LOG.Errorf("handle stop docker.socket failed, err: %v", err)
			}
		}
	}
	if req.Operation == "restart" {
		if err := validateDockerConfig(); err != nil {
			return err
		}
	}
	if err := controller.Handle(req.Operation, service); err != nil {
		return err
	}
	if req.Operation == "start" || req.Operation == "restart" {
		return RestoreDockerPortGuard(context.Background())
	}
	return nil
}

func changeLogOption(daemonMap map[string]interface{}, logMaxFile, logMaxSize string) {
	if opts, ok := daemonMap["log-opts"]; ok {
		if len(logMaxFile) != 0 || len(logMaxSize) != 0 {
			daemonMap["log-driver"] = "json-file"
		}
		optsMap, isMap := opts.(map[string]interface{})
		if isMap {
			if len(logMaxFile) != 0 {
				optsMap["max-file"] = logMaxFile
			} else {
				delete(optsMap, "max-file")
			}
			if len(logMaxSize) != 0 {
				optsMap["max-size"] = logMaxSize
			} else {
				delete(optsMap, "max-size")
			}
			if len(optsMap) == 0 {
				delete(daemonMap, "log-opts")
			}
		} else {
			optsMap := make(map[string]interface{})
			if len(logMaxFile) != 0 {
				optsMap["max-file"] = logMaxFile
			}
			if len(logMaxSize) != 0 {
				optsMap["max-size"] = logMaxSize
			}
			if len(optsMap) != 0 {
				daemonMap["log-opts"] = optsMap
			}
		}
	} else {
		if len(logMaxFile) != 0 || len(logMaxSize) != 0 {
			daemonMap["log-driver"] = "json-file"
		}
		optsMap := make(map[string]interface{})
		if len(logMaxFile) != 0 {
			optsMap["max-file"] = logMaxFile
		}
		if len(logMaxSize) != 0 {
			optsMap["max-size"] = logMaxSize
		}
		if len(optsMap) != 0 {
			daemonMap["log-opts"] = optsMap
		}
	}
}

func validateDockerConfig() error {
	if !cmd.Which("dockerd") {
		return nil
	}
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("dockerd", "--validate")
	if strings.Contains(stdout, "unknown flag: --validate") {
		return nil
	}
	if err != nil || (stdout != "" && strings.TrimSpace(stdout) != "configuration OK") {
		return fmt.Errorf("Docker configuration validation failed, %v", err)
	}
	return nil
}

func updateDockerNVIDIA(value string, withRestart bool) error {
	if value != "enable" && value != "disable" {
		return fmt.Errorf("unsupported NVIDIA setting %q", value)
	}
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	if client.DaemonHost() != "unix:///var/run/docker.sock" {
		return fmt.Errorf("NVIDIA configuration is only supported for the local default Docker socket")
	}
	enabled := value == "enable"
	original, readErr := os.ReadFile(constant.DaemonJsonPath)
	existed := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}
	daemonMap := make(map[string]interface{})
	if existed {
		if err := json.Unmarshal(original, &daemonMap); err != nil {
			return fmt.Errorf("failed to parse Docker configuration: %w", err)
		}
		if daemonMap == nil {
			return fmt.Errorf("Docker configuration must be a JSON object")
		}
	}
	runtimePath, err := applyDockerNVIDIAConfig(daemonMap, enabled)
	if err != nil {
		return err
	}
	if enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		check := checkNVIDIAToolkit(ctx, func(name string) string {
			if name == "nvidia-container-runtime" {
				return runtimePath
			}
			return findGPUContainerCommand(name)
		}, func(ctx context.Context, executable string) (string, error) {
			return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(3*time.Second)).RunWithStdout(executable, "--version")
		})
		cancel()
		if check.Status != "passed" {
			return fmt.Errorf("NVIDIA Container Toolkit is not ready: %s", check.Reason)
		}
	}
	updated, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path.Dir(constant.DaemonJsonPath), 0755); err != nil {
		return err
	}
	if err := writeDockerNVIDIAConfig(updated); err != nil {
		return err
	}
	restore := func(cause error, restart bool) error {
		var restoreErr error
		if existed {
			restoreErr = writeDockerNVIDIAConfig(original)
		} else {
			restoreErr = os.Remove(constant.DaemonJsonPath)
		}
		if restoreErr != nil {
			return fmt.Errorf("%v; failed to restore Docker configuration: %w", cause, restoreErr)
		}
		if restart {
			if err := controller.HandleRestart("docker"); err != nil {
				return fmt.Errorf("%v; configuration restored but Docker restart failed: %w", cause, err)
			}
		}
		return cause
	}
	if global.CACHE != nil {
		defer global.CACHE.Del(gpuEnvironmentCacheKey)
	}
	if err := validateDockerConfig(); err != nil {
		return restore(err, false)
	}
	if !withRestart {
		return nil
	}
	if err := controller.HandleRestart("docker"); err != nil {
		return restore(fmt.Errorf("failed to restart Docker: %w", err), true)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	info, err := client.Info(ctx)
	if err != nil {
		return restore(fmt.Errorf("failed to verify Docker NVIDIA runtime: %w", err), true)
	}
	if enabled {
		check := checkNVIDIADockerConfig(info, readDockerDaemonConfig, findGPUContainerCommand)
		if check.Status != "passed" {
			return restore(fmt.Errorf("Docker NVIDIA runtime did not become available: %s", check.Reason), true)
		}
	} else if _, registered := info.Runtimes["nvidia"]; registered {
		return restore(fmt.Errorf("Docker still registers the NVIDIA runtime; check other Docker configuration sources"), true)
	}
	return nil
}

func applyDockerNVIDIAConfig(config map[string]interface{}, enabled bool) (string, error) {
	runtimes, ok := config["runtimes"].(map[string]interface{})
	if config["runtimes"] != nil && !ok {
		return "", fmt.Errorf("Docker runtimes configuration must be a JSON object")
	}
	if !enabled {
		delete(runtimes, "nvidia")
		if len(runtimes) == 0 {
			delete(config, "runtimes")
		}
		if config["default-runtime"] == "nvidia" {
			delete(config, "default-runtime")
		}
		return "", nil
	}
	if runtimes == nil {
		runtimes = make(map[string]interface{})
		config["runtimes"] = runtimes
	}
	nvidia, ok := runtimes["nvidia"].(map[string]interface{})
	if runtimes["nvidia"] != nil && !ok {
		return "", fmt.Errorf("Docker NVIDIA runtime configuration must be a JSON object")
	}
	if nvidia == nil {
		nvidia = map[string]interface{}{"runtimeArgs": []string{}}
		runtimes["nvidia"] = nvidia
	}
	configured, ok := nvidia["path"].(string)
	if nvidia["path"] != nil && !ok {
		return "", fmt.Errorf("Docker NVIDIA runtime path must be a string")
	}
	runtimePath := findNVIDIAContainerCommand("nvidia-container-runtime", config)
	if runtimePath == "" || !cmd.Which(runtimePath) {
		if configured != "" {
			return "", fmt.Errorf("configured NVIDIA runtime %q is not executable", configured)
		}
		return "", fmt.Errorf("nvidia-container-runtime is not installed or executable")
	}
	if configured == "" {
		nvidia["path"] = runtimePath
	}
	return runtimePath, nil
}

func writeDockerNVIDIAConfig(data []byte) error {
	target := constant.DaemonJsonPath
	mode := os.FileMode(0640)
	info, err := os.Lstat(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		target, err = filepath.EvalSymlinks(target)
		if err != nil {
			return err
		}
		info, err = os.Stat(target)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("Docker configuration must be a regular file")
		}
		mode = info.Mode().Perm()
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".daemon-nvidia-*.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = temp.Close()
		_ = os.Remove(temp.Name())
	}()
	if info != nil {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			if err := temp.Chown(int(stat.Uid), int(stat.Gid)); err != nil {
				return err
			}
		}
	}
	if err := temp.Chmod(mode); err != nil {
		return err
	}
	if _, err := temp.Write(data); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), target)
}
