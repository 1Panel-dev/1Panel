package dto

import (
	"time"
)

type PageContainer struct {
	PageInfo
	Name            string `json:"name"`
	State           string `json:"state" validate:"required,oneof=all created running paused restarting removing exited dead"`
	OrderBy         string `json:"orderBy" validate:"required,oneof=name state created_at"`
	Order           string `json:"order" validate:"required,oneof=null ascending descending"`
	Filters         string `json:"filters"`
	ExcludeAppStore bool   `json:"excludeAppStore"`
}

type InspectReq struct {
	ID   string `json:"id" validate:"required"`
	Type string `json:"type" validate:"required"`
}

type ContainerInfo struct {
	ContainerID string `json:"containerID"`
	Name        string `json:"name"`
	ImageId     string `json:"imageID"`
	ImageName   string `json:"imageName"`
	CreateTime  string `json:"createTime"`
	State       string `json:"state"`
	RunTime     string `json:"runTime"`

	Network []string `json:"network"`
	Ports   []string `json:"ports"`

	IsFromApp     bool `json:"isFromApp"`
	IsFromCompose bool `json:"isFromCompose"`

	AppName        string   `json:"appName"`
	AppInstallName string   `json:"appInstallName"`
	Websites       []string `json:"websites"`
}

type ResourceLimit struct {
	CPU    int    `json:"cpu"`
	Memory uint64 `json:"memory"`
}

type ContainerOperate struct {
	ContainerID     string         `json:"containerID"`
	ForcePull       bool           `json:"forcePull"`
	Name            string         `json:"name" validate:"required"`
	Image           string         `json:"image" validate:"required"`
	Network         string         `json:"network"`
	Ipv4            string         `json:"ipv4"`
	Ipv6            string         `json:"ipv6"`
	PublishAllPorts bool           `json:"publishAllPorts"`
	ExposedPorts    []PortHelper   `json:"exposedPorts"`
	Tty             bool           `json:"tty"`
	OpenStdin       bool           `json:"openStdin"`
	Cmd             []string       `json:"cmd"`
	Entrypoint      []string       `json:"entrypoint"`
	CPUShares       int64          `json:"cpuShares"`
	NanoCPUs        float64        `json:"nanoCPUs"`
	Memory          float64        `json:"memory"`
	Privileged      bool           `json:"privileged"`
	AutoRemove      bool           `json:"autoRemove"`
	Volumes         []VolumeHelper `json:"volumes"`
	Labels          []string       `json:"labels"`
	Env             []string       `json:"env"`
	RestartPolicy   string         `json:"restartPolicy"`
}

type ContainerUpgrade struct {
	Name      string `json:"name" validate:"required"`
	Image     string `json:"image" validate:"required"`
	ForcePull bool   `json:"forcePull"`
}

type ContainerListStats struct {
	ContainerID string `json:"containerID"`

	CPUTotalUsage uint64  `json:"cpuTotalUsage"`
	SystemUsage   uint64  `json:"systemUsage"`
	CPUPercent    float64 `json:"cpuPercent"`
	PercpuUsage   int     `json:"percpuUsage"`

	MemoryCache   uint64  `json:"memoryCache"`
	MemoryUsage   uint64  `json:"memoryUsage"`
	MemoryLimit   uint64  `json:"memoryLimit"`
	MemoryPercent float64 `json:"memoryPercent"`
}

type ContainerStats struct {
	CPUPercent float64 `json:"cpuPercent"`
	Memory     float64 `json:"memory"`
	Cache      float64 `json:"cache"`
	IORead     float64 `json:"ioRead"`
	IOWrite    float64 `json:"ioWrite"`
	NetworkRX  float64 `json:"networkRX"`
	NetworkTX  float64 `json:"networkTX"`

	ShotTime time.Time `json:"shotTime"`
}

type VolumeHelper struct {
	Type         string `json:"type"`
	SourceDir    string `json:"sourceDir"`
	ContainerDir string `json:"containerDir"`
	Mode         string `json:"mode"`
}
type PortHelper struct {
	HostIP        string `json:"hostIP"`
	HostPort      string `json:"hostPort"`
	ContainerPort string `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

type ContainerOperation struct {
	Names     []string `json:"names" validate:"required"`
	Operation string   `json:"operation" validate:"required,oneof=up start stop restart kill pause unpause remove"`
}

type ContainerRename struct {
	Name    string `json:"name" validate:"required"`
	NewName string `json:"newName" validate:"required"`
}

type ContainerCommit struct {
	ContainerId   string `json:"containerID" validate:"required"`
	ContainerName string `json:"containerName"`
	NewImageName  string `json:"newImageName"`
	Comment       string `json:"comment"`
	Author        string `json:"author"`
	Pause         bool   `json:"pause"`
}

type ContainerPrune struct {
	PruneType  string `json:"pruneType" validate:"required,oneof=container image volume network buildcache"`
	WithTagAll bool   `json:"withTagAll"`
}

type ContainerPruneReport struct {
	DeletedNumber  int `json:"deletedNumber"`
	SpaceReclaimed int `json:"spaceReclaimed"`
}

type Network struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Labels     []string  `json:"labels"`
	Driver     string    `json:"driver"`
	IPAMDriver string    `json:"ipamDriver"`
	Subnet     string    `json:"subnet"`
	Gateway    string    `json:"gateway"`
	CreatedAt  time.Time `json:"createdAt"`
	Attachable bool      `json:"attachable"`
}
type NetworkCreate struct {
	Name       string          `json:"name" validate:"required"`
	Driver     string          `json:"driver" validate:"required"`
	Options    []string        `json:"options"`
	Ipv4       bool            `json:"ipv4"`
	Subnet     string          `json:"subnet"`
	Gateway    string          `json:"gateway"`
	IPRange    string          `json:"ipRange"`
	AuxAddress []SettingUpdate `json:"auxAddress"`

	Ipv6         bool            `json:"ipv6"`
	SubnetV6     string          `json:"subnetV6"`
	GatewayV6    string          `json:"gatewayV6"`
	IPRangeV6    string          `json:"ipRangeV6"`
	AuxAddressV6 []SettingUpdate `json:"auxAddressV6"`
	Labels       []string        `json:"labels"`
}

type Volume struct {
	Name       string    `json:"name"`
	Labels     []string  `json:"labels"`
	Driver     string    `json:"driver"`
	Mountpoint string    `json:"mountpoint"`
	CreatedAt  time.Time `json:"createdAt"`
}
type VolumeCreate struct {
	Name    string   `json:"name" validate:"required"`
	Driver  string   `json:"driver" validate:"required"`
	Options []string `json:"options"`
	Labels  []string `json:"labels"`
}

type BatchDelete struct {
	Force bool     `json:"force"`
	Names []string `json:"names" validate:"required"`
}

type ComposeInfo struct {
	Name            string             `json:"name"`
	CreatedAt       string             `json:"createdAt"`
	CreatedBy       string             `json:"createdBy"`
	ContainerNumber int                `json:"containerNumber"`
	ConfigFile      string             `json:"configFile"`
	Workdir         string             `json:"workdir"`
	Path            string             `json:"path"`
	Containers      []ComposeContainer `json:"containers"`
	Env             []string           `json:"env"`
}
type ComposeContainer struct {
	ContainerID string `json:"containerID"`
	Name        string `json:"name"`
	CreateTime  string `json:"createTime"`
	State       string `json:"state"`
}
type ComposeCreate struct {
	Name     string   `json:"name"`
	From     string   `json:"from" validate:"required,oneof=edit path template"`
	File     string   `json:"file"`
	Path     string   `json:"path"`
	Template uint     `json:"template"`
	Env      []string `json:"env"`
}
type ComposeOperation struct {
	Name      string `json:"name" validate:"required"`
	Path      string `json:"path"`
	Operation string `json:"operation" validate:"required,oneof=up start stop down delete"`
	WithFile  bool   `json:"withFile"`
}
type ComposeUpdate struct {
	Name    string   `json:"name" validate:"required"`
	Path    string   `json:"path" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Env     []string `json:"env"`
}

type ContainerLog struct {
	Container     string `json:"container" validate:"required"`
	Since         string `json:"since"`
	Tail          uint   `json:"tail"`
	ContainerType string `json:"containerType"`
}
