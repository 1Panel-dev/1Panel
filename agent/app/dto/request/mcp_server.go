package request

import "github.com/1Panel-dev/1Panel/agent/app/dto"

type McpServerSearch struct {
	dto.PageInfo
	Name string `json:"name"`
	Sync bool   `json:"sync"`
}

type McpServerCreate struct {
	Name          string        `json:"name" validate:"required"`
	Command       string        `json:"command" validate:"required"`
	Environments  []Environment `json:"environments"`
	Volumes       []Volume      `json:"volumes"`
	Port          int           `json:"port" validate:"required"`
	ContainerName string        `json:"containerName"`
	BaseURL       string        `json:"baseUrl"`
	SsePath       string        `json:"ssePath"`
	HostIP        string        `json:"hostIP"`
}

type McpServerUpdate struct {
	ID uint `json:"id" validate:"required"`
	McpServerCreate
}

type McpServerDelete struct {
	ID uint `json:"id" validate:"required"`
}

type McpServerOperate struct {
	ID      uint   `json:"id" validate:"required"`
	Operate string `json:"operate" validate:"required"`
}

type McpBindDomain struct {
	Domain string `json:"domain" validate:"required"`
	SSLID  uint   `json:"sslID"`
	IPList string `json:"ipList"`
}

type McpBindDomainUpdate struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	SSLID     uint   `json:"sslID"`
	IPList    string `json:"ipList"`
}
