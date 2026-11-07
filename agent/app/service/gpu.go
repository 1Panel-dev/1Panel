package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/accelerator"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/npu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/system"
)

const gpuEnvironmentCacheKey = "gpu:environment"

var gpuEnvironmentLock = make(chan struct{}, 1)

var gpuVendors = []string{"nvidia", "intel", "ascend", "amd"}

type gpuEnvironmentProbe struct {
	CheckedAt     time.Time                 `json:"checkedAt"`
	Hardware      map[string]int            `json:"hardware"`
	HardwareError string                    `json:"hardwareError,omitempty"`
	Commands      map[string]bool           `json:"commands"`
	NVIDIAChecked bool                      `json:"nvidiaChecked"`
	Checks        []dto.GPUEnvironmentCheck `json:"checks"`
}

func LoadGPUInfo(ctx context.Context, refresh bool) (*dto.GPURealtimeInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	snapshot := &accelerator.Snapshot{}
	var collectErr error
	if available, client := accelerator.New(); available {
		collected, err := client.Collect(ctx)
		collectErr = err
		if collected != nil {
			snapshot = collected
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if collectErr != nil {
		snapshot.Info.Warnings = append(snapshot.Info.Warnings, collectErr.Error())
	}
	if snapshot.Info.CollectedAt.IsZero() {
		snapshot.Info.CollectedAt = time.Now()
	}
	if global.LOG != nil {
		for _, warning := range snapshot.Info.Warnings {
			global.LOG.Warnf("load realtime accelerator data: %s", warning)
		}
	}
	summaries := summarizeGPUs(snapshot)
	nvidiaPresent := slices.ContainsFunc(summaries, func(item dto.GPUSummary) bool { return item.Vendor == "nvidia" })
	probe := loadGPUEnvironmentProbe(ctx, refresh, nvidiaPresent)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	environment := assembleGPUEnvironment(probe, summaries, collectErr, snapshot.Info)
	for _, check := range environment.Checks {
		if check.Item == "driver" && check.Vendor != "" && !slices.ContainsFunc(summaries, func(item dto.GPUSummary) bool { return item.Vendor == check.Vendor }) {
			summaries = append(summaries, dto.GPUSummary{Vendor: check.Vendor, MemoryPartial: true})
		}
	}
	return &dto.GPURealtimeInfo{Info: snapshot.Info, Environment: environment, Summaries: summaries}, nil
}

func loadGPUEnvironmentProbe(ctx context.Context, refresh, nvidiaPresent bool) gpuEnvironmentProbe {
	select {
	case gpuEnvironmentLock <- struct{}{}:
		defer func() { <-gpuEnvironmentLock }()
	case <-ctx.Done():
		return gpuEnvironmentProbe{CheckedAt: time.Now(), HardwareError: ctx.Err().Error()}
	}
	if !refresh && global.CACHE != nil {
		var cached gpuEnvironmentProbe
		if value := global.CACHE.Get(gpuEnvironmentCacheKey); value != "" && json.Unmarshal([]byte(value), &cached) == nil && (!nvidiaPresent || cached.NVIDIAChecked) {
			return cached
		}
	}
	probe := gpuEnvironmentProbe{CheckedAt: time.Now(), Commands: map[string]bool{}}
	var err error
	probe.Hardware, err = detectGPUHardware("/sys/bus/pci/devices")
	if err != nil {
		probe.HardwareError = err.Error()
	}
	for vendor := range gpu.ManagementCommands() {
		probe.Commands[vendor] = true
	}
	probe.Commands["intel"], _ = xpu.New()
	probe.Commands["ascend"], _ = npu.New()
	if nvidiaPresent || probe.Hardware["nvidia"] > 0 {
		probe.NVIDIAChecked = true
		probe.Checks = checkNVIDIAContainerEnvironment(ctx)
	}
	if ctx.Err() == nil && global.CACHE != nil {
		if data, err := json.Marshal(probe); err == nil {
			global.CACHE.SetWithTTL(gpuEnvironmentCacheKey, string(data), time.Minute)
		}
	}
	return probe
}

func detectGPUHardware(directory string) (map[string]int, error) {
	devices := map[string]int{}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return devices, err
	}
	var readErrors []string
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		vendorData, err := os.ReadFile(filepath.Join(path, "vendor"))
		if err != nil {
			readErrors = append(readErrors, err.Error())
			continue
		}
		vendor := map[string]string{"0x10de": "nvidia", "0x8086": "intel", "0x19e5": "ascend", "0x1002": "amd"}[strings.ToLower(strings.TrimSpace(string(vendorData)))]
		if vendor == "" {
			continue
		}
		classData, err := os.ReadFile(filepath.Join(path, "class"))
		if err != nil {
			readErrors = append(readErrors, err.Error())
			continue
		}
		class, err := strconv.ParseUint(strings.TrimSpace(string(classData)), 0, 32)
		if err != nil {
			readErrors = append(readErrors, fmt.Sprintf("read PCI class for %s: %v", entry.Name(), err))
			continue
		}
		baseClass := class >> 16
		if (vendor == "ascend" && baseClass == 0x12) || (vendor != "ascend" && (baseClass == 0x03 || (vendor == "nvidia" && baseClass == 0x12))) {
			devices[vendor]++
		}
	}
	if len(readErrors) > 0 {
		return devices, fmt.Errorf("PCI inventory incomplete: %s", strings.Join(readErrors, "; "))
	}
	return devices, nil
}

func assembleGPUEnvironment(probe gpuEnvironmentProbe, summaries []dto.GPUSummary, collectErr error, info accelerator.Info) dto.GPUEnvironment {
	result := dto.GPUEnvironment{CheckedAt: probe.CheckedAt, HardwareStatus: "not_detected", Checks: []dto.GPUEnvironmentCheck{}}
	if len(probe.Hardware) > 0 || len(summaries) > 0 {
		result.HardwareStatus = "detected"
	} else if probe.HardwareError != "" {
		result.HardwareStatus = "unknown"
	}
	for _, vendor := range gpuVendors {
		index := slices.IndexFunc(summaries, func(item dto.GPUSummary) bool { return item.Vendor == vendor })
		if index < 0 && probe.Hardware[vendor] == 0 && !(probe.HardwareError != "" && probe.Commands[vendor] && collectErr != nil) {
			continue
		}
		check := dto.GPUEnvironmentCheck{Vendor: vendor, Item: "driver", Status: "unknown"}
		switch {
		case index >= 0:
			check.Status = "passed"
			check.Version = summaries[index].DriverVersion
			if vendor == "intel" {
				var readErrors []string
				for _, device := range info.XPUs {
					if device.ReadError != "" {
						readErrors = append(readErrors, device.ReadError)
					}
				}
				if len(readErrors) > 0 {
					check.Status = "unknown"
					if len(readErrors) == len(info.XPUs) {
						check.Status = "failed"
					}
					check.Reason = strings.Join(readErrors, "; ")
				}
			}
		case !probe.Commands[vendor]:
			check.Status = "missing"
			check.Reason = "The device management utility was not found; check the driver and management tool installation."
		case collectErr != nil:
			check.Status = "failed"
			check.Reason = collectErr.Error()
		default:
			check.Reason = "The management utility returned no readable devices."
			if len(info.Warnings) > 0 {
				check.Status = "failed"
				check.Reason = strings.Join(info.Warnings, "; ")
			}
		}
		result.Checks = append(result.Checks, check)
	}
	result.Checks = append(result.Checks, probe.Checks...)
	if probe.HardwareError != "" {
		result.Checks = append(result.Checks, dto.GPUEnvironmentCheck{Item: "driver", Status: "unknown", Reason: probe.HardwareError})
	}
	return result
}

func findGPUContainerCommand(name string) string {
	for _, path := range []string{name, filepath.Join("/usr/bin", name), filepath.Join("/usr/sbin", name), filepath.Join("/usr/local/bin", name), filepath.Join("/usr/local/sbin", name), filepath.Join("/usr/local/nvidia/toolkit", name)} {
		if cmd.Which(path) {
			return path
		}
		if filepath.IsAbs(path) {
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				return path
			}
		}
	}
	return ""
}

func findNVIDIAContainerCommand(name string, config map[string]interface{}) string {
	if name == "nvidia-container-runtime" {
		runtimes, _ := config["runtimes"].(map[string]interface{})
		nvidia, _ := runtimes["nvidia"].(map[string]interface{})
		if configured, ok := nvidia["path"].(string); ok && configured != "" {
			name = configured
		}
	}
	return findGPUContainerCommand(name)
}

func checkNVIDIAContainerEnvironment(ctx context.Context) []dto.GPUEnvironmentCheck {
	config, configErr := readDockerDaemonConfig()
	toolkitCtx, cancelToolkit := context.WithTimeout(ctx, 5*time.Second)
	toolkit := checkNVIDIAToolkit(toolkitCtx, func(name string) string {
		return findNVIDIAContainerCommand(name, config)
	}, func(ctx context.Context, path string) (string, error) {
		return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(3*time.Second)).RunWithStdout(path, "--version")
	})
	cancelToolkit()
	dockerCtx, cancelDocker := context.WithTimeout(ctx, 5*time.Second)
	defer cancelDocker()
	dockerCheck := dto.GPUEnvironmentCheck{Vendor: "nvidia", Item: "docker", Status: "failed"}
	cli, err := docker.NewDockerClient()
	if err == nil {
		defer cli.Close()
		var info system.Info
		info, err = cli.Info(dockerCtx)
		if err == nil {
			dockerCheck = checkNVIDIADockerInfo(info, false)
			if cli.DaemonHost() == "unix:///var/run/docker.sock" {
				dockerCheck = checkNVIDIADockerConfig(info, func() (map[string]interface{}, error) { return config, configErr }, findGPUContainerCommand)
			} else {
				dockerCheck.Status = "unknown"
				dockerCheck.Reason = "The selected Docker daemon's NVIDIA runtime cannot be verified using local executables or the default daemon.json."
			}
		}
	}
	if err != nil {
		dockerCheck.Reason = err.Error()
	}
	return []dto.GPUEnvironmentCheck{toolkit, dockerCheck}
}

func checkNVIDIAToolkit(ctx context.Context, findCommand func(string) string, runVersion func(context.Context, string) (string, error)) dto.GPUEnvironmentCheck {
	toolkit := dto.GPUEnvironmentCheck{Vendor: "nvidia", Item: "toolkit", Status: "missing"}
	cli := findCommand("nvidia-container-cli")
	if cli == "" {
		toolkit.Reason = "The required nvidia-container-cli component was not found."
		return toolkit
	}
	runtime := findCommand("nvidia-container-runtime")
	if runtime == "" {
		runtime = findCommand("nvidia-container-runtime-hook")
	}
	if runtime == "" {
		toolkit.Reason = "The NVIDIA container runtime and runtime hook were not found."
		return toolkit
	}
	data, err := runVersion(ctx, cli)
	if err != nil || strings.TrimSpace(data) == "" {
		toolkit.Status = "failed"
		toolkit.Reason = "nvidia-container-cli returned an empty version."
		if err != nil {
			toolkit.Reason = err.Error()
		}
		return toolkit
	}
	toolkit.Status = "passed"
	toolkit.Version = strings.SplitN(strings.TrimSpace(data), "\n", 2)[0]
	if data, err := runVersion(ctx, runtime); err != nil || strings.TrimSpace(data) == "" {
		toolkit.Status = "failed"
		toolkit.Reason = "The NVIDIA container runtime returned an empty version."
		if err != nil {
			toolkit.Reason = err.Error()
		}
		return toolkit
	}
	return toolkit
}

func checkNVIDIADockerInfo(info system.Info, hookInstalled bool) dto.GPUEnvironmentCheck {
	result := dto.GPUEnvironmentCheck{Vendor: "nvidia", Item: "docker", Status: "unconfigured", Version: info.ServerVersion}
	if _, ok := info.Runtimes["nvidia"]; ok {
		result.Status = "passed"
		return result
	}
	if hookInstalled {
		result.Status = "unknown"
		result.Reason = "NVIDIA runtime hook is installed, but Docker does not expose its GPU device-driver registration."
	}
	return result
}

func checkNVIDIADockerConfig(info system.Info, readConfig func() (map[string]interface{}, error), findCommand func(string) string) dto.GPUEnvironmentCheck {
	result := checkNVIDIADockerInfo(info, false)
	if result.Status == "passed" {
		runtimePath := info.Runtimes["nvidia"].Path
		if runtimePath == "" {
			result.Status = "unknown"
			result.Reason = "Docker registered an NVIDIA runtime without an executable path; its availability cannot be verified."
			return result
		}
		resolved := runtimePath
		if !filepath.IsAbs(runtimePath) {
			resolved = findCommand(runtimePath)
		}
		if resolved == "" || !cmd.Which(resolved) {
			result.Status = "failed"
			result.Code = "dockerConfigInvalid"
			result.Reason = fmt.Sprintf("The registered NVIDIA runtime %q cannot be found or executed.", runtimePath)
		}
		return result
	}
	if hook := findCommand("nvidia-container-runtime-hook"); hook != "" && cmd.Which(hook) {
		result = checkNVIDIADockerInfo(info, true)
	}
	config, err := readConfig()
	if err != nil {
		var syntaxError *json.SyntaxError
		var typeError *json.UnmarshalTypeError
		switch {
		case os.IsNotExist(err):
			if result.Status == "unknown" {
				return result
			}
			result.Code = "dockerConfigMissing"
			result.Reason = fmt.Sprintf("Docker configuration %s is missing; configure runtimes.nvidia.path and restart Docker.", constant.DaemonJsonPath)
		case errors.As(err, &syntaxError), errors.As(err, &typeError):
			result.Status = "failed"
			result.Code = "dockerConfigInvalid"
			result.Reason = fmt.Sprintf("Cannot parse Docker configuration %s: %v", constant.DaemonJsonPath, err)
		default:
			result.Status = "failed"
			result.Code = "dockerConfigUnreadable"
			result.Reason = fmt.Sprintf("Cannot read Docker configuration %s: %v", constant.DaemonJsonPath, err)
		}
		return result
	}
	if config == nil {
		result.Status = "failed"
		result.Code = "dockerConfigInvalid"
		result.Reason = "Docker configuration must be a JSON object."
		return result
	}
	runtimes, ok := config["runtimes"].(map[string]interface{})
	if config["runtimes"] == nil || (ok && runtimes["nvidia"] == nil) {
		if result.Status == "unknown" {
			return result
		}
		result.Code = "dockerConfigMissing"
		result.Reason = fmt.Sprintf("Docker configuration %s has no runtimes.nvidia entry.", constant.DaemonJsonPath)
		return result
	}
	nvidia, ok := runtimes["nvidia"].(map[string]interface{})
	runtimePath, pathOK := nvidia["path"].(string)
	if !ok || !pathOK || strings.TrimSpace(runtimePath) == "" || (!filepath.IsAbs(runtimePath) && filepath.Base(runtimePath) != runtimePath) {
		result.Status = "failed"
		result.Code = "dockerConfigInvalid"
		result.Reason = "runtimes.nvidia.path must contain a runtime executable name or absolute path."
		return result
	}
	resolved := findCommand(runtimePath)
	if resolved == "" || !cmd.Which(resolved) {
		result.Status = "failed"
		result.Code = "dockerConfigInvalid"
		result.Reason = fmt.Sprintf("The configured NVIDIA runtime %q cannot be found or executed.", runtimePath)
		return result
	}
	result.Status = "unconfigured"
	result.Code = "dockerConfigPending"
	result.Reason = "The default daemon.json contains runtimes.nvidia.path, but the running Docker daemon has not registered the NVIDIA runtime; confirm the daemon uses this configuration file, then restart Docker and check again."
	return result
}

func summarizeGPUs(snapshot *accelerator.Snapshot) []dto.GPUSummary {
	result := []dto.GPUSummary{}
	for _, vendor := range gpuVendors {
		summary := dto.GPUSummary{Vendor: vendor}
		cards, chips := map[string]bool{}, map[string]bool{}
		var versions []string
		for _, device := range snapshot.Devices {
			deviceVendor := device.Vendor
			if device.XPU != nil {
				deviceVendor = "intel"
			}
			if deviceVendor != vendor || device.ParentID != "" || chips[device.ID] {
				continue
			}
			chips[device.ID] = true
			cardID := device.ID
			if device.NPU != nil {
				cardID = strconv.Itoa(device.NPUIndex)
			}
			cards[cardID] = true
			if device.DriverVersion != "" && !slices.Contains(versions, device.DriverVersion) {
				versions = append(versions, device.DriverVersion)
			}
			if !device.Metrics.MemoryTotal.Available() {
				summary.MemoryPartial = true
				continue
			}
			bytes := device.Metrics.MemoryTotal.ValueOrZero() * 1024 * 1024
			if device.GPU != nil && device.GPU.SharedMemory {
				if summary.SharedMemoryTotal == nil || bytes > *summary.SharedMemoryTotal {
					summary.SharedMemoryTotal = &bytes
				}
			} else {
				if summary.MemoryTotal == nil {
					summary.MemoryTotal = new(float64)
				}
				*summary.MemoryTotal += bytes
			}
		}
		if len(cards) > 0 {
			count := len(cards)
			summary.CardCount = &count
			summary.DriverVersion = strings.Join(versions, " / ")
			if vendor == "nvidia" {
				summary.CudaVersion = snapshot.Info.CudaVersion
			}
			result = append(result, summary)
		}
	}
	return result
}
