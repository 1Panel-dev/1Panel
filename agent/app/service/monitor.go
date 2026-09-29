package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/accelerator"
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

var (
	monitorCancel context.CancelFunc
	hostSysPath   = loadHostSysPath()

	blockDevicePartitionCache sync.Map
)

type IMonitorService interface {
	Run()
	LoadMonitorData(req dto.MonitorSearch) ([]dto.MonitorData, error)
	LoadSetting() (*dto.MonitorSetting, error)
	UpdateSetting(key, value string) error
	CleanData() error

	LoadGPUOptions() dto.MonitorGPUOptions
	LoadGPUMonitorData(req dto.MonitorGPUSearch) (dto.MonitorGPUData, error)

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
	seen := make(map[string]bool)
	if exist, client := accelerator.New(); exist {
		snapshot, err := client.Collect(context.Background())
		if err != nil {
			global.LOG.Warnf("Load accelerator options failed: %v", err)
		} else {
			data = loadGPUOptions(snapshot)
			for _, item := range data.ChartHide {
				seen[item.DeviceID] = true
			}
		}
	}
	devices, err := monitorRepo.GetGPUDevices()
	if err != nil {
		global.LOG.Warnf("Load accelerator history options failed: %v", err)
		return data
	}
	for _, device := range devices {
		key := device.DeviceID
		if key == "" {
			key = "legacy:" + device.ProductName
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		data.ChartHide = append(data.ChartHide, dto.GPUChartHide{DeviceID: device.DeviceID, ProductName: device.ProductName, Type: device.DeviceType, Legacy: device.DeviceID == ""})
		data.Options = append(data.Options, device.ProductName)
	}
	return data
}

func (m *MonitorService) LoadGPUMonitorData(req dto.MonitorGPUSearch) (dto.MonitorGPUData, error) {
	var data dto.MonitorGPUData
	if req.StartTime.IsZero() || req.EndTime.IsZero() || !req.EndTime.After(req.StartTime) {
		return data, fmt.Errorf("invalid GPU history time range")
	}
	if req.DeviceID == "" && req.ProductName == "" {
		return data, fmt.Errorf("GPU history requires a device")
	}
	if req.Aggregation != "" && req.Aggregation != "avg" && req.Aggregation != "max" {
		return data, fmt.Errorf("invalid GPU history aggregation")
	}
	loc, err := time.LoadLocation(common.LoadTimeZoneByCmd())
	if err != nil {
		return data, err
	}
	req.StartTime, req.EndTime = req.StartTime.In(loc), req.EndTime.In(loc)
	opts := []repo.DBOption{repo.WithByCreatedAt(req.StartTime, req.EndTime), monitorRepo.WithByGPUDevice(req.DeviceID, req.ProductName, req.Legacy)}
	data.SampleCount, err = monitorRepo.CountGPU(opts...)
	if err != nil || data.SampleCount == 0 {
		return data, err
	}
	if data.SampleCount > 1200 {
		seconds := req.EndTime.Unix() - req.StartTime.Unix() + 1
		data.BucketSeconds = (seconds + 599) / 600
	}
	points, err := monitorRepo.GetGPUHistory(req.StartTime, data.BucketSeconds, req.Aggregation, opts...)
	if err != nil {
		return data, err
	}
	samples := make([]repo.GPUHistoryPoint, 0, len(points))
	if data.BucketSeconds > 0 {
		next := 0
		for bucket := int64(0); bucket <= (req.EndTime.Unix()-req.StartTime.Unix())/data.BucketSeconds; bucket++ {
			point := repo.GPUHistoryPoint{}
			if next < len(points) && points[next].Bucket == bucket {
				point = points[next]
				next++
			}
			point.CreatedAt = time.Unix(req.StartTime.Unix()+bucket*data.BucketSeconds, 0).In(loc)
			if bucket == 0 {
				point.CreatedAt = req.StartTime
			}
			samples = append(samples, point)
		}
	} else {
		for i, point := range points {
			if i > 0 && points[i-1].IntervalSeconds > 0 {
				interval := time.Duration(points[i-1].IntervalSeconds) * time.Second
				if point.CreatedAt.Sub(points[i-1].CreatedAt) > 2*interval {
					samples = append(samples, repo.GPUHistoryPoint{MonitorGPU: model.MonitorGPU{BaseModel: model.BaseModel{CreatedAt: points[i-1].CreatedAt.Add(interval)}}})
				}
			}
			samples = append(samples, point)
		}
	}
	for _, point := range samples {
		data.Date = append(data.Date, point.CreatedAt)
		data.MemoryActivity = append(data.MemoryActivity, point.MemoryActivity)
		data.EncoderUtil = append(data.EncoderUtil, point.EncoderUtil)
		data.DecoderUtil = append(data.DecoderUtil, point.DecoderUtil)
		data.JPEGUtil = append(data.JPEGUtil, point.JPEGUtil)
		data.OFAUtil = append(data.OFAUtil, point.OFAUtil)
		data.MediaUtil = append(data.MediaUtil, point.MediaUtil)
		data.ComputeUtil = append(data.ComputeUtil, point.ComputeUtil)
		data.CopyUtil = append(data.CopyUtil, point.CopyUtil)
		data.HotspotTemperature = append(data.HotspotTemperature, point.HotspotTemperature)
		data.FanRPM = append(data.FanRPM, point.FanRPM)
		data.AICPUUtil = append(data.AICPUUtil, point.AICPUUtil)
		data.CtrlCPUUtil = append(data.CtrlCPUUtil, point.CtrlCPUUtil)
		data.DDRUsed = append(data.DDRUsed, point.DDRUsed)
		data.DDRTotal = append(data.DDRTotal, point.DDRTotal)
		data.HBMUsed = append(data.HBMUsed, point.HBMUsed)
		data.HBMTotal = append(data.HBMTotal, point.HBMTotal)
		data.DDRBandwidth = append(data.DDRBandwidth, point.DDRBandwidth)
		data.HBMBandwidth = append(data.HBMBandwidth, point.HBMBandwidth)
		data.MemoryBandwidth = append(data.MemoryBandwidth, point.MemoryBandwidth)
		data.MediaFrequency = append(data.MediaFrequency, point.MediaFrequency)
		data.HugepagesUsed = append(data.HugepagesUsed, point.HugepagesUsed)
		data.HugepagesTotal = append(data.HugepagesTotal, point.HugepagesTotal)
		data.GPUValue = append(data.GPUValue, point.GPUUtil)
		data.TemperatureValue = append(data.TemperatureValue, point.Temperature)
		data.MemoryTemperatureValue = append(data.MemoryTemperatureValue, point.MemoryTemperature)
		data.PowerUsed = append(data.PowerUsed, point.PowerDraw)
		data.PowerTotal = append(data.PowerTotal, point.MaxPowerLimit)
		data.PowerPercent = append(data.PowerPercent, point.PowerPercent)
		data.MemoryPercent = append(data.MemoryPercent, point.MemoryPercent)
		data.MemoryTotal = append(data.MemoryTotal, point.MemTotal)
		data.MemoryUsed = append(data.MemoryUsed, point.MemUsed)
		data.SpeedValue = append(data.SpeedValue, point.FanSpeed)
		data.FrequencyValue = append(data.FrequencyValue, point.Frequency)
		data.MemoryFrequencyValue = append(data.MemoryFrequencyValue, point.MemoryFrequency)
		data.ProcessCount = append(data.ProcessCount, point.ProcessCount)
		var processes []dto.GPUProcess
		if data.BucketSeconds == 0 && point.ProcessCount != nil {
			_ = json.Unmarshal([]byte(point.Processes), &processes)
		}
		data.GPUProcesses = append(data.GPUProcesses, processes)
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
	m.saveGPUData()

	MonitorStoreDays, err := settingRepo.Get(settingRepo.WithByKey("MonitorStoreDays"))
	if err != nil {
		return
	}
	storeDays, _ := strconv.Atoi(MonitorStoreDays.Value)
	timeForDelete := time.Now().AddDate(0, 0, -storeDays)
	_ = monitorRepo.DelMonitorBase(timeForDelete)
	_ = monitorRepo.DelMonitorIO(timeForDelete)
	_ = monitorRepo.DelMonitorNet(timeForDelete)
	_ = monitorRepo.DelMonitorGPU(timeForDelete)
}

func (m *MonitorService) loadDiskIO() {
	ioStat, _ := disk.IOCounters()
	var diskIOList []disk.IOCountersStat
	for _, io := range ioStat {
		diskIOList = append(diskIOList, io)
	}
	diskIOList = append(diskIOList, sumDiskIOCounters(ioStat))
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

func loadGPUOptions(snapshot *accelerator.Snapshot) dto.MonitorGPUOptions {
	var data dto.MonitorGPUOptions
	hasGPUOrNPU := false
	hasXPU := false
	for _, item := range snapshot.Devices {
		if item.Kind == accelerator.KindXPU {
			hasXPU = true
		} else {
			hasGPUOrNPU = true
		}
	}
	switch {
	case hasGPUOrNPU && hasXPU:
		data.GPUType = "mixed"
	case hasXPU:
		data.GPUType = "xpu"
	case hasGPUOrNPU:
		data.GPUType = "gpu"
	}

	sort.Slice(snapshot.Devices, func(i, j int) bool {
		if snapshot.Devices[i].Kind != snapshot.Devices[j].Kind {
			return snapshot.Devices[i].Kind < snapshot.Devices[j].Kind
		}
		if snapshot.Devices[i].Vendor != snapshot.Devices[j].Vendor {
			return snapshot.Devices[i].Vendor < snapshot.Devices[j].Vendor
		}
		if snapshot.Devices[i].NPUIndex != snapshot.Devices[j].NPUIndex {
			return snapshot.Devices[i].NPUIndex < snapshot.Devices[j].NPUIndex
		}
		if snapshot.Devices[i].ChipIndex != snapshot.Devices[j].ChipIndex {
			return snapshot.Devices[i].ChipIndex < snapshot.Devices[j].ChipIndex
		}
		return snapshot.Devices[i].Index < snapshot.Devices[j].Index
	})
	for _, item := range snapshot.Devices {
		chartHide := dto.GPUChartHide{
			DeviceID:    item.ID,
			ProductName: item.Label,
			Type:        string(item.Kind),
		}
		data.ChartHide = append(data.ChartHide, chartHide)
		data.Options = append(data.Options, chartHide.ProductName)
	}
	return data
}

func (m *MonitorService) saveGPUData() {
	status, err := settingRepo.GetValueByKey("MonitorStatus")
	if err != nil {
		global.LOG.Errorf("load monitor status failed: %v", err)
		return
	}
	if status != constant.StatusEnable {
		return
	}
	exist, client := accelerator.New()
	if !exist {
		return
	}
	snapshot, err := client.Collect(context.Background())
	if err != nil {
		global.LOG.Errorf("load accelerator monitor data failed, err: %v", err)
		return
	}
	if warning := snapshot.Warning(); warning != nil {
		global.LOG.Warnf("load accelerator monitor data partially failed, err: %v", warning)
	}
	intervalSeconds := 0
	if setting, err := settingRepo.Get(settingRepo.WithByKey("MonitorInterval")); err == nil {
		intervalSeconds, _ = strconv.Atoi(setting.Value)
	}
	list := make([]model.MonitorGPU, 0, len(snapshot.Devices))
	for _, device := range snapshot.Devices {
		item := newMonitorGPU(device)
		item.CreatedAt = snapshot.Info.CollectedAt
		item.IntervalSeconds = intervalSeconds
		list = append(list, item)
	}
	if err := monitorRepo.BatchCreateMonitorGPU(list); err != nil {
		global.LOG.Errorf("batch create accelerator monitor data failed, err: %v", err)
	}
}

func newMonitorGPU(device accelerator.Device) model.MonitorGPU {
	item := model.MonitorGPU{
		MemoryUtil:         device.Metrics.MemoryUtil.Value,
		MemoryActivity:     device.Metrics.MemoryActivity.Value,
		EncoderUtil:        device.Metrics.EncoderUtil.Value,
		DecoderUtil:        device.Metrics.DecoderUtil.Value,
		JPEGUtil:           device.Metrics.JPEGUtil.Value,
		OFAUtil:            device.Metrics.OFAUtil.Value,
		MediaUtil:          device.Metrics.MediaUtil.Value,
		ComputeUtil:        device.Metrics.ComputeUtil.Value,
		CopyUtil:           device.Metrics.CopyUtil.Value,
		HotspotTemperature: device.Metrics.HotspotTemperature.Value,
		FanRPM:             device.Metrics.FanRPM.Value,
		AICPUUtil:          device.Metrics.AICPUUtil.Value,
		CtrlCPUUtil:        device.Metrics.CtrlCPUUtil.Value,
		DDRUsed:            device.Metrics.DDRUsed.Value,
		DDRTotal:           device.Metrics.DDRTotal.Value,
		HBMUsed:            device.Metrics.HBMUsed.Value,
		HBMTotal:           device.Metrics.HBMTotal.Value,
		DDRBandwidth:       device.Metrics.DDRBandwidth.Value,
		HBMBandwidth:       device.Metrics.HBMBandwidth.Value,
		MemoryBandwidth:    device.Metrics.MemoryBandwidth.Value,
		MediaFrequency:     device.Metrics.MediaFrequency.Value,
		HugepagesUsed:      device.Metrics.HugepagesUsed.Value,
		HugepagesTotal:     device.Metrics.HugepagesTotal.Value,

		ProductName:       device.Label,
		DeviceID:          device.ID,
		DeviceType:        string(device.Kind),
		ProcessStatus:     device.ProcessStatus,
		Frequency:         device.Metrics.Frequency.Value,
		MemoryFrequency:   device.Metrics.MemoryFrequency.Value,
		MemoryTemperature: device.Metrics.MemoryTemperature.Value,
		GPUUtil:           device.Metrics.Utilization.Value,
		Temperature:       device.Metrics.Temperature.Value,
		PowerDraw:         device.Metrics.Power.Value,
		MaxPowerLimit:     device.Metrics.PowerLimit.Value,
		MemUsed:           device.Metrics.MemoryUsed.Value,
		MemTotal:          device.Metrics.MemoryTotal.Value,
		FanSpeed:          device.Metrics.FanSpeed.Value,
	}
	if device.ProcessStatus != "ok" {
		return item
	}
	processes := make([]dto.GPUProcess, 0, len(device.Processes))
	for _, process := range device.Processes {
		processes = append(processes, dto.GPUProcess{
			Pid:         process.PID,
			Type:        process.Type,
			ProcessName: process.Name,
			UsedMemory:  process.Memory,
		})
	}
	processData, err := json.Marshal(processes)
	if err == nil {
		item.Processes = string(processData)
	}
	return item
}

func sumDiskIOCounters(ioStats map[string]disk.IOCountersStat) disk.IOCountersStat {
	total := disk.IOCountersStat{Name: "all"}
	for name, stat := range ioStats {
		if isBlockDevicePartition(name) {
			continue
		}

		total.ReadCount += stat.ReadCount
		total.MergedReadCount += stat.MergedReadCount
		total.WriteCount += stat.WriteCount
		total.MergedWriteCount += stat.MergedWriteCount
		total.ReadBytes += stat.ReadBytes
		total.WriteBytes += stat.WriteBytes
		total.ReadTime += stat.ReadTime
		total.WriteTime += stat.WriteTime
		total.IopsInProgress += stat.IopsInProgress
		total.IoTime += stat.IoTime
		total.WeightedIO += stat.WeightedIO
	}
	return total
}

func isBlockDevicePartition(name string) bool {
	deviceName := filepath.Base(name)
	if cached, ok := blockDevicePartitionCache.Load(deviceName); ok {
		return cached.(bool)
	}

	_, err := os.Stat(filepath.Join(hostSysPath, "class", "block", deviceName, "partition"))
	isPartition := err == nil
	actual, _ := blockDevicePartitionCache.LoadOrStore(deviceName, isPartition)
	return actual.(bool)
}

func loadHostSysPath() string {
	hostSys := os.Getenv("HOST_SYS")
	if hostSys == "" {
		return "/sys"
	}
	return hostSys
}
