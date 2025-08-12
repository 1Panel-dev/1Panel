package service

import (
	"context"
	"encoding/json"
	"fmt"
	network "net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/copier"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type DashboardService struct{}

type IDashboardService interface {
	LoadOsInfo() (*dto.OsInfo, error)
	LoadBaseInfo(ioOption string, netOption string) (*dto.DashboardBase, error)
	LoadCurrentInfoForNode() *dto.NodeCurrent
	LoadCurrentInfo(ioOption string, netOption string) *dto.DashboardCurrent

	LoadAppLauncher(ctx *gin.Context) ([]dto.AppLauncher, error)
	ChangeShow(req dto.SettingUpdate) error
	ListLauncherOption(filter string) ([]dto.LauncherOption, error)
	Restart(operation string) error
}

func NewIDashboardService() IDashboardService {
	return &DashboardService{}
}

func (u *DashboardService) Restart(operation string) error {
	if operation != "1panel" && operation != "system" && operation != "1panel-agent" {
		return fmt.Errorf("handle restart operation %s failed, err: nonsupport such operation", operation)
	}
	itemCmd := fmt.Sprintf("%s systemctl restart 1panel-agent.service && %s systemctl restart 1panel-core.service", cmd.SudoHandleCmd(), cmd.SudoHandleCmd())
	if operation == "system" {
		itemCmd = fmt.Sprintf("%s reboot", cmd.SudoHandleCmd())
	}
	if operation == "1panel-agent" {
		itemCmd = fmt.Sprintf("%s systemctl restart 1panel-agent.service", cmd.SudoHandleCmd())
	}
	go func() {
		stdout, err := cmd.RunDefaultWithStdoutBashC(itemCmd)
		if err != nil {
			global.LOG.Errorf("handle %s failed, err: %v", itemCmd, stdout)
		}
	}()
	return nil
}

func (u *DashboardService) LoadOsInfo() (*dto.OsInfo, error) {
	var baseInfo dto.OsInfo
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}
	baseInfo.OS = hostInfo.OS
	baseInfo.Platform = hostInfo.Platform
	baseInfo.PlatformFamily = hostInfo.PlatformFamily
	baseInfo.KernelArch = hostInfo.KernelArch
	baseInfo.KernelVersion = hostInfo.KernelVersion

	diskInfo, err := disk.Usage(global.Dir.BaseDir)
	if err == nil {
		baseInfo.DiskSize = int64(diskInfo.Free)
	}

	if baseInfo.KernelArch == "armv7l" {
		baseInfo.KernelArch = "armv7"
	}
	if baseInfo.KernelArch == "x86_64" {
		baseInfo.KernelArch = "amd64"
	}
	return &baseInfo, nil
}

func (u *DashboardService) LoadCurrentInfoForNode() *dto.NodeCurrent {
	var currentInfo dto.NodeCurrent

	currentInfo.CPUTotal, _ = cpu.Counts(true)
	totalPercent, _ := cpu.Percent(100*time.Millisecond, false)
	if len(totalPercent) == 1 {
		currentInfo.CPUUsedPercent = totalPercent[0]
		currentInfo.CPUUsed = currentInfo.CPUUsedPercent * 0.01 * float64(currentInfo.CPUTotal)
	}

	loadInfo, _ := load.Avg()
	currentInfo.Load1 = loadInfo.Load1
	currentInfo.Load5 = loadInfo.Load5
	currentInfo.Load15 = loadInfo.Load15
	currentInfo.LoadUsagePercent = loadInfo.Load1 / (float64(currentInfo.CPUTotal*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	currentInfo.MemoryTotal = memoryInfo.Total
	currentInfo.MemoryAvailable = memoryInfo.Available
	currentInfo.MemoryUsed = memoryInfo.Used + memoryInfo.Shared
	currentInfo.MemoryUsedPercent = memoryInfo.UsedPercent

	swapInfo, _ := mem.SwapMemory()
	currentInfo.SwapMemoryTotal = swapInfo.Total
	currentInfo.SwapMemoryAvailable = swapInfo.Free
	currentInfo.SwapMemoryUsed = swapInfo.Used
	currentInfo.SwapMemoryUsedPercent = swapInfo.UsedPercent

	return &currentInfo
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
	baseInfo.IpV4Addr = GetOutboundIP()
	httpProxy := os.Getenv("http_proxy")
	if httpProxy == "" {
		httpProxy = os.Getenv("HTTP_PROXY")
	}
	if httpProxy != "" {
		baseInfo.SystemProxy = httpProxy
	}
	baseInfo.SystemProxy = "noProxy"

	appInstall, err := appInstallRepo.ListBy(context.Background())
	if err != nil {
		return nil, err
	}
	baseInfo.AppInstalledNumber = len(appInstall)
	postgresqlDbs, err := postgresqlRepo.List()
	if err != nil {
		return nil, err
	}
	mysqlDbs, err := mysqlRepo.List()
	if err != nil {
		return nil, err
	}
	baseInfo.DatabaseNumber = len(mysqlDbs) + len(postgresqlDbs)
	website, err := websiteRepo.GetBy()
	if err != nil {
		return nil, err
	}
	baseInfo.WebsiteNumber = len(website)
	cronjobs, err := cronjobRepo.List()
	if err != nil {
		return nil, err
	}
	baseInfo.CronjobNumber = len(cronjobs)

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
	currentInfo.TimeSinceUptime = time.Now().Add(-time.Duration(hostInfo.Uptime) * time.Second).Format(constant.DateTimeLayout)
	currentInfo.Procs = hostInfo.Procs

	currentInfo.CPUTotal, _ = cpu.Counts(true)
	totalPercent, _ := cpu.Percent(100*time.Millisecond, false)
	if len(totalPercent) == 1 {
		currentInfo.CPUUsedPercent = totalPercent[0]
		currentInfo.CPUUsed = currentInfo.CPUUsedPercent * 0.01 * float64(currentInfo.CPUTotal)
	}
	currentInfo.CPUPercent, _ = cpu.Percent(100*time.Millisecond, true)

	loadInfo, _ := load.Avg()
	currentInfo.Load1 = loadInfo.Load1
	currentInfo.Load5 = loadInfo.Load5
	currentInfo.Load15 = loadInfo.Load15
	currentInfo.LoadUsagePercent = loadInfo.Load1 / (float64(currentInfo.CPUTotal*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	currentInfo.MemoryTotal = memoryInfo.Total
	currentInfo.MemoryUsed = memoryInfo.Used + memoryInfo.Shared
	currentInfo.MemoryFree = memoryInfo.Free
	currentInfo.MemoryCache = memoryInfo.Cached + memoryInfo.Buffers
	currentInfo.MemoryShard = memoryInfo.Shared
	currentInfo.MemoryAvailable = memoryInfo.Available
	currentInfo.MemoryUsedPercent = memoryInfo.UsedPercent

	swapInfo, _ := mem.SwapMemory()
	currentInfo.SwapMemoryTotal = swapInfo.Total
	currentInfo.SwapMemoryAvailable = swapInfo.Free
	currentInfo.SwapMemoryUsed = swapInfo.Used
	currentInfo.SwapMemoryUsedPercent = swapInfo.UsedPercent

	currentInfo.DiskData = loadDiskInfo()
	currentInfo.GPUData = loadGPUInfo()
	currentInfo.XPUData = loadXpuInfo()

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

func (u *DashboardService) LoadAppLauncher(ctx *gin.Context) ([]dto.AppLauncher, error) {
	var (
		data          []dto.AppLauncher
		recommendList []dto.AppLauncher
	)
	appInstalls, err := appInstallRepo.ListBy(context.Background())
	if err != nil {
		return data, err
	}
	apps, err := appRepo.GetBy()
	if err != nil {
		return data, err
	}

	showList, _ := launcherRepo.ListName()
	defaultList := []string{"openresty", "mysql", "halo", "redis", "maxkb", "wordpress"}
	allList := common.RemoveRepeatStr(append(defaultList, showList...))
	for _, showItem := range allList {
		var itemData dto.AppLauncher
		for _, app := range apps {
			if showItem == app.Key {
				itemData.Key = app.Key
				itemData.Type = app.Type
				itemData.Name = app.Name
				itemData.Icon = app.Icon
				itemData.Limit = app.Limit
				itemData.Recommend = app.Recommend
				itemData.Description = app.GetDescription(ctx)
				break
			}
		}
		if len(itemData.Icon) == 0 {
			continue
		}
		for _, install := range appInstalls {
			if install.App.Key == showItem {
				itemData.IsInstall = true
				itemData.Detail = append(itemData.Detail, dto.InstallDetail{
					InstallID: install.ID,
					DetailID:  install.AppDetailId,
					Name:      install.Name,
					Version:   install.Version,
					Status:    install.Status,
					Path:      install.GetPath(),
					WebUI:     install.WebUI,
					HttpPort:  install.HttpPort,
					HttpsPort: install.HttpsPort,
				})
			}
		}
		if ArryContains(defaultList, showItem) && len(itemData.Detail) == 0 {
			itemData.IsRecommend = true
			recommendList = append(recommendList, itemData)
			continue
		}
		if !ArryContains(showList, showItem) && len(itemData.Detail) != 0 {
			continue
		}
		data = append(data, itemData)
	}

	sort.Slice(recommendList, func(i, j int) bool {
		return recommendList[i].Recommend < recommendList[j].Recommend
	})
	sort.Slice(data, func(i, j int) bool {
		return data[i].Name < data[j].Name
	})
	data = append(data, recommendList...)
	return data, nil
}

func (u *DashboardService) ChangeShow(req dto.SettingUpdate) error {
	launcher, _ := launcherRepo.Get(repo.WithByKey(req.Key))
	if req.Value == constant.StatusEnable && launcher.ID == 0 {
		if err := launcherRepo.Create(&model.AppLauncher{Key: req.Key}); err != nil {
			return err
		}
	}
	if req.Value == constant.StatusDisable && launcher.ID != 0 {
		if err := launcherRepo.Delete(repo.WithByKey(req.Key)); err != nil {
			return err
		}
	}
	return nil
}

func (u *DashboardService) ListLauncherOption(filter string) ([]dto.LauncherOption, error) {
	showList, _ := launcherRepo.ListName()
	var data []dto.LauncherOption
	optionMap := make(map[string]bool)
	appInstalls, err := appInstallRepo.ListBy(context.Background())
	if err != nil {
		return data, err
	}

	for _, install := range appInstalls {
		isShow := false
		for _, item := range showList {
			if install.App.Key == item {
				isShow = true
				break
			}
		}
		optionMap[install.App.Key] = isShow
	}
	for key, val := range optionMap {
		if len(filter) != 0 && !strings.Contains(key, filter) {
			continue
		}
		data = append(data, dto.LauncherOption{Key: key, IsShow: val})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Key < data[j].Key
	})
	return data, nil
}

type diskInfo struct {
	Type   string
	Mount  string
	Device string
}

func loadDiskInfo() []dto.DiskInfo {
	var datas []dto.DiskInfo
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashC("timeout 2 df -hT -P | awk 'NR>1 && !/tmpfs|snap\\/core|udev/ {print}'")
	if err != nil {
		cmdMgr2 := cmd.NewCommandMgr(cmd.WithTimeout(1 * time.Second))
		stdout, err = cmdMgr2.RunWithStdoutBashC("timeout 1 df -lhT -P | awk 'NR>1 && !/tmpfs|snap\\/core|udev/ {print}'")
		if err != nil {
			return datas
		}
	}
	lines := strings.Split(stdout, "\n")

	var mounts []diskInfo
	var excludes = []string{"/mnt/cdrom", "/boot", "/boot/efi", "/dev", "/dev/shm", "/run/lock", "/run", "/run/shm", "/run/user"}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		if strings.HasPrefix(fields[6], "/snap") || len(strings.Split(fields[6], "/")) > 10 {
			continue
		}
		if strings.TrimSpace(fields[1]) == "tmpfs" || strings.TrimSpace(fields[1]) == "overlay" {
			continue
		}
		if strings.Contains(fields[2], "K") {
			continue
		}
		if strings.Contains(fields[6], "docker") || strings.Contains(fields[6], "podman") || strings.Contains(fields[6], "containerd") || strings.HasPrefix(fields[6], "/var/lib/containers") {
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
		mounts = append(mounts, diskInfo{Type: fields[1], Device: fields[0], Mount: strings.Join(fields[6:], " ")})
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	wg.Add(len(mounts))
	for i := 0; i < len(mounts); i++ {
		go func(timeoutCh <-chan time.Time, mount diskInfo) {
			defer wg.Done()

			var itemData dto.DiskInfo
			itemData.Path = mount.Mount
			itemData.Type = mount.Type
			itemData.Device = mount.Device
			select {
			case <-timeoutCh:
				mu.Lock()
				datas = append(datas, itemData)
				mu.Unlock()
				global.LOG.Errorf("load disk info from %s failed, err: timeout", mount.Mount)
			default:
				state, err := disk.Usage(mount.Mount)
				if err != nil {
					mu.Lock()
					datas = append(datas, itemData)
					mu.Unlock()
					global.LOG.Errorf("load disk info from %s failed, err: %v", mount.Mount, err)
					return
				}
				itemData.Total = state.Total
				itemData.Free = state.Free
				itemData.Used = state.Used
				itemData.UsedPercent = state.UsedPercent
				itemData.InodesTotal = state.InodesTotal
				itemData.InodesUsed = state.InodesUsed
				itemData.InodesFree = state.InodesFree
				itemData.InodesUsedPercent = state.InodesUsedPercent
				mu.Lock()
				datas = append(datas, itemData)
				mu.Unlock()
			}
		}(time.After(5*time.Second), mounts[i])
	}
	wg.Wait()

	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Path < datas[j].Path
	})
	return datas
}

func loadGPUInfo() []dto.GPUInfo {
	ok, client := gpu.New()
	var list []interface{}
	if ok {
		info, err := client.LoadGpuInfo()
		if err != nil || len(info.GPUs) == 0 {
			return nil
		}
		for _, item := range info.GPUs {
			list = append(list, item)
		}
	}
	if len(list) == 0 {
		return nil
	}
	var data []dto.GPUInfo
	for _, gpu := range list {
		var dataItem dto.GPUInfo
		if err := copier.Copy(&dataItem, &gpu); err != nil {
			continue
		}
		dataItem.PowerUsage = dataItem.PowerDraw + " / " + dataItem.MaxPowerLimit
		dataItem.MemoryUsage = dataItem.MemUsed + " / " + dataItem.MemTotal
		data = append(data, dataItem)
	}
	return data
}

type AppLauncher struct {
	Key string `json:"key"`
}

func ArryContains(arr []string, element string) bool {
	for _, v := range arr {
		if v == element {
			return true
		}
	}
	return false
}

func loadXpuInfo() []dto.XPUInfo {
	var list []interface{}
	ok, xpuClient := xpu.New()
	if ok {
		xpus, err := xpuClient.LoadDashData()
		if err != nil || len(xpus) == 0 {
			return nil
		}
		for _, item := range xpus {
			list = append(list, item)
		}
	}
	if len(list) == 0 {
		return nil
	}
	var data []dto.XPUInfo
	for _, gpu := range list {
		var dataItem dto.XPUInfo
		if err := copier.Copy(&dataItem, &gpu); err != nil {
			continue
		}
		data = append(data, dataItem)
	}
	return data
}

func GetOutboundIP() string {
	conn, err := network.Dial("udp", "8.8.8.8:80")

	if err != nil {
		return "IPNotFound"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*network.UDPAddr)
	return localAddr.IP.String()
}
