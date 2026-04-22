package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/psutil"
	"github.com/robfig/cron/v3"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type MonitorService struct {
	DiskIO chan ([]disk.IOCountersStat)
	NetIO  chan ([]net.IOCountersStat)
}

var monitorCancel context.CancelFunc

type IMonitorService interface {
	Run()
	LoadMonitorData(req dto.MonitorSearch) ([]dto.MonitorData, error)
	LoadGPUOptions() dto.MonitorGPUOptions
	LoadGPUMonitorData(req dto.MonitorGPUSearch) (dto.MonitorGPUData, error)
	LoadSetting() (*dto.MonitorSetting, error)
	UpdateSetting(key, value string) error
	CleanData() error

	saveIODataToDB(ctx context.Context, interval float64)
	saveNetDataToDB(ctx context.Context, interval float64)
}

func NewIMonitorService() IMonitorService {
	return &MonitorService{
		DiskIO: make(chan []disk.IOCountersStat, 2),
		NetIO:  make(chan []net.IOCountersStat, 2),
	}
}

func (m *MonitorService) LoadMonitorData(req dto.MonitorSearch) ([]dto.MonitorData, error) {
	loc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	req.StartTime = req.StartTime.In(loc)
	req.EndTime = req.EndTime.In(loc)

	var data []dto.MonitorData
	if req.Param == "all" || req.Param == "cpu" || req.Param == "memory" || req.Param == "load" {
		bases, err := monitorRepo.GetBase(repo.WithByCreatedAt(req.StartTime, req.EndTime))
		if err != nil {
			return nil, err
		}

		var itemData dto.MonitorData
		itemData.Param = "base"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			if req.Param == "all" || req.Param == "cpu" {
				var processes []dto.Process
				_ = json.Unmarshal([]byte(base.TopCPU), &processes)
				base.TopCPUItems = processes
				base.TopCPU = ""
			}
			if req.Param == "all" || req.Param == "mem" {
				var processes []dto.Process
				_ = json.Unmarshal([]byte(base.TopMem), &processes)
				base.TopMemItems = processes
				base.TopMem = ""
			}
			itemData.Value = append(itemData.Value, base)
		}
		data = append(data, itemData)
	}
	if req.Param == "all" || req.Param == "io" {
		bases, err := monitorRepo.GetIO(repo.WithByName(req.IO), repo.WithByCreatedAt(req.StartTime, req.EndTime))
		if err != nil {
			return nil, err
		}

		var itemData dto.MonitorData
		itemData.Param = "io"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			itemData.Value = append(itemData.Value, base)
		}
		data = append(data, itemData)
	}
	if req.Param == "all" || req.Param == "network" {
		bases, err := monitorRepo.GetNetwork(repo.WithByName(req.Network), repo.WithByCreatedAt(req.StartTime, req.EndTime))
		if err != nil {
			return nil, err
		}

		var itemData dto.MonitorData
		itemData.Param = "network"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			itemData.Value = append(itemData.Value, base)
		}
		data = append(data, itemData)
	}
	return data, nil
}

func (m *MonitorService) LoadGPUOptions() dto.MonitorGPUOptions {
	var data dto.MonitorGPUOptions
	gpuExist, gpuClient := gpu.New()
	xpuExist, xpuClient := xpu.New()
	if !gpuExist && !xpuExist {
		return data
	}
	if gpuExist {
		data.GPUType = "gpu"
		gpuInfo, err := gpuClient.LoadGpuInfo()
		if err != nil || len(gpuInfo.GPUs) == 0 {
			global.LOG.Error("Load GPU info failed or no GPU found, err: ", err)
			return data
		}
		sort.Slice(gpuInfo.GPUs, func(i, j int) bool {
			return gpuInfo.GPUs[i].Index < gpuInfo.GPUs[j].Index
		})
		for _, item := range gpuInfo.GPUs {
			var chartHide dto.GPUChartHide
			chartHide.ProductName = fmt.Sprintf("%d - %s", item.Index, item.ProductName)
			chartHide.GPU = item.GPUUtil == "" || item.GPUUtil == "N/A"
			if (item.MemTotal == "" || item.MemTotal == "N/A") && (item.MemUsed == "" || item.MemUsed == "N/A") {
				chartHide.Memory = true
			}
			if (item.MaxPowerLimit == "" || item.MaxPowerLimit == "N/A") && (item.PowerDraw == "" || item.PowerDraw == "N/A") {
				chartHide.Power = true
			}
			chartHide.Temperature = item.Temperature == "" || item.Temperature == "N/A"
			chartHide.Speed = item.FanSpeed == "" || item.FanSpeed == "N/A"
			data.ChartHide = append(data.ChartHide, chartHide)
			data.Options = append(data.Options, fmt.Sprintf("%d - %s", item.Index, item.ProductName))
		}
		return data
	} else {
		data.GPUType = "xpu"
		xpu, err := xpuClient.LoadGpuInfo()
		if err != nil || len(xpu.Xpu) == 0 {
			global.LOG.Error("Load XPU info failed or no XPU found, err: ", err)
		}
		sort.Slice(xpu.Xpu, func(i, j int) bool {
			return xpu.Xpu[i].Basic.DeviceID < xpu.Xpu[j].Basic.DeviceID
		})
		for _, item := range xpu.Xpu {
			var chartHide dto.GPUChartHide
			chartHide.GPU = true
			chartHide.Speed = true
			chartHide.ProductName = fmt.Sprintf("%d - %s", item.Basic.DeviceID, item.Basic.DeviceName)
			if (item.Stats.MemoryUsed == "" || item.Stats.MemoryUsed == "N/A") && (item.Basic.Memory == "" || item.Basic.FreeMemory == "N/A") {
				chartHide.Memory = true
			}
			if item.Stats.Power == "" || item.Stats.Power == "N/A" {
				chartHide.Power = true
			}
			chartHide.Temperature = item.Stats.Temperature == "" || item.Stats.Temperature == "N/A"
			data.ChartHide = append(data.ChartHide, chartHide)
			data.Options = append(data.Options, fmt.Sprintf("%d - %s", item.Basic.DeviceID, item.Basic.DeviceName))
		}
		return data
	}
}

func (m *MonitorService) LoadGPUMonitorData(req dto.MonitorGPUSearch) (dto.MonitorGPUData, error) {
	loc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	req.StartTime = req.StartTime.In(loc)
	req.EndTime = req.EndTime.In(loc)
	var data dto.MonitorGPUData
	gpuList, err := monitorRepo.GetGPU(repo.WithByCreatedAt(req.StartTime, req.EndTime), monitorRepo.WithByProductName(req.ProductName))
	if err != nil {
		return data, err
	}

	for _, gpu := range gpuList {
		data.Date = append(data.Date, gpu.CreatedAt)
		data.GPUValue = append(data.GPUValue, gpu.GPUUtil)
		data.TemperatureValue = append(data.TemperatureValue, gpu.Temperature)
		data.PowerUsed = append(data.PowerUsed, gpu.PowerDraw)
		data.PowerTotal = append(data.PowerTotal, gpu.MaxPowerLimit)
		if gpu.MaxPowerLimit != 0 {
			data.PowerPercent = append(data.PowerPercent, gpu.PowerDraw/gpu.MaxPowerLimit*100)
		} else {
			data.PowerPercent = append(data.PowerPercent, float64(0))
		}

		data.MemoryTotal = append(data.MemoryTotal, gpu.MemTotal)
		data.MemoryUsed = append(data.MemoryUsed, gpu.MemUsed)
		if gpu.MemTotal != 0 {
			data.MemoryPercent = append(data.MemoryPercent, gpu.MemUsed/gpu.MemTotal*100)
		} else {
			data.MemoryPercent = append(data.MemoryPercent, float64(0))
		}
		var process []dto.GPUProcess
		if err := json.Unmarshal([]byte(gpu.Processes), &process); err == nil {
			data.ProcessCount = append(data.ProcessCount, len(process))
			data.GPUProcesses = append(data.GPUProcesses, process)
		} else {
			data.ProcessCount = append(data.ProcessCount, 0)
			data.GPUProcesses = append(data.GPUProcesses, []dto.GPUProcess{})
		}
		data.SpeedValue = append(data.SpeedValue, gpu.FanSpeed)
	}
	return data, nil
}

func (m *MonitorService) LoadSetting() (*dto.MonitorSetting, error) {
	setting, err := settingRepo.GetList()
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}
	settingMap := make(map[string]string)
	for _, set := range setting {
		settingMap[set.Key] = set.Value
	}
	var info dto.MonitorSetting
	arr, err := json.Marshal(settingMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr, &info); err != nil {
		return nil, err
	}
	return &info, err
}

func (m *MonitorService) UpdateSetting(key, value string) error {
	switch key {
	case "MonitorStatus":
		if value == constant.StatusEnable && global.MonitorCronID == 0 {
			interval, err := settingRepo.Get(settingRepo.WithByKey("MonitorInterval"))
			if err != nil {
				return err
			}
			if err := StartMonitor(false, interval.Value); err != nil {
				return err
			}
		}
		if value == constant.StatusDisable && global.MonitorCronID != 0 {
			monitorCancel()
			global.Cron.Remove(cron.EntryID(global.MonitorCronID))
			global.MonitorCronID = 0
		}
	case "MonitorInterval":
		status, err := settingRepo.Get(settingRepo.WithByKey("MonitorStatus"))
		if err != nil {
			return err
		}
		if status.Value == constant.StatusEnable && global.MonitorCronID != 0 {
			if err := StartMonitor(true, value); err != nil {
				return err
			}
		}
	}
	return settingRepo.Update(key, value)
}

func (m *MonitorService) CleanData() error {
	if err := global.MonitorDB.Exec("DELETE FROM monitor_bases").Error; err != nil {
		return err
	}
	if err := global.MonitorDB.Exec("DELETE FROM monitor_ios").Error; err != nil {
		return err
	}
	if err := global.MonitorDB.Exec("DELETE FROM monitor_networks").Error; err != nil {
		return err
	}
	_ = global.GPUMonitorDB.Exec("DELETE FROM monitor_gpus").Error
	return nil
}

func (m *MonitorService) Run() {
	saveGPUDataToDB()
	saveXPUDataToDB()
	var itemModel model.MonitorBase
	totalPercent, _ := cpu.Percent(3*time.Second, false)
	if len(totalPercent) == 1 {
		itemModel.Cpu = totalPercent[0]
	}
	topCPU := loadTopCPU()
	if len(topCPU) != 0 {
		topItemCPU, err := json.Marshal(topCPU)
		if err == nil {
			itemModel.TopCPU = string(topItemCPU)
		}
	}
	cpuCount, _ := psutil.CPUInfo.GetPhysicalCores(false)
	loadInfo, _ := load.Avg()
	itemModel.CpuLoad1 = loadInfo.Load1
	itemModel.CpuLoad5 = loadInfo.Load5
	itemModel.CpuLoad15 = loadInfo.Load15
	itemModel.LoadUsage = loadInfo.Load1 / (float64(cpuCount*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	itemModel.Memory = memoryInfo.UsedPercent
	topMem := loadTopMem()
	if len(topMem) != 0 {
		topMemItem, err := json.Marshal(topMem)
		if err == nil {
			itemModel.TopMem = string(topMemItem)
		}
	}

	if err := monitorRepo.CreateMonitorBase(itemModel); err != nil {
		global.LOG.Errorf("Insert basic monitoring data failed, err: %v", err)
	}

	m.loadDiskIO()
	m.loadNetIO()

	MonitorStoreDays, err := settingRepo.Get(settingRepo.WithByKey("MonitorStoreDays"))
	if err != nil {
		return
	}
	storeDays, _ := strconv.Atoi(MonitorStoreDays.Value)
	timeForDelete := time.Now().AddDate(0, 0, -storeDays)
	_ = monitorRepo.DelMonitorBase(timeForDelete)
	_ = monitorRepo.DelMonitorIO(timeForDelete)
	_ = monitorRepo.DelMonitorNet(timeForDelete)
}

func (m *MonitorService) loadDiskIO() {
	ioStat, _ := disk.IOCounters()
	var diskIOList []disk.IOCountersStat
	var ioStatAll disk.IOCountersStat
	for _, io := range ioStat {
		ioStatAll.Name = "all"
		ioStatAll.ReadBytes += io.ReadBytes
		ioStatAll.WriteBytes += io.WriteBytes
		ioStatAll.ReadTime += io.ReadTime
		ioStatAll.WriteTime += io.WriteTime
		ioStatAll.WriteCount += io.WriteCount
		ioStatAll.ReadCount += io.ReadCount
		diskIOList = append(diskIOList, io)
	}
	diskIOList = append(diskIOList, ioStatAll)
	m.DiskIO <- diskIOList
}

func (m *MonitorService) loadNetIO() {
	netStat, _ := net.IOCounters(true)
	netStatAll, _ := net.IOCounters(false)
	var netList []net.IOCountersStat
	netList = append(netList, netStat...)
	netList = append(netList, netStatAll...)
	m.NetIO <- netList
}

func (m *MonitorService) saveIODataToDB(ctx context.Context, interval float64) {
	defer close(m.DiskIO)
	for {
		select {
		case <-ctx.Done():
			return
		case ioStat := <-m.DiskIO:
			select {
			case <-ctx.Done():
				return
			case ioStat2 := <-m.DiskIO:
				var ioList []model.MonitorIO
				for _, io2 := range ioStat2 {
					for _, io1 := range ioStat {
						if io2.Name == io1.Name {
							var itemIO model.MonitorIO
							itemIO.Name = io1.Name
							if io2.ReadBytes != 0 && io1.ReadBytes != 0 && io2.ReadBytes > io1.ReadBytes {
								itemIO.Read = uint64(float64(io2.ReadBytes-io1.ReadBytes) / interval)
							}
							if io2.WriteBytes != 0 && io1.WriteBytes != 0 && io2.WriteBytes > io1.WriteBytes {
								itemIO.Write = uint64(float64(io2.WriteBytes-io1.WriteBytes) / interval)
							}

							if io2.ReadCount != 0 && io1.ReadCount != 0 && io2.ReadCount > io1.ReadCount {
								itemIO.Count = uint64(float64(io2.ReadCount-io1.ReadCount) / interval)
							}
							writeCount := uint64(0)
							if io2.WriteCount != 0 && io1.WriteCount != 0 && io2.WriteCount > io1.WriteCount {
								writeCount = uint64(float64(io2.WriteCount-io1.WriteCount) / interval)
							}
							if writeCount > itemIO.Count {
								itemIO.Count = writeCount
							}

							if io2.ReadTime != 0 && io1.ReadTime != 0 && io2.ReadTime > io1.ReadTime {
								itemIO.Time = uint64(float64(io2.ReadTime-io1.ReadTime) / interval)
							}
							writeTime := uint64(0)
							if io2.WriteTime != 0 && io1.WriteTime != 0 && io2.WriteTime > io1.WriteTime {
								writeTime = uint64(float64(io2.WriteTime-io1.WriteTime) / interval)
							}
							if writeTime > itemIO.Time {
								itemIO.Time = writeTime
							}
							ioList = append(ioList, itemIO)
							break
						}
					}
				}
				_ = monitorRepo.BatchCreateMonitorIO(ioList)
				m.DiskIO <- ioStat2
			}
		}
	}
}

func (m *MonitorService) saveNetDataToDB(ctx context.Context, interval float64) {
	defer close(m.NetIO)
	for {
		select {
		case <-ctx.Done():
			return
		case netStat := <-m.NetIO:
			select {
			case <-ctx.Done():
				return
			case netStat2 := <-m.NetIO:
				var netList []model.MonitorNetwork
				for _, net2 := range netStat2 {
					for _, net1 := range netStat {
						if net2.Name == net1.Name {
							var itemNet model.MonitorNetwork
							itemNet.Name = net1.Name

							if net2.BytesSent != 0 && net1.BytesSent != 0 && net2.BytesSent > net1.BytesSent {
								itemNet.Up = float64(net2.BytesSent-net1.BytesSent) / 1024 / interval
							}
							if net2.BytesRecv != 0 && net1.BytesRecv != 0 && net2.BytesRecv > net1.BytesRecv {
								itemNet.Down = float64(net2.BytesRecv-net1.BytesRecv) / 1024 / interval
							}
							netList = append(netList, itemNet)
							break
						}
					}
				}

				_ = monitorRepo.BatchCreateMonitorNet(netList)
				m.NetIO <- netStat2
			}
		}
	}
}

func loadTopCPU() []dto.Process {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	top5 := make([]dto.Process, 0, 5)
	for _, p := range processes {
		percent, err := p.CPUPercent()
		if err != nil {
			continue
		}
		minIndex := 0
		if len(top5) >= 5 {
			minCPU := top5[0].Percent
			for i := 1; i < len(top5); i++ {
				if top5[i].Percent < minCPU {
					minCPU = top5[i].Percent
					minIndex = i
				}
			}
			if percent < minCPU {
				continue
			}
		}
		name, err := p.Name()
		if err != nil {
			name = "undefined"
		}
		cmd, err := p.Cmdline()
		if err != nil {
			cmd = "undefined"
		}
		user, err := p.Username()
		if err != nil {
			user = "undefined"
		}
		if len(top5) == 5 {
			top5[minIndex] = dto.Process{Percent: percent, Pid: p.Pid, User: user, Name: name, Cmd: cmd}
		} else {
			top5 = append(top5, dto.Process{Percent: percent, Pid: p.Pid, User: user, Name: name, Cmd: cmd})
		}
	}
	sort.Slice(top5, func(i, j int) bool {
		return top5[i].Percent > top5[j].Percent
	})

	return top5
}

func loadTopMem() []dto.Process {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	top5 := make([]dto.Process, 0, 5)
	for _, p := range processes {
		stat, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		memItem := stat.RSS
		minIndex := 0
		if len(top5) >= 5 {
			min := top5[0].Memory
			for i := 1; i < len(top5); i++ {
				if top5[i].Memory < min {
					min = top5[i].Memory
					minIndex = i
				}
			}
			if memItem < min {
				continue
			}
		}
		name, err := p.Name()
		if err != nil {
			name = "undefined"
		}
		cmd, err := p.Cmdline()
		if err != nil {
			cmd = "undefined"
		}
		user, err := p.Username()
		if err != nil {
			user = "undefined"
		}
		percent, _ := p.MemoryPercent()
		if len(top5) == 5 {
			top5[minIndex] = dto.Process{Percent: float64(percent), Pid: p.Pid, User: user, Name: name, Cmd: cmd, Memory: memItem}
		} else {
			top5 = append(top5, dto.Process{Percent: float64(percent), Pid: p.Pid, User: user, Name: name, Cmd: cmd, Memory: memItem})
		}
	}

	sort.Slice(top5, func(i, j int) bool {
		return top5[i].Memory > top5[j].Memory
	})
	return top5
}

func StartMonitor(removeBefore bool, interval string) error {
	if removeBefore {
		monitorCancel()
		global.Cron.Remove(cron.EntryID(global.MonitorCronID))
	}
	intervalItem, err := strconv.Atoi(interval)
	if err != nil {
		return err
	}

	service := NewIMonitorService()
	ctx, cancel := context.WithCancel(context.Background())
	monitorCancel = cancel
	now := time.Now()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute)
	time.AfterFunc(time.Until(nextMinute), func() {
		monitorID, err := global.Cron.AddJob(fmt.Sprintf("@every %ss", interval), service)
		if err != nil {
			return
		}
		global.MonitorCronID = monitorID
	})

	service.Run()

	go service.saveIODataToDB(ctx, float64(intervalItem))
	go service.saveNetDataToDB(ctx, float64(intervalItem))

	return nil
}

func saveGPUDataToDB() {
	exist, client := gpu.New()
	if !exist {
		return
	}
	gpuInfo, err := client.LoadGpuInfo()
	if err != nil {
		return
	}
	var list []model.MonitorGPU
	for _, gpuItem := range gpuInfo.GPUs {
		item := model.MonitorGPU{
			ProductName:   fmt.Sprintf("%d - %s", gpuItem.Index, gpuItem.ProductName),
			GPUUtil:       loadGPUInfoFloat(gpuItem.GPUUtil),
			Temperature:   loadGPUInfoFloat(gpuItem.Temperature),
			PowerDraw:     loadGPUInfoFloat(gpuItem.PowerDraw),
			MaxPowerLimit: loadGPUInfoFloat(gpuItem.MaxPowerLimit),
			MemUsed:       loadGPUInfoFloat(gpuItem.MemUsed),
			MemTotal:      loadGPUInfoFloat(gpuItem.MemTotal),
			FanSpeed:      loadGPUInfoInt(gpuItem.FanSpeed),
		}
		process, _ := json.Marshal(gpuItem.Processes)
		if len(process) != 0 {
			item.Processes = string(process)
		}
		list = append(list, item)
	}
	if err := repo.NewIMonitorRepo().BatchCreateMonitorGPU(list); err != nil {
		global.LOG.Errorf("batch create gpu monitor data failed, err: %v", err)
		return
	}
}
func saveXPUDataToDB() {
	exist, client := xpu.New()
	if !exist {
		return
	}
	xpuInfo, err := client.LoadGpuInfo()
	if err != nil {
		return
	}
	var list []model.MonitorGPU
	for _, xpuItem := range xpuInfo.Xpu {
		item := model.MonitorGPU{
			ProductName: fmt.Sprintf("%d - %s", xpuItem.Basic.DeviceID, xpuItem.Basic.DeviceName),
			Temperature: loadGPUInfoFloat(xpuItem.Stats.Temperature),
			PowerDraw:   loadGPUInfoFloat(xpuItem.Stats.Power),
			MemUsed:     loadGPUInfoFloat(xpuItem.Stats.MemoryUsed),
			MemTotal:    loadGPUInfoFloat(xpuItem.Basic.Memory),
		}
		if len(xpuItem.Processes) != 0 {
			var processItem []dto.GPUProcess
			for _, ps := range xpuItem.Processes {
				processItem = append(processItem, dto.GPUProcess{
					Pid:         fmt.Sprintf("%v", ps.PID),
					Type:        ps.SHR,
					ProcessName: ps.Command,
					UsedMemory:  ps.Memory,
				})
			}
			process, _ := json.Marshal(processItem)
			if len(process) != 0 {
				item.Processes = string(process)
			}
		}
		list = append(list, item)
	}
	if err := repo.NewIMonitorRepo().BatchCreateMonitorGPU(list); err != nil {
		global.LOG.Errorf("batch create gpu monitor data failed, err: %v", err)
		return
	}
}
func loadGPUInfoInt(val string) int {
	val = strings.TrimSuffix(val, "%")
	val = strings.TrimSpace(val)
	data, _ := strconv.Atoi(val)
	return data
}
func loadGPUInfoFloat(val string) float64 {
	val = strings.TrimSpace(val)
	suffixes := []string{"W", "MB", "MiB", "°C", "C", "%"}
	for _, suffix := range suffixes {
		val = strings.TrimSuffix(val, suffix)
	}
	val = strings.TrimSpace(val)
	data, _ := strconv.ParseFloat(val, 64)
	return data
}
