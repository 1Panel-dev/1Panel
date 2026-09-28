package xpu

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const xpuSMICommand = "xpu-smi"

type Client struct{}

func New() (bool, Client) {
	return cmd.Which(xpuSMICommand), Client{}
}

func (c Client) LoadInfo() (*Info, error) {
	return c.LoadInfoContext(context.Background())
}

func (c Client) LoadInfoContext(ctx context.Context) (*Info, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second))
	data, err := cmdMgr.RunWithStdout(xpuSMICommand, "discovery", "-j")
	if err != nil {
		return nil, fmt.Errorf("calling %s failed: %w", xpuSMICommand, err)
	}
	var deviceInfo discoveryInfo
	if err := json.Unmarshal([]byte(data), &deviceInfo); err != nil {
		return nil, fmt.Errorf("deviceInfo json unmarshal failed, err: %w", err)
	}
	res := &Info{
		Type: "xpu",
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, device := range deviceInfo.DeviceList {
		wg.Add(1)
		go c.loadDeviceInfo(ctx, device, &wg, res, &mu)
	}

	wg.Wait()

	processData, err := cmdMgr.RunWithStdout(xpuSMICommand, "ps", "-j")
	if err != nil {
		res.Warnings = append(res.Warnings, fmt.Sprintf("xpu-smi ps: %v", err))
	} else {
		var psList DeviceUtilByProcList
		if err := json.Unmarshal([]byte(processData), &psList); err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("xpu-smi ps output: %v", err))
		} else {
			for i := range res.Devices {
				res.Devices[i].ProcessStatus = "ok"
			}
			for _, ps := range psList.DeviceUtilByProcList {
				process := Process{
					PID:     ps.ProcessID,
					Command: ps.ProcessName,
				}
				if ps.SharedMemSize > 0 {
					process.SHR = fmt.Sprintf("%.1f MiB", ps.SharedMemSize/1024)
				}
				if ps.MemSize > 0 {
					process.Memory = fmt.Sprintf("%.1f MiB", ps.MemSize/1024)
				}
				for index, xpu := range res.Devices {
					if xpu.Basic.DeviceID == ps.DeviceID {
						res.Devices[index].Processes = append(res.Devices[index].Processes, process)
					}
				}
			}
		}
	}
	sort.Slice(res.Devices, func(i, j int) bool {
		return res.Devices[i].Basic.DeviceID < res.Devices[j].Basic.DeviceID
	})

	return res, nil
}

func (c Client) loadDeviceInfo(ctx context.Context, device discoveryDevice, wg *sync.WaitGroup, res *Info, mu *sync.Mutex) {
	defer wg.Done()

	xpu := Device{
		ProcessStatus: "unavailable",
		Basic: Basic{
			UUID:          device.UUID,
			DeviceID:      device.DeviceID,
			DeviceName:    device.DeviceName,
			VendorName:    device.VendorName,
			PciBdfAddress: device.PciBdfAddress,
		},
	}

	var xpuData, statsData string
	var xpuErr, statsErr error

	var wgCmd sync.WaitGroup
	wgCmd.Add(2)

	cmdMgr := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second))
	go func() {
		defer wgCmd.Done()
		xpuData, xpuErr = cmdMgr.RunWithStdout(xpuSMICommand, "discovery", "-d", strconv.Itoa(device.DeviceID), "-j")
	}()

	go func() {
		defer wgCmd.Done()
		statsData, statsErr = cmdMgr.RunWithStdout(xpuSMICommand, "stats", "-d", strconv.Itoa(device.DeviceID), "-j")
	}()

	wgCmd.Wait()

	if xpuErr != nil {
		mu.Lock()
		res.Warnings = append(res.Warnings, fmt.Sprintf("xpu-smi device %d: %v", device.DeviceID, xpuErr))
		res.Devices = append(res.Devices, xpu)
		mu.Unlock()
		return
	}

	var info discoveryDevice
	if err := json.Unmarshal([]byte(xpuData), &info); err != nil {
		mu.Lock()
		res.Warnings = append(res.Warnings, fmt.Sprintf("xpu-smi device %d output: %v", device.DeviceID, err))
		res.Devices = append(res.Devices, xpu)
		mu.Unlock()
		return
	}

	xpu.Basic.DriverVersion = info.DriverVersion

	bytes, err := strconv.ParseInt(info.MemoryPhysicalSizeByte, 10, 64)
	if err != nil {
		global.LOG.Warnf("Error parsing memory size for device %d, err: %v", device.DeviceID, err)
		xpu.Basic.Memory = info.MemoryPhysicalSizeByte
	} else {
		xpu.Basic.Memory = formatMemoryBytes(bytes)
	}
	freeBytes, err := strconv.ParseInt(info.MemoryFreeSizeByte, 10, 64)
	if err != nil {
		xpu.Basic.FreeMemory = info.MemoryFreeSizeByte
	} else {
		xpu.Basic.FreeMemory = formatMemoryBytes(freeBytes)
	}

	if statsErr != nil {
		mu.Lock()
		res.Warnings = append(res.Warnings, fmt.Sprintf("xpu-smi stats %d: %v", device.DeviceID, statsErr))
		mu.Unlock()
	} else {
		var stats DeviceStats
		if err := json.Unmarshal([]byte(statsData), &stats); err != nil {
			mu.Lock()
			res.Warnings = append(res.Warnings, fmt.Sprintf("xpu-smi stats %d output: %v", device.DeviceID, err))
			mu.Unlock()
		} else {
			loadStats(&xpu.Stats, stats.DeviceLevel)
			for _, tile := range stats.TileLevel {
				item := TileStats{TileID: tile.TileID}
				loadStats(&item.Stats, tile.DataList)
				xpu.Tiles = append(xpu.Tiles, item)
			}
		}
	}

	mu.Lock()
	if res.DriverVersion == "" {
		res.DriverVersion = info.DriverVersion
	}
	res.Devices = append(res.Devices, xpu)
	mu.Unlock()
}

func loadStats(stats *Stats, metrics []DeviceLevelMetric) {
	for _, stat := range metrics {
		if stat.Value == nil || math.IsNaN(*stat.Value) || math.IsInf(*stat.Value, 0) {
			continue
		}
		value := *stat.Value
		if value < 0 && stat.MetricsType != "XPUM_STATS_GPU_CORE_TEMPERATURE" && stat.MetricsType != "XPUM_STATS_MEMORY_TEMPERATURE" {
			continue
		}
		switch stat.MetricsType {
		case "XPUM_STATS_MEMORY_TEMPERATURE":
			stats.MemoryTemperature = fmt.Sprintf("%.1f°C", value)
		case "XPUM_STATS_ENGINE_GROUP_COMPUTE_ALL_UTILIZATION":
			stats.ComputeUtil = fmt.Sprintf("%.1f%%", value)
		case "XPUM_STATS_ENGINE_GROUP_MEDIA_ALL_UTILIZATION":
			stats.MediaUtil = fmt.Sprintf("%.1f%%", value)
		case "XPUM_STATS_ENGINE_GROUP_COPY_ALL_UTILIZATION":
			stats.CopyUtil = fmt.Sprintf("%.1f%%", value)
		case "XPUM_STATS_MEDIA_ENGINE_FREQUENCY":
			stats.MediaFrequency = fmt.Sprintf("%.1fMHz", value)
		case "XPUM_STATS_POWER":
			stats.Power = fmt.Sprintf("%.1fW", value)
		case "XPUM_STATS_GPU_UTILIZATION":
			stats.GPUUtil = fmt.Sprintf("%.1f%%", value)
		case "XPUM_STATS_GPU_FREQUENCY":
			stats.Frequency = fmt.Sprintf("%.1fMHz", value)
		case "XPUM_STATS_GPU_CORE_TEMPERATURE":
			stats.Temperature = fmt.Sprintf("%.1f°C", value)
		case "XPUM_STATS_MEMORY_USED":
			stats.MemoryUsed = fmt.Sprintf("%.1f MiB", value)
		case "XPUM_STATS_MEMORY_UTILIZATION":
			stats.MemoryUtil = fmt.Sprintf("%.1f%%", value)
		case "XPUM_STATS_MEMORY_BANDWIDTH", "XPUM_STATS_MEMORY_BANDWIDTH_UTILIZATION":
			stats.MemoryBandwidthUtil = fmt.Sprintf("%.1f%%", value)
		}
	}
}

func formatMemoryBytes(bytes int64) string {
	return fmt.Sprintf("%.1f MiB", float64(bytes)/(1024*1024))
}
