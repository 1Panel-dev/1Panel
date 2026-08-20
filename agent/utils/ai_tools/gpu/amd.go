package gpu

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	amdSMICommand     = "amd-smi"
	amdSMIDefaultPath = "/opt/rocm/bin/amd-smi"
)

type amdSMI struct {
	command string
}

func findAMDSMI() (string, bool) {
	for _, command := range []string{amdSMICommand, amdSMIDefaultPath} {
		if cmd.Which(command) {
			return command, true
		}
	}
	return "", false
}

func (a amdSMI) LoadInfo(ctx context.Context) (*Info, error) {
	var (
		staticData  string
		metricData  string
		processData string
		staticErr   error
		metricErr   error
		processErr  error
		wg          sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		staticData, staticErr = runAMDSMI(ctx, a.command, "static", "--asic", "--bus", "--driver", "--limit", "--json")
	}()
	go func() {
		defer wg.Done()
		metricData, metricErr = runAMDSMI(ctx, a.command, "metric", "--usage", "--power", "--temperature", "--mem-usage", "--fan", "--perf-level", "--json")
	}()
	go func() {
		defer wg.Done()
		processData, processErr = runAMDSMI(ctx, a.command, "process", "--general", "--json")
	}()
	wg.Wait()

	if staticErr != nil {
		return nil, fmt.Errorf("calling %s static failed: %w", a.command, staticErr)
	}
	info, err := parseAMDStatic(staticData)
	if err != nil {
		return nil, fmt.Errorf("parsing %s static output failed: %w", a.command, err)
	}

	if metricErr != nil {
		global.LOG.Warnf("calling %s metric failed, metrics will be omitted: %v", a.command, metricErr)
	} else if err := applyAMDMetrics(info, metricData); err != nil {
		global.LOG.Warnf("parsing %s metric output failed, metrics will be omitted: %v", a.command, err)
	}
	if processErr != nil {
		global.LOG.Warnf("calling %s process failed, process information will be omitted: %v", a.command, processErr)
	} else if err := applyAMDProcesses(info, processData); err != nil {
		global.LOG.Warnf("parsing %s process output failed, process information will be omitted: %v", a.command, err)
	}
	return info, nil
}

func runAMDSMI(ctx context.Context, command string, args ...string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(10*time.Second))
	return cmdMgr.RunWithStdout(command, args...)
}

func parseAMDStatic(data string) (*Info, error) {
	rows, err := decodeAMDRows(data)
	if err != nil {
		return nil, err
	}

	info := &Info{Type: "amd"}
	for _, row := range rows {
		index, ok := amdGPUIndex(row)
		if !ok {
			continue
		}
		device := Device{
			Type:             "amd",
			Index:            index,
			ProductName:      amdStringAt(row, "asic.market_name", "market_name", "gpu_name"),
			PersistenceMode:  "N/A",
			BusID:            amdStringAt(row, "bus.bdf", "bdf"),
			DisplayActive:    "N/A",
			ECC:              "N/A",
			FanSpeed:         "N/A",
			Temperature:      "N/A",
			PerformanceState: "N/A",
			PowerDraw:        "N/A",
			MaxPowerLimit: amdMetricAt(row, "W",
				"limit.ppt0.max_power_limit",
				"limit.max_power_limit",
				"limit.max_power",
			),
			MemUsed:     "N/A",
			MemTotal:    "N/A",
			GPUUtil:     "N/A",
			ComputeMode: "N/A",
			MigMode:     "N/A",
		}
		if device.ProductName == "" {
			device.ProductName = "AMD GPU"
		}
		if device.MaxPowerLimit == "" {
			device.MaxPowerLimit = "N/A"
		}
		if info.DriverVersion == "" {
			info.DriverVersion = amdStringAt(row, "driver.version", "driver_version", "amdgpu_version")
		}
		info.Devices = append(info.Devices, device)
	}
	sort.Slice(info.Devices, func(i, j int) bool {
		return info.Devices[i].Index < info.Devices[j].Index
	})
	return info, nil
}

func applyAMDMetrics(info *Info, data string) error {
	rows, err := decodeAMDRows(data)
	if err != nil {
		return err
	}
	devices := amdDevicesByIndex(info)
	for _, row := range rows {
		index, ok := amdGPUIndex(row)
		if !ok {
			continue
		}
		device, ok := devices[index]
		if !ok {
			continue
		}
		setAMDMetric(&device.GPUUtil, row, "%", "usage.gfx_activity", "usage.gfx", "gfx_activity", "gfx_usage")
		setAMDMetric(&device.Temperature, row, "°C", "temperature.hotspot", "temperature.edge", "hotspot_temperature", "gpu_temperature", "gpu_temp")
		setAMDMetric(&device.PowerDraw, row, "W", "power.socket_power", "socket_power", "power_usage")
		setAMDMetric(&device.MemUsed, row, "MB", "mem_usage.used_vram", "vram.used", "used_vram", "vram_used")
		setAMDMetric(&device.MemTotal, row, "MB", "mem_usage.total_vram", "vram.total", "total_vram", "vram_total")
		setAMDMetric(&device.FanSpeed, row, "%", "fan.speed", "fan_speed")
		if value := amdStringAt(row, "perf_level", "performance_level"); value != "" {
			device.PerformanceState = value
		}
	}
	return nil
}

func applyAMDProcesses(info *Info, data string) error {
	rows, err := decodeAMDRows(data)
	if err != nil {
		return err
	}
	devices := amdDevicesByIndex(info)
	for _, row := range rows {
		index, ok := amdGPUIndex(row)
		if !ok {
			continue
		}
		device, ok := devices[index]
		if !ok {
			continue
		}
		processList, _ := amdValueAt(row, "process_list")
		items := amdObjectList(processList)
		if len(items) == 0 {
			if _, ok := amdValueAt(row, "process_info"); ok {
				items = amdSMIRows{row}
			}
		}
		for _, processRow := range items {
			pid := amdStringAt(processRow, "process_info.pid", "pid")
			if pid == "" || strings.EqualFold(pid, "N/A") {
				continue
			}
			device.Processes = append(device.Processes, Process{
				PID:         pid,
				Type:        "C",
				ProcessName: amdStringAt(processRow, "process_info.name", "name"),
				UsedMemory: amdMetricAt(processRow, "B",
					"process_info.mem_usage",
					"process_info.mem",
					"process_info.memory_usage.vram_mem",
					"process_info.vram_mem",
					"mem_usage",
				),
			})
		}
	}
	return nil
}

func amdDevicesByIndex(info *Info) map[uint]*Device {
	devices := make(map[uint]*Device, len(info.Devices))
	for index := range info.Devices {
		devices[info.Devices[index].Index] = &info.Devices[index]
	}
	return devices
}

func setAMDMetric(target *string, row amdSMIRow, defaultUnit string, paths ...string) {
	if value := amdMetricAt(row, defaultUnit, paths...); value != "" {
		*target = value
	}
}
