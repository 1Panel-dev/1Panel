package dto

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
