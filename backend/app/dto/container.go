package dto

import "time"

type PageContainer struct {
	PageInfo
	Status string `json:"status" validate:"required,oneof=all running"`
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
	ContainerPort int `json:"containerPort"`
	HostPort      int `json:"hostPort"`
}

type ContainerLog struct {
	ContainerID string `json:"containerID" validate:"required"`
	Mode        string `json:"mode" validate:"required"`
}

type ContainerOperation struct {
	ContainerID string `json:"containerID" validate:"required"`
	Operation   string `json:"operation" validate:"required,oneof=start stop reStart kill pause unPause reName remove"`
	NewName     string `json:"newName"`
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
	Ids []string `json:"ids" validate:"required"`
}
