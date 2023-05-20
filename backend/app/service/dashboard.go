package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
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

	appInstall, err := appInstallRepo.ListBy()
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

	currentInfo.DiskData = loadDiskInfo()

	if ioOption == "all" {
		diskInfo, _ := disk.IOCounters()
		for _, state := range diskInfo {
			currentInfo.IOReadBytes += state.ReadBytes
			currentInfo.IOWriteBytes += state.WriteBytes
			currentInfo.IOCount += (state.ReadCount + state.WriteCount)
			currentInfo.IOReadTime += state.ReadTime
			currentInfo.IOWriteTime += state.WriteTime
		}
	} else {
		diskInfo, _ := disk.IOCounters(ioOption)
		for _, state := range diskInfo {
			currentInfo.IOReadBytes += state.ReadBytes
			currentInfo.IOWriteBytes += state.WriteBytes
			currentInfo.IOCount += (state.ReadCount + state.WriteCount)
			currentInfo.IOReadTime += state.ReadTime
			currentInfo.IOWriteTime += state.WriteTime
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

type diskInfo struct {
	Type   string
	Mount  string
	Device string
}

func loadDiskInfo() []dto.DiskInfo {
	var datas []dto.DiskInfo
	stdout, err := cmd.Exec("df -hT -P|grep '/'|grep -v tmpfs|grep -v 'snap/core'|grep -v udev")
	if err != nil {
		return datas
	}
	lines := strings.Split(stdout, "\n")

	var mounts []diskInfo
	var excludes = []string{"/mnt/cdrom", "/boot", "/boot/efi", "/dev", "/dev/shm", "/run/lock", "/run", "/run/shm", "/run/user"}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		if fields[1] == "tmpfs" {
			continue
		}
		if strings.Contains(fields[2], "M") || strings.Contains(fields[2], "K") {
			continue
		}
		if strings.Contains(fields[6], "docker") {
			continue
		}
		isExclude := false
		for _, exclude := range excludes {
			if exclude == fields[6] {
				isExclude = true
			}
		}
		if isExclude {
			continue
		}
		mounts = append(mounts, diskInfo{Type: fields[1], Device: fields[0], Mount: fields[6]})
	}

	for i := 0; i < len(mounts); i++ {
		state, err := disk.Usage(mounts[i].Mount)
		if err != nil {
			continue
		}
		var itemData dto.DiskInfo
		itemData.Path = mounts[i].Mount
		itemData.Type = mounts[i].Type
		itemData.Device = mounts[i].Device
		itemData.Total = state.Total
		itemData.Free = state.Free
		itemData.Used = state.Used
		itemData.UsedPercent = state.UsedPercent
		itemData.InodesTotal = state.InodesTotal
		itemData.InodesUsed = state.InodesUsed
		itemData.InodesFree = state.InodesFree
		itemData.InodesUsedPercent = state.InodesUsedPercent
		datas = append(datas, itemData)
	}
	return datas
}
