package service

import (
	"encoding/json"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type DashboardService struct{}

type IDashboardService interface {
	LoadBaseInfo(ioOption string, netOption string) (*dto.DashboardBase, error)
	LoadCurrentInfo(ioOption string, netOption string) *dto.DashboardCurrent
}

func NewIDashboardService() IDashboardService {
	return &DashboardService{}
}
func (u *DashboardService) LoadBaseInfo(ioOption string, netOption string) (*dto.DashboardBase, error) {
	var baseInfo dto.DashboardBase
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}
	baseInfo.Hostname = hostInfo.Hostname
	baseInfo.OS = hostInfo.OS
	baseInfo.Platform = hostInfo.Platform
	baseInfo.PlatformFamily = hostInfo.PlatformFamily
	baseInfo.PlatformVersion = hostInfo.PlatformVersion
	baseInfo.KernelArch = hostInfo.KernelArch
	baseInfo.KernelVersion = hostInfo.KernelVersion
	ss, _ := json.Marshal(hostInfo)
	baseInfo.VirtualizationSystem = string(ss)

	apps, err := appRepo.GetBy()
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		switch app.Key {
		case "dateease":
			baseInfo.DateeaseID = app.ID
		case "halo":
			baseInfo.HaloID = app.ID
		case "metersphere":
			baseInfo.MeterSphereID = app.ID
		case "jumpserver":
			baseInfo.JumpServerID = app.ID
		case "kubeoperator":
			baseInfo.KubeoperatorID = app.ID
		case "kubepi":
			baseInfo.KubepiID = app.ID
		}
	}

	appInstall, err := appInstallRepo.GetBy()
	if err != nil {
		return nil, err
	}
	baseInfo.AppInstalldNumber = len(appInstall)
	dbs, err := mysqlRepo.List()
	if err != nil {
		return nil, err
	}
	baseInfo.DatabaseNumber = len(dbs)
	website, err := websiteRepo.GetBy()
	if err != nil {
		return nil, err
	}
	baseInfo.WebsiteNumber = len(website)
	cornjobs, err := cronjobRepo.List()
	if err != nil {
		return nil, err
	}
	baseInfo.CronjobNumber = len(cornjobs)

	cpuInfo, err := cpu.Info()
	if err == nil {
		baseInfo.CPUModelName = cpuInfo[0].ModelName
	}

	baseInfo.CPUCores, _ = cpu.Counts(false)
	baseInfo.CPULogicalCores, _ = cpu.Counts(true)

	baseInfo.CurrentInfo = *u.LoadCurrentInfo(ioOption, netOption)
	return &baseInfo, nil
}

func (u *DashboardService) LoadCurrentInfo(ioOption string, netOption string) *dto.DashboardCurrent {
	var currentInfo dto.DashboardCurrent
	hostInfo, _ := host.Info()
	currentInfo.Uptime = hostInfo.Uptime
	currentInfo.TimeSinceUptime = time.Now().Add(-time.Duration(hostInfo.Uptime) * time.Second).Format("2006-01-02 15:04:05")
	currentInfo.Procs = hostInfo.Procs

	currentInfo.CPUTotal, _ = cpu.Counts(true)
	totalPercent, _ := cpu.Percent(0, false)
	if len(totalPercent) == 1 {
		currentInfo.CPUUsedPercent = totalPercent[0]
		currentInfo.CPUUsed = currentInfo.CPUUsedPercent * 0.01 * float64(currentInfo.CPUTotal)
	}
	currentInfo.CPUPercent, _ = cpu.Percent(0, true)

	loadInfo, _ := load.Avg()
	currentInfo.Load1 = loadInfo.Load1
	currentInfo.Load5 = loadInfo.Load5
	currentInfo.Load15 = loadInfo.Load15
	currentInfo.LoadUsagePercent = loadInfo.Load1 / (float64(currentInfo.CPUTotal*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	currentInfo.MemoryTotal = memoryInfo.Total
	currentInfo.MemoryAvailable = memoryInfo.Available
	currentInfo.MemoryUsed = memoryInfo.Used
	currentInfo.MemoryUsedPercent = memoryInfo.UsedPercent

	state, _ := disk.Usage("/")
	currentInfo.Total = state.Total
	currentInfo.Free = state.Free
	currentInfo.Used = state.Used
	currentInfo.UsedPercent = state.UsedPercent
	currentInfo.InodesTotal = state.InodesTotal
	currentInfo.InodesUsed = state.InodesUsed
	currentInfo.InodesFree = state.InodesFree
	currentInfo.InodesUsedPercent = state.InodesUsedPercent

	if ioOption == "all" {
		diskInfo, _ := disk.IOCounters()
		for _, state := range diskInfo {
			currentInfo.IOReadBytes += state.ReadBytes
			currentInfo.IOWriteBytes += state.WriteBytes
			currentInfo.IOCount += (state.ReadCount + state.WriteCount)
			currentInfo.IOTime += state.ReadTime / 1000 / 1000
			if state.WriteTime > state.ReadTime {
				currentInfo.IOTime += state.WriteTime / 1000 / 1000
			}
		}
	} else {
		diskInfo, _ := disk.IOCounters(ioOption)
		for _, state := range diskInfo {
			currentInfo.IOReadBytes += state.ReadBytes
			currentInfo.IOWriteBytes += state.WriteBytes
			currentInfo.IOTime += state.ReadTime / 1000 / 1000
			if state.WriteTime > state.ReadTime {
				currentInfo.IOTime += state.WriteTime / 1000 / 1000
			}
		}
	}

	if netOption == "all" {
		netInfo, _ := net.IOCounters(false)
		if len(netInfo) != 0 {
			currentInfo.NetBytesSent = netInfo[0].BytesSent
			currentInfo.NetBytesRecv = netInfo[0].BytesRecv
		}
	} else {
		netInfo, _ := net.IOCounters(true)
		for _, state := range netInfo {
			if state.Name == netOption {
				currentInfo.NetBytesSent = state.BytesSent
				currentInfo.NetBytesRecv = state.BytesRecv
			}
		}
	}

	currentInfo.ShotTime = time.Now()
	return &currentInfo
}
