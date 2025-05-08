package dto

import "time"

type DashboardBase struct {
	WebsiteNumber      int `json:"websiteNumber"`
	DatabaseNumber     int `json:"databaseNumber"`
	CronjobNumber      int `json:"cronjobNumber"`
	AppInstalledNumber int `json:"appInstalledNumber"`

	Hostname             string `json:"hostname"`
	OS                   string `json:"os"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platformFamily"`
	PlatformVersion      string `json:"platformVersion"`
	KernelArch           string `json:"kernelArch"`
	KernelVersion        string `json:"kernelVersion"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	IpV4Addr             string `json:"ipV4Addr"`
	SystemProxy          string `json:"systemProxy"`

	CPUCores        int    `json:"cpuCores"`
	CPULogicalCores int    `json:"cpuLogicalCores"`
	CPUModelName    string `json:"cpuModelName"`

	CurrentInfo DashboardCurrent `json:"currentInfo"`
}

type OsInfo struct {
	OS             string `json:"os"`
	Platform       string `json:"platform"`
	PlatformFamily string `json:"platformFamily"`
	KernelArch     string `json:"kernelArch"`
	KernelVersion  string `json:"kernelVersion"`

	DiskSize int64 `json:"diskSize"`
}

type NodeCurrent struct {
	Load1            float64 `json:"load1"`
	Load5            float64 `json:"load5"`
	Load15           float64 `json:"load15"`
	LoadUsagePercent float64 `json:"loadUsagePercent"`

	CPUUsedPercent float64 `json:"cpuUsedPercent"`
	CPUUsed        float64 `json:"cpuUsed"`
	CPUTotal       int     `json:"cpuTotal"`

	MemoryTotal       uint64  `json:"memoryTotal"`
	MemoryAvailable   uint64  `json:"memoryAvailable"`
	MemoryUsed        uint64  `json:"memoryUsed"`
	MemoryUsedPercent float64 `json:"memoryUsedPercent"`

	SwapMemoryTotal       uint64  `json:"swapMemoryTotal"`
	SwapMemoryAvailable   uint64  `json:"swapMemoryAvailable"`
	SwapMemoryUsed        uint64  `json:"swapMemoryUsed"`
	SwapMemoryUsedPercent float64 `json:"swapMemoryUsedPercent"`
}

type DashboardCurrent struct {
	Uptime          uint64 `json:"uptime"`
	TimeSinceUptime string `json:"timeSinceUptime"`

	Procs uint64 `json:"procs"`

	Load1            float64 `json:"load1"`
	Load5            float64 `json:"load5"`
	Load15           float64 `json:"load15"`
	LoadUsagePercent float64 `json:"loadUsagePercent"`

	CPUPercent     []float64 `json:"cpuPercent"`
	CPUUsedPercent float64   `json:"cpuUsedPercent"`
	CPUUsed        float64   `json:"cpuUsed"`
	CPUTotal       int       `json:"cpuTotal"`

	MemoryTotal       uint64  `json:"memoryTotal"`
	MemoryAvailable   uint64  `json:"memoryAvailable"`
	MemoryUsed        uint64  `json:"memoryUsed"`
	MemoryUsedPercent float64 `json:"memoryUsedPercent"`

	SwapMemoryTotal       uint64  `json:"swapMemoryTotal"`
	SwapMemoryAvailable   uint64  `json:"swapMemoryAvailable"`
	SwapMemoryUsed        uint64  `json:"swapMemoryUsed"`
	SwapMemoryUsedPercent float64 `json:"swapMemoryUsedPercent"`

	IOReadBytes  uint64 `json:"ioReadBytes"`
	IOWriteBytes uint64 `json:"ioWriteBytes"`
	IOCount      uint64 `json:"ioCount"`
	IOReadTime   uint64 `json:"ioReadTime"`
	IOWriteTime  uint64 `json:"ioWriteTime"`

	DiskData []DiskInfo `json:"diskData"`

	NetBytesSent uint64 `json:"netBytesSent"`
	NetBytesRecv uint64 `json:"netBytesRecv"`

	GPUData []GPUInfo `json:"gpuData"`
	XPUData []XPUInfo `json:"xpuData"`

	ShotTime time.Time `json:"shotTime"`
}

type AppLauncherSync struct {
	Keys []string `json:"keys"`
}

type DiskInfo struct {
	Path        string  `json:"path"`
	Type        string  `json:"type"`
	Device      string  `json:"device"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`

	InodesTotal       uint64  `json:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`
}

type GPUInfo struct {
	Index            uint   `json:"index"`
	ProductName      string `json:"productName"`
	GPUUtil          string `json:"gpuUtil"`
	Temperature      string `json:"temperature"`
	PerformanceState string `json:"performanceState"`
	PowerUsage       string `json:"powerUsage"`
	PowerDraw        string `json:"powerDraw"`
	MaxPowerLimit    string `json:"maxPowerLimit"`
	MemoryUsage      string `json:"memoryUsage"`
	MemUsed          string `json:"memUsed"`
	MemTotal         string `json:"memTotal"`
	FanSpeed         string `json:"fanSpeed"`
}

type AppLauncher struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Limit       int    `json:"limit"`
	Description string `json:"description"`
	Recommend   int    `json:"recommend"`

	IsInstall   bool            `json:"isInstall"`
	IsRecommend bool            `json:"isRecommend"`
	Detail      []InstallDetail `json:"detail"`
}

type InstallDetail struct {
	InstallID uint   `json:"installID"`
	DetailID  uint   `json:"detailID"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Status    string `json:"status"`
	WebUI     string `json:"webUI"`
	HttpPort  int    `json:"httpPort"`
	HttpsPort int    `json:"httpsPort"`
}

type LauncherOption struct {
	Key    string `json:"key"`
	IsShow bool   `json:"isShow"`
}

type XPUInfo struct {
	DeviceID    int    `json:"deviceID"`
	DeviceName  string `json:"deviceName"`
	Memory      string `json:"memory"`
	Temperature string `json:"temperature"`
	MemoryUsed  string `json:"memoryUsed"`
	Power       string `json:"power"`
	MemoryUtil  string `json:"memoryUtil"`
}
