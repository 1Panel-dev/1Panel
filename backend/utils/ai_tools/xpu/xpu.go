package xpu

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type XpuSMI struct{}

func New() (bool, XpuSMI) {
	return cmd.Which("xpu-smi"), XpuSMI{}
}

func (x XpuSMI) loadDeviceData(device Device, wg *sync.WaitGroup, res *[]XPUSimpleInfo, mu *sync.Mutex) {
	defer wg.Done()

	var xpu XPUSimpleInfo
	xpu.DeviceID = device.DeviceID
	xpu.DeviceName = device.DeviceName

	var xpuData, statsData string
	var xpuErr, statsErr error

	var wgCmd sync.WaitGroup
	wgCmd.Add(2)

	go func() {
		defer wgCmd.Done()
		xpuData, xpuErr = cmd.ExecWithTimeOut(fmt.Sprintf("xpu-smi discovery -d %d -j", device.DeviceID), 5*time.Second)
	}()

	go func() {
		defer wgCmd.Done()
		statsData, statsErr = cmd.ExecWithTimeOut(fmt.Sprintf("xpu-smi stats -d %d -j", device.DeviceID), 5*time.Second)
	}()

	wgCmd.Wait()

	if xpuErr != nil {
		global.LOG.Errorf("calling xpu-smi discovery failed for device %d, err: %v\n", device.DeviceID, xpuErr)
		return
	}

	var info Device
	if err := json.Unmarshal([]byte(xpuData), &info); err != nil {
		global.LOG.Errorf("xpuData json unmarshal failed for device %d, err: %v\n", device.DeviceID, err)
		return
	}

	bytes, err := strconv.ParseInt(info.MemoryPhysicalSizeByte, 10, 64)
	if err != nil {
		global.LOG.Errorf("Error parsing memory size for device %d, err: %v\n", device.DeviceID, err)
		return
	}
	xpu.Memory = fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))

	if statsErr != nil {
		global.LOG.Errorf("calling xpu-smi stats failed for device %d, err: %v\n", device.DeviceID, statsErr)
		return
	}

	var stats DeviceStats
	if err := json.Unmarshal([]byte(statsData), &stats); err != nil {
		global.LOG.Errorf("statsData json unmarshal failed for device %d, err: %v\n", device.DeviceID, err)
		return
	}

	for _, stat := range stats.DeviceLevel {
		switch stat.MetricsType {
		case "XPUM_STATS_POWER":
			xpu.Power = fmt.Sprintf("%.1fW", stat.Value)
		case "XPUM_STATS_GPU_CORE_TEMPERATURE":
			xpu.Temperature = fmt.Sprintf("%.1f°C", stat.Value)
		case "XPUM_STATS_MEMORY_USED":
			xpu.MemoryUsed = fmt.Sprintf("%.1fMB", stat.Value)
		case "XPUM_STATS_MEMORY_UTILIZATION":
			xpu.MemoryUtil = fmt.Sprintf("%.1f%%", stat.Value)
		}
	}

	mu.Lock()
	*res = append(*res, xpu)
	mu.Unlock()
}

func (x XpuSMI) LoadDashData() ([]XPUSimpleInfo, error) {
	data, err := cmd.ExecWithTimeOut("xpu-smi discovery -j", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("calling xpu-smi failed, err: %w", err)
	}

	var deviceInfo DeviceInfo
	if err := json.Unmarshal([]byte(data), &deviceInfo); err != nil {
		return nil, fmt.Errorf("deviceInfo json unmarshal failed, err: %w", err)
	}

	var res []XPUSimpleInfo
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, device := range deviceInfo.DeviceList {
		wg.Add(1)
		go x.loadDeviceData(device, &wg, &res, &mu)
	}

	wg.Wait()

	sort.Slice(res, func(i, j int) bool {
		return res[i].DeviceID < res[j].DeviceID
	})
	return res, nil
}

func (x XpuSMI) LoadGpuInfo() (*XpuInfo, error) {
	data, err := cmd.ExecWithTimeOut("xpu-smi discovery -j", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("calling xpu-smi  failed, err: %w", err)
	}

	var deviceInfo DeviceInfo
	if err := json.Unmarshal([]byte(data), &deviceInfo); err != nil {
		return nil, fmt.Errorf("deviceInfo json unmarshal failed, err: %w", err)
	}

	res := &XpuInfo{
		Type: "xpu",
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, device := range deviceInfo.DeviceList {
		wg.Add(1)
		go x.loadDeviceInfo(device, &wg, res, &mu)
	}

	wg.Wait()

	processData, err := cmd.ExecWithTimeOut("xpu-smi ps -j", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("calling xpu-smi ps failed, err: %w", err)
	}
	var psList DeviceUtilByProcList
	if err := json.Unmarshal([]byte(processData), &psList); err != nil {
		return nil, fmt.Errorf("processData json unmarshal failed, err: %w", err)
	}
	for _, ps := range psList.DeviceUtilByProcList {
		process := Process{
			PID:     ps.ProcessID,
			Command: ps.ProcessName,
		}
		if ps.SharedMemSize > 0 {
			process.SHR = fmt.Sprintf("%.1f MB", ps.SharedMemSize/1024)
		}
		if ps.MemSize > 0 {
			process.Memory = fmt.Sprintf("%.1f MB", ps.MemSize/1024)
		}
		for index, xpu := range res.Xpu {
			if xpu.Basic.DeviceID == ps.DeviceID {
				res.Xpu[index].Processes = append(res.Xpu[index].Processes, process)
			}
		}
	}

	return res, nil
}

func (x XpuSMI) loadDeviceInfo(device Device, wg *sync.WaitGroup, res *XpuInfo, mu *sync.Mutex) {
	defer wg.Done()

	xpu := Xpu{
		Basic: Basic{
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

	go func() {
		defer wgCmd.Done()
		xpuData, xpuErr = cmd.ExecWithTimeOut(fmt.Sprintf("xpu-smi discovery -d %d -j", device.DeviceID), 5*time.Second)
	}()

	go func() {
		defer wgCmd.Done()
		statsData, statsErr = cmd.ExecWithTimeOut(fmt.Sprintf("xpu-smi stats -d %d -j", device.DeviceID), 5*time.Second)
	}()

	wgCmd.Wait()

	if xpuErr != nil {
		global.LOG.Errorf("calling xpu-smi discovery failed for device %d, err: %v\n", device.DeviceID, xpuErr)
		return
	}

	var info Device
	if err := json.Unmarshal([]byte(xpuData), &info); err != nil {
		global.LOG.Errorf("xpuData json unmarshal failed for device %d, err: %v\n", device.DeviceID, err)
		return
	}

	res.DriverVersion = info.DriverVersion
	xpu.Basic.DriverVersion = info.DriverVersion

	bytes, err := strconv.ParseInt(info.MemoryPhysicalSizeByte, 10, 64)
	if err != nil {
		global.LOG.Errorf("Error parsing memory size for device %d, err: %v\n", device.DeviceID, err)
		return
	}
	xpu.Basic.Memory = fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	xpu.Basic.FreeMemory = info.MemoryFreeSizeByte

	if statsErr != nil {
		global.LOG.Errorf("calling xpu-smi stats failed for device %d, err: %v\n", device.DeviceID, statsErr)
		return
	}

	var stats DeviceStats
	if err := json.Unmarshal([]byte(statsData), &stats); err != nil {
		global.LOG.Errorf("statsData json unmarshal failed for device %d, err: %v\n", device.DeviceID, err)
		return
	}

	for _, stat := range stats.DeviceLevel {
		switch stat.MetricsType {
		case "XPUM_STATS_POWER":
			xpu.Stats.Power = fmt.Sprintf("%.1fW", stat.Value)
		case "XPUM_STATS_GPU_FREQUENCY":
			xpu.Stats.Frequency = fmt.Sprintf("%.1fMHz", stat.Value)
		case "XPUM_STATS_GPU_CORE_TEMPERATURE":
			xpu.Stats.Temperature = fmt.Sprintf("%.1f°C", stat.Value)
		case "XPUM_STATS_MEMORY_USED":
			xpu.Stats.MemoryUsed = fmt.Sprintf("%.1fMB", stat.Value)
		case "XPUM_STATS_MEMORY_UTILIZATION":
			xpu.Stats.MemoryUtil = fmt.Sprintf("%.1f%%", stat.Value)
		}
	}

	mu.Lock()
	res.Xpu = append(res.Xpu, xpu)
	mu.Unlock()
}
