package dto

import "time"

type PageContainer struct {
	PageInfo
	Status string `json:"status" validate:"required,oneof=all running"`
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
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Labels      []string  `json:"labels"`
	Driver      string    `json:"driver"`
	IPAMDriver  string    `json:"ipamDriver"`
	IPV4Subnet  string    `json:"ipv4Subnet"`
	IPV4Gateway string    `json:"ipv4Gateway"`
	IPV6Subnet  string    `json:"ipv6Subnet"`
	IPV6Gateway string    `json:"ipv6Gateway"`
	CreatedAt   time.Time `json:"createdAt"`
	Attachable  bool      `json:"attachable"`
}
type NetworkCreat struct {
	Name        string   `json:"name"`
	Driver      string   `json:"driver"`
	Options     []string `json:"options"`
	IPV4Subnet  string   `json:"ipv4Subnet"`
	IPV4Gateway string   `json:"ipv4Gateway"`
	Scope       string   `json:"scope"`
	Labels      []string `json:"labels"`
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
