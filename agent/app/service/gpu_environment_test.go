package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	badgerdb "github.com/1Panel-dev/1Panel/agent/init/cache/db"
	"github.com/docker/docker/api/types/system"
)

func TestDetectGPUHardwareFiltersNonAccelerators(t *testing.T) {
	directory := t.TempDir()
	for name, values := range map[string][2]string{
		"nvidia-display": {"0x10de", "0x030200"},
		"nvidia-compute": {"0x10de", "0x120000"},
		"intel-display":  {"0x8086", "0x030000"},
		"intel-network":  {"0x8086", "0x020000"},
		"ascend-compute": {"0x19e5", "0x120000"},
		"huawei-bmc":     {"0x19e5", "0x030000"},
		"amd-display":    {"0x1002", "0x030000"},
		"other-display":  {"0x1234", "0x030000"},
	} {
		path := filepath.Join(directory, name)
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		for filename, value := range map[string]string{"vendor": values[0], "class": values[1]} {
			if err := os.WriteFile(filepath.Join(path, filename), []byte(value+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	devices, err := detectGPUHardware(directory)
	if err != nil {
		t.Fatal(err)
	}
	for vendor, count := range map[string]int{"nvidia": 2, "intel": 1, "ascend": 1, "amd": 1} {
		if devices[vendor] != count {
			t.Errorf("%s: got %d devices, want %d", vendor, devices[vendor], count)
		}
	}
}

func TestGPUEnvironmentDistinguishesMissingFailedAndAbsent(t *testing.T) {
	tests := []struct {
		name     string
		probe    gpuEnvironmentProbe
		err      error
		hardware string
		status   string
	}{
		{name: "no hardware", probe: gpuEnvironmentProbe{}, hardware: "not_detected"},
		{name: "stale management utility without hardware", probe: gpuEnvironmentProbe{Commands: map[string]bool{"nvidia": true}}, err: errors.New("nvidia-smi: No devices were found"), hardware: "not_detected"},
		{name: "missing management utility", probe: gpuEnvironmentProbe{Hardware: map[string]int{"nvidia": 1}}, hardware: "detected", status: "missing"},
		{name: "command failed", probe: gpuEnvironmentProbe{Hardware: map[string]int{"nvidia": 1}, Commands: map[string]bool{"nvidia": true}}, err: errors.New("nvidia-smi timed out"), hardware: "detected", status: "failed"},
		{name: "unreadable inventory", probe: gpuEnvironmentProbe{HardwareError: "permission denied"}, hardware: "unknown", status: "unknown"},
		{name: "incomplete inventory with detected device", probe: gpuEnvironmentProbe{Hardware: map[string]int{"intel": 1}, HardwareError: "PCI inventory incomplete"}, hardware: "detected", status: "missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := assembleGPUEnvironment(tt.probe, nil, tt.err, nil)
			if result.HardwareStatus != tt.hardware {
				t.Fatalf("hardware status = %s, want %s", result.HardwareStatus, tt.hardware)
			}
			if tt.status == "" {
				if len(result.Checks) != 0 {
					t.Fatalf("no hardware should not request driver installation: %+v", result.Checks)
				}
				return
			}
			if len(result.Checks) == 0 || result.Checks[0].Status != tt.status {
				t.Fatalf("checks = %+v, want first status %s", result.Checks, tt.status)
			}
			if tt.probe.HardwareError != "" && result.Checks[len(result.Checks)-1].Reason != tt.probe.HardwareError {
				t.Fatalf("inventory error was lost: %+v", result.Checks)
			}
		})
	}
}

func TestGPUEnvironmentRetainsWorkingVendorWhenAnotherFails(t *testing.T) {
	probe := gpuEnvironmentProbe{
		Hardware: map[string]int{"nvidia": 1, "intel": 1},
		Commands: map[string]bool{"nvidia": true, "intel": true},
		Checks:   []dto.GPUEnvironmentCheck{{Vendor: "nvidia", Item: "toolkit", Status: "missing"}},
	}
	result := assembleGPUEnvironment(probe, []dto.GPUSummary{{Vendor: "nvidia", DriverVersion: "570.1"}}, nil, []string{"xpu-smi discovery failed"})
	if len(result.Checks) != 3 || result.Checks[0].Status != "passed" || result.Checks[1].Status != "failed" || result.Checks[2].Item != "toolkit" {
		t.Fatalf("unexpected independent checks: %+v", result.Checks)
	}
}

func TestGPUEnvironmentKeepsNVIDIAChecksIndependent(t *testing.T) {
	for _, tt := range []struct {
		name          string
		driverPresent bool
		toolkitStatus string
	}{
		{name: "driver missing only", toolkitStatus: "passed"},
		{name: "toolkit missing only", driverPresent: true, toolkitStatus: "missing"},
		{name: "driver and toolkit missing", toolkitStatus: "missing"},
		{name: "all checks passed", driverPresent: true, toolkitStatus: "passed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			probe := gpuEnvironmentProbe{
				Hardware: map[string]int{"nvidia": 1},
				Commands: map[string]bool{"nvidia": tt.driverPresent},
				Checks: []dto.GPUEnvironmentCheck{
					{Vendor: "nvidia", Item: "toolkit", Status: tt.toolkitStatus},
					{Vendor: "nvidia", Item: "docker", Status: "passed"},
				},
			}
			var summaries []dto.GPUSummary
			driverStatus := "missing"
			if tt.driverPresent {
				summaries = []dto.GPUSummary{{Vendor: "nvidia", DriverVersion: "570.1"}}
				driverStatus = "passed"
			}
			result := assembleGPUEnvironment(probe, summaries, nil, nil)
			if result.HardwareStatus != "detected" || len(result.Checks) != 3 {
				t.Fatalf("expected detected hardware and three independent checks: %+v", result)
			}
			want := map[string]string{"driver": driverStatus, "toolkit": tt.toolkitStatus, "docker": "passed"}
			for _, check := range result.Checks {
				expected, ok := want[check.Item]
				if !ok || check.Vendor != "nvidia" || check.Status != expected {
					t.Fatalf("unexpected check %+v, expected statuses %+v", check, want)
				}
				delete(want, check.Item)
			}
			if len(want) != 0 {
				t.Fatalf("missing checks: %+v", want)
			}
		})
	}
}

func TestNVIDIAToolkitRequiresRuntimeComponentsAndExecutableCLI(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		err      error
		output   string
		status   string
	}{
		{name: "not installed", status: "missing"},
		{name: "base package only", commands: []string{"nvidia-ctk"}, status: "missing"},
		{name: "CLI without runtime", commands: []string{"nvidia-container-cli"}, status: "missing"},
		{name: "runtime without CLI", commands: []string{"nvidia-container-runtime"}, status: "missing"},
		{name: "modern installation", commands: []string{"nvidia-ctk", "nvidia-container-cli", "nvidia-container-runtime"}, output: "version 1.17.4\nmore details", status: "passed"},
		{name: "legacy installation", commands: []string{"nvidia-container-cli", "nvidia-container-runtime-hook"}, output: "cli-version: 1.17.4", status: "passed"},
		{name: "broken CLI", commands: []string{"nvidia-container-cli", "nvidia-container-runtime"}, err: errors.New("permission denied"), status: "failed"},
		{name: "empty version", commands: []string{"nvidia-container-cli", "nvidia-container-runtime"}, status: "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := map[string]bool{}
			for _, command := range tt.commands {
				commands[command] = true
			}
			check := checkNVIDIAToolkit(context.Background(), func(name string) string {
				if commands[name] {
					return name
				}
				return ""
			}, func(context.Context, string) (string, error) { return tt.output, tt.err })
			if check.Status != tt.status {
				t.Fatalf("status = %s, want %s: %+v", check.Status, tt.status, check)
			}
			if check.Status == "passed" && (check.Version == "" || strings.Contains(check.Version, "\n")) {
				t.Fatalf("invalid displayed version: %q", check.Version)
			}
		})
	}
}

func TestNVIDIADockerConfigurationDoesNotRequireDefaultRuntime(t *testing.T) {
	info := system.Info{DefaultRuntime: "runc", Runtimes: map[string]system.RuntimeWithStatus{"nvidia": {}}}
	if check := checkNVIDIADockerInfo(info, false); check.Status != "passed" {
		t.Fatalf("registered NVIDIA runtime should pass with default runc: %+v", check)
	}
	if check := checkNVIDIADockerInfo(system.Info{}, true); check.Status != "unknown" {
		t.Fatalf("hook-based DeviceRequests cannot be confirmed from Docker Info: %+v", check)
	}
	if check := checkNVIDIADockerInfo(system.Info{}, false); check.Status != "unconfigured" {
		t.Fatalf("absent configuration should be reported: %+v", check)
	}
}

func TestNVIDIADockerConfigValidatesDefaultFileWithoutChangingIt(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		missing    bool
		unreadable bool
		lookup     string
		status     string
		code       string
	}{
		{name: "missing file", missing: true, status: "unconfigured", code: "dockerConfigMissing"},
		{name: "missing runtimes", content: `{}`, status: "unconfigured", code: "dockerConfigMissing"},
		{name: "missing NVIDIA runtime", content: `{"runtimes":{"runc":{"path":"runc"}}}`, status: "unconfigured", code: "dockerConfigMissing"},
		{name: "null runtimes", content: `{"runtimes":null}`, status: "unconfigured", code: "dockerConfigMissing"},
		{name: "malformed JSON", content: `{`, status: "failed", code: "dockerConfigInvalid"},
		{name: "null document", content: `null`, status: "failed", code: "dockerConfigInvalid"},
		{name: "array document", content: `[]`, status: "failed", code: "dockerConfigInvalid"},
		{name: "string document", content: `"configuration"`, status: "failed", code: "dockerConfigInvalid"},
		{name: "invalid runtimes type", content: `{"runtimes":"nvidia"}`, status: "failed", code: "dockerConfigInvalid"},
		{name: "invalid NVIDIA entry type", content: `{"runtimes":{"nvidia":"nvidia-container-runtime"}}`, status: "failed", code: "dockerConfigInvalid"},
		{name: "missing path", content: `{"runtimes":{"nvidia":{}}}`, status: "failed", code: "dockerConfigInvalid"},
		{name: "empty path", content: `{"runtimes":{"nvidia":{"path":""}}}`, status: "failed", code: "dockerConfigInvalid"},
		{name: "whitespace path", content: `{"runtimes":{"nvidia":{"path":"   "}}}`, status: "failed", code: "dockerConfigInvalid"},
		{name: "invalid path type", content: `{"runtimes":{"nvidia":{"path":42}}}`, status: "failed", code: "dockerConfigInvalid"},
		{name: "relative directory path", content: `{"runtimes":{"nvidia":{"path":"bin/nvidia-container-runtime"}}}`, status: "failed", code: "dockerConfigInvalid"},
		{name: "read permission denied", unreadable: true, status: "failed", code: "dockerConfigUnreadable"},
		{name: "command name without args not loaded", content: `{"runtimes":{"nvidia":{"path":"nvidia-container-runtime"}}}`, lookup: "executable", status: "unconfigured", code: "dockerConfigPending"},
		{name: "absolute path with empty args not loaded", content: `{"runtimes":{"nvidia":{"path":RUNTIME_PATH,"args":[]}}}`, lookup: "executable", status: "unconfigured", code: "dockerConfigPending"},
		{name: "optional runtimeArgs not loaded", content: `{"runtimes":{"nvidia":{"path":"nvidia-container-runtime","runtimeArgs":[]}}}`, lookup: "executable", status: "unconfigured", code: "dockerConfigPending"},
		{name: "configured executable not found", content: `{"runtimes":{"nvidia":{"path":"nvidia-container-runtime"}}}`, lookup: "missing", status: "failed", code: "dockerConfigInvalid"},
		{name: "configured file not executable", content: `{"runtimes":{"nvidia":{"path":RUNTIME_PATH}}}`, lookup: "not-executable", status: "failed", code: "dockerConfigInvalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previousPath := constant.DaemonJsonPath
			constant.DaemonJsonPath = filepath.Join(t.TempDir(), "daemon.json")
			t.Cleanup(func() { constant.DaemonJsonPath = previousPath })
			runtimePath := filepath.Join(t.TempDir(), "nvidia-container-runtime")
			mode := os.FileMode(0o755)
			if tt.lookup == "not-executable" {
				mode = 0o644
			}
			if tt.lookup == "executable" || tt.lookup == "not-executable" {
				if err := os.WriteFile(runtimePath, []byte("#!/bin/sh\nexit 0\n"), mode); err != nil {
					t.Fatal(err)
				}
			}
			encodedPath, err := json.Marshal(runtimePath)
			if err != nil {
				t.Fatal(err)
			}
			content := strings.ReplaceAll(tt.content, "RUNTIME_PATH", string(encodedPath))
			if !tt.missing && !tt.unreadable {
				if err := os.WriteFile(constant.DaemonJsonPath, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			readConfig := readDockerDaemonConfig
			if tt.unreadable {
				// Inject EACCES rather than relying on Unix file mode behavior
				// when a test runner has elevated privileges.
				readConfig = func() (map[string]interface{}, error) {
					return nil, &os.PathError{Op: "open", Path: constant.DaemonJsonPath, Err: os.ErrPermission}
				}
			}
			check := checkNVIDIADockerConfig(system.Info{DefaultRuntime: "runc"}, readConfig, func(name string) string {
				if tt.lookup == "" {
					t.Fatalf("invalid or missing configuration should not resolve an executable: %q", name)
				}
				if name != "nvidia-container-runtime" && name != runtimePath {
					t.Fatalf("unexpected configured runtime path: %q", name)
				}
				if tt.lookup == "missing" {
					return ""
				}
				return runtimePath
			})
			if check.Vendor != "nvidia" || check.Item != "docker" || check.Status != tt.status || check.Code != tt.code || check.Reason == "" {
				t.Fatalf("unexpected Docker check: %+v, want status=%s code=%s", check, tt.status, tt.code)
			}
			if tt.missing {
				if _, err := os.Stat(constant.DaemonJsonPath); !os.IsNotExist(err) {
					t.Fatalf("checking a missing configuration must not create it: %v", err)
				}
			} else if !tt.unreadable {
				data, err := os.ReadFile(constant.DaemonJsonPath)
				if err != nil || string(data) != content {
					t.Fatalf("configuration changed while inspecting it: data=%q err=%v", data, err)
				}
			}
		})
	}
}

func TestNVIDIADockerConfigTrustsLoadedRuntimeWithoutReadingDefaultFile(t *testing.T) {
	info := system.Info{
		ServerVersion:  "28.5.2",
		DefaultRuntime: "runc",
		Runtimes:       map[string]system.RuntimeWithStatus{"nvidia": {}},
	}
	reads, lookups := 0, 0
	check := checkNVIDIADockerConfig(info, func() (map[string]interface{}, error) {
		reads++
		return nil, os.ErrNotExist
	}, func(string) string {
		lookups++
		return ""
	})
	if check.Status != "passed" || check.Code != "" || check.Version != info.ServerVersion || reads != 0 || lookups != 0 {
		t.Fatalf("loaded runtime must not depend on the default config file or local executable lookup: check=%+v reads=%d lookups=%d", check, reads, lookups)
	}
}

func TestNVIDIAToolkitRejectsBrokenRuntime(t *testing.T) {
	check := checkNVIDIAToolkit(context.Background(), func(name string) string {
		return name
	}, func(_ context.Context, name string) (string, error) {
		if name == "nvidia-container-runtime" {
			return "", errors.New("runtime: shared library could not be loaded")
		}
		return "version 1.17.4", nil
	})
	if check.Status != "failed" || !strings.Contains(check.Reason, "shared library") {
		t.Fatalf("broken runtime was treated as usable: %+v", check)
	}
}

func TestGPUEnvironmentProbeReusesExistingCache(t *testing.T) {
	previous := global.CACHE
	global.CACHE = badgerdb.NewCacheDB()
	t.Cleanup(func() { global.CACHE = previous })
	probe := gpuEnvironmentProbe{CheckedAt: time.Now().Add(-time.Second), Hardware: map[string]int{"nvidia": 1}, NVIDIAChecked: true}
	data, err := json.Marshal(probe)
	if err != nil {
		t.Fatal(err)
	}
	global.CACHE.SetWithTTL(gpuEnvironmentCacheKey, string(data), time.Minute)
	result := loadGPUEnvironmentProbe(context.Background(), false, true)
	if !result.CheckedAt.Equal(probe.CheckedAt) || result.Hardware["nvidia"] != 1 {
		t.Fatalf("cache was not reused: %+v", result)
	}
}

func TestGPUEnvironmentProbeCancellationWhileWaiting(t *testing.T) {
	gpuEnvironmentLock <- struct{}{}
	defer func() { <-gpuEnvironmentLock }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := loadGPUEnvironmentProbe(ctx, false, false)
	if result.HardwareError != context.Canceled.Error() {
		t.Fatalf("canceled wait should finish without probing: %+v", result)
	}
}
