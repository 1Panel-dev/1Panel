package dto

import "time"

type DashboardBase struct {
	HaloID         uint `json:"haloID"`
	DateeaseID     uint `json:"dateeaseID"`
	JumpServerID   uint `json:"jumpserverID"`
	MeterSphereID  uint `json:"metersphereID"`
	KubeoperatorID uint `json:"kubeoperatorID"`
	KubepiID       uint `json:"kubepiID"`

	WebsiteNumber     int `json:"websiteNumber"`
	DatabaseNumber    int `json:"databaseNumber"`
	CronjobNumber     int `json:"cronjobNumber"`
	AppInstalldNumber int `json:"appInstalldNumber"`

	Hostname             string `json:"hostname"`
	OS                   string `json:"os"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platformFamily"`
	PlatformVersion      string `json:"platformVersion"`
	KernelArch           string `json:"kernelArch"`
	KernelVersion        string `json:"kernelVersion"`
	VirtualizationSystem string `json:"virtualizationSystem"`

	CPUCores        int    `json:"cpuCores"`
	CPULogicalCores int    `json:"cpuLogicalCores"`
	CPUModelName    string `json:"cpuModelName"`

	CurrentInfo DashboardCurrent `json:"currentInfo"`
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
	MemoryUsedPercent float64 `json:"MemoryUsedPercent"`

	IOReadBytes  uint64 `json:"ioReadBytes"`
	IOWriteBytes uint64 `json:"ioWriteBytes"`
	IOCount      uint64 `json:"ioCount"`
	IOReadTime   uint64 `json:"ioReadTime"`
	IOWriteTime  uint64 `json:"ioWriteTime"`

	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`

	InodesTotal       uint64  `json:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`

	NetBytesSent uint64 `json:"netBytesSent"`
	NetBytesRecv uint64 `json:"netBytesRecv"`

	ShotTime time.Time `json:"shotTime"`
}
