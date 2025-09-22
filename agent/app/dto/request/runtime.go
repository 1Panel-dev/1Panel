package request

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

type RuntimeSearch struct {
	dto.PageInfo
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type RuntimeCreate struct {
	AppDetailID uint   `json:"appDetailId"`
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Image       string `json:"image"`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	CodeDir     string `json:"codeDir"`
	Remark      string `json:"remark"`

	Params map[string]interface{} `json:"params"`
	NodeConfig
}

type NodeConfig struct {
	Install      bool          `json:"install"`
	Clean        bool          `json:"clean"`
	ExposedPorts []ExposedPort `json:"exposedPorts"`
	Environments []Environment `json:"environments"`
	Volumes      []Volume      `json:"volumes"`
}

type Environment struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Volume struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type ExposedPort struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	HostIP        string `json:"hostIP"`
}

type RuntimeDelete struct {
	ID          uint `json:"id"`
	ForceDelete bool `json:"forceDelete"`
}

type RuntimeUpdate struct {
	Name    string `json:"name"`
	ID      uint   `json:"id"`
	Image   string `json:"image"`
	Version string `json:"version"`
	Rebuild bool   `json:"rebuild"`
	Source  string `json:"source"`
	CodeDir string `json:"codeDir"`
	Remark  string `json:"remark"`

	Params map[string]interface{} `json:"params"`
	NodeConfig
}

type NodePackageReq struct {
	CodeDir string `json:"codeDir"`
}

type RuntimeOperate struct {
	Operate string `json:"operate"`
	ID      uint   `json:"ID"`
}

type NodeModuleOperateReq struct {
	Operate    string `json:"operate" validate:"oneof=install uninstall update"`
	ID         uint   `json:"ID" validate:"required"`
	Module     string `json:"module"`
	PkgManager string `json:"pkgManager" validate:"oneof=npm yarn"`
}

type NodeModuleReq struct {
	ID uint `json:"ID" validate:"required"`
}

type PHPExtensionInstallReq struct {
	ID     uint   `json:"ID" validate:"required"`
	Name   string `json:"name" validate:"required"`
	TaskID string `json:"taskID"`
}

type PHPConfigUpdate struct {
	ID               uint              `json:"id" validate:"required"`
	Params           map[string]string `json:"params"`
	Scope            string            `json:"scope" validate:"required"`
	DisableFunctions []string          `json:"disableFunctions"`
	UploadMaxSize    string            `json:"uploadMaxSize"`
	MaxExecutionTime string            `json:"maxExecutionTime"`
}

type PHPFileUpdate struct {
	ID      uint   `json:"id" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type PHPFileReq struct {
	ID   uint   `json:"id" validate:"required"`
	Type string `json:"type" validate:"required"`
}

type FPMConfig struct {
	ID     uint                   `json:"id" validate:"required"`
	Params map[string]interface{} `json:"params" validate:"required"`
}

type PHPSupervisorProcessConfig struct {
	ID uint `json:"id" validate:"required"`
	SupervisorProcessConfig
}

type PHPSupervisorProcessFileReq struct {
	ID uint `json:"id" validate:"required"`
	SupervisorProcessFileReq
}

type PHPContainerConfig struct {
	ID            uint          `json:"id" validate:"required"`
	ContainerName string        `json:"containerName"`
	ExposedPorts  []ExposedPort `json:"exposedPorts"`
	Environments  []Environment `json:"environments"`
	Volumes       []Volume      `json:"volumes"`
}

type RuntimeRemark struct {
	ID     uint   `json:"id" validate:"required"`
	Remark string `json:"remark"`
}
