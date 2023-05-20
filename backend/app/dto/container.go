package dto

import "time"

type PageContainer struct {
	PageInfo
	Name    string `json:"name"`
	Filters string `json:"filters"`
}

type InspectReq struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ContainerInfo struct {
	ContainerID string `json:"containerID"`
	Name        string `json:"name"`
	ImageId     string `json:"imageID"`
	ImageName   string `json:"imageName"`
	CreateTime  string `json:"createTime"`
	State       string `json:"state"`
	RunTime     string `json:"runTime"`

	CPUPercent    float64  `json:"cpuPercent"`
	MemoryPercent float64  `json:"memoryPercent"`
	Ports         []string `json:"ports"`

	IsFromApp     bool `json:"isFromApp"`
	IsFromCompose bool `json:"isFromCompose"`
}

type ContainerCreate struct {
	Name            string         `json:"name"`
	Image           string         `json:"image"`
	PublishAllPorts bool           `json:"publishAllPorts"`
	ExposedPorts    []PortHelper   `json:"exposedPorts"`
	Cmd             []string       `json:"cmd"`
	NanoCPUs        int64          `json:"nanoCPUs"`
	Memory          int64          `json:"memory"`
	AutoRemove      bool           `json:"autoRemove"`
	Volumes         []VolumeHelper `json:"volumes"`
	Labels          []string       `json:"labels"`
	Env             []string       `json:"env"`
	RestartPolicy   string         `json:"restartPolicy"`
}

type ContainterStats struct {
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

type ContainerLog struct {
	ContainerID string `json:"containerID" validate:"required"`
	Mode        string `json:"mode" validate:"required"`
}

type ContainerOperation struct {
	Name      string `json:"name" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=start stop restart kill pause unpause rename remove"`
	NewName   string `json:"newName"`
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
type NetworkCreat struct {
	Name    string   `json:"name"`
	Driver  string   `json:"driver"`
	Options []string `json:"options"`
	Subnet  string   `json:"subnet"`
	Gateway string   `json:"gateway"`
	IPRange string   `json:"ipRange"`
	Labels  []string `json:"labels"`
}

type Volume struct {
	Name       string    `json:"name"`
	Labels     []string  `json:"labels"`
	Driver     string    `json:"driver"`
	Mountpoint string    `json:"mountpoint"`
	CreatedAt  time.Time `json:"createdAt"`
}
type VolumeCreat struct {
	Name    string   `json:"name"`
	Driver  string   `json:"driver"`
	Options []string `json:"options"`
	Labels  []string `json:"labels"`
}

type BatchDelete struct {
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
}
type ComposeContainer struct {
	ContainerID string `json:"containerID"`
	Name        string `json:"name"`
	CreateTime  string `json:"createTime"`
	State       string `json:"state"`
}
type ComposeCreate struct {
	Name     string `json:"name"`
	From     string `json:"from" validate:"required,oneof=edit path template"`
	File     string `json:"file"`
	Path     string `json:"path"`
	Template uint   `json:"template"`
}
type ComposeOperation struct {
	Name      string `json:"name" validate:"required"`
	Path      string `json:"path" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=start stop down"`
	WithFile  bool   `json:"withFile"`
}
type ComposeUpdate struct {
	Name    string `json:"name" validate:"required"`
	Path    string `json:"path" validate:"required"`
	Content string `json:"content" validate:"required"`
}
