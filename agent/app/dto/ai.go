package dto

import "time"

type ModelGPUDevice struct {
	Index       int    `json:"index"`
	MemoryUsed  string `json:"memoryUsed"`
	MemoryTotal string `json:"memoryTotal"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	UUID        string `json:"uuid,omitempty"`
	Available   bool   `json:"available"`
	NUMANode    *int   `json:"numaNode,omitempty"`
}

type ModelGPUDevices struct {
	Devices   []ModelGPUDevice `json:"devices"`
	Warnings  []string         `json:"warnings"`
	NUMANodes []ModelNUMANode  `json:"numaNodes"`
}

type ModelNUMANode struct {
	ID     int    `json:"id"`
	CPUs   string `json:"cpus"`
	Memory bool   `json:"memory"`
}

type ModelNUMAConfig struct {
	Mode  string `json:"mode"`
	Nodes []int  `json:"nodes"`
}

type GPUGroupSource struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Revision uint   `json:"revision"`
}

type GPUGroupSave struct {
	Description string         `json:"description" validate:"max=500"`
	ID          uint           `json:"id"`
	Name        string         `json:"name" validate:"required,max=128"`
	ImageType   string         `json:"imageType" validate:"required,oneof=nvidia intel ascend"`
	GPU         ModelGPUConfig `json:"gpu"`
}

type GPUGroupInfo struct {
	GPUGroupSave
	Revision    uint     `json:"revision"`
	Deployments []string `json:"deployments"`
}

type ModelGPUConfig struct {
	Group        *GPUGroupSource  `json:"group,omitempty"`
	NUMA         *ModelNUMAConfig `json:"numa,omitempty"`
	Mode         string           `json:"mode"`
	Count        int              `json:"count,omitempty"`
	Error        string           `json:"error,omitempty"`
	Enabled      bool             `json:"enabled"`
	DeviceIDs    []string         `json:"deviceIDs"`
	Capabilities []string         `json:"capabilities"`
}

type OllamaModelInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Size         string `json:"size"`
	From         string `json:"from"`
	LogFileExist bool   `json:"logFileExist"`

	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type OllamaModelName struct {
	Name   string `json:"name"`
	TaskID string `json:"taskID"`
}

type OllamaModelAccountSync struct {
	ID                      uint   `json:"id" validate:"required"`
	ModelAccountBaseURLType string `json:"modelAccountBaseURLType" validate:"required,oneof=container localhost systemIP custom"`
	ModelAccountBaseURL     string `json:"modelAccountBaseURL"`
}

type OllamaBindDomain struct {
	Domain       string `json:"domain" validate:"required"`
	AppInstallID uint   `json:"appInstallID" validate:"required"`
	SSLID        uint   `json:"sslID"`
	WebsiteID    uint   `json:"websiteID"`
	IPList       string `json:"ipList"`
}

type OllamaBindDomainReq struct {
	AppInstallID uint `json:"appInstallID" validate:"required"`
}

type OllamaBindDomainRes struct {
	Domain        string   `json:"domain"`
	SSLID         uint     `json:"sslID"`
	AllowIPs      []string `json:"allowIPs"`
	WebsiteID     uint     `json:"websiteID"`
	ConnUrl       string   `json:"connUrl"`
	AcmeAccountID uint     `json:"acmeAccountID"`
}

type ModelGPUUsageRequest struct {
	DockerCompose    string `json:"dockerCompose" validate:"required"`
	Env              string `json:"env"`
	ImageType        string `json:"imageType" validate:"required,oneof=nvidia intel ascend"`
	ExcludeInstallID uint   `json:"excludeInstallID"`
}

type ModelGPUDeployment struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Possible bool   `json:"possible"`
}

type ModelGPUUsageDevice struct {
	ModelGPUDevice
	Possible    bool                 `json:"possible"`
	Deployments []ModelGPUDeployment `json:"deployments"`
}

type ModelGPUUsage struct {
	Devices  []ModelGPUUsageDevice `json:"devices"`
	Warnings []string              `json:"warnings"`
}
