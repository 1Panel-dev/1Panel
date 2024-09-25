package request

import "github.com/1Panel-dev/1Panel/agent/app/dto"

type RuntimeSearch struct {
	dto.PageInfo
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type RuntimeCreate struct {
	AppDetailID uint                   `json:"appDetailId"`
	Name        string                 `json:"name"`
	Params      map[string]interface{} `json:"params"`
	Resource    string                 `json:"resource"`
	Image       string                 `json:"image"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	Source      string                 `json:"source"`
	CodeDir     string                 `json:"codeDir"`
	NodeConfig
}

type NodeConfig struct {
	Install      bool          `json:"install"`
	Clean        bool          `json:"clean"`
	Port         int           `json:"port"`
	ExposedPorts []ExposedPort `json:"exposedPorts"`
}

type ExposedPort struct {
	HostPort      int `json:"hostPort"`
	ContainerPort int `json:"containerPort"`
}

type RuntimeDelete struct {
	ID          uint `json:"id"`
	ForceDelete bool `json:"forceDelete"`
}

type RuntimeUpdate struct {
	Name    string                 `json:"name"`
	ID      uint                   `json:"id"`
	Params  map[string]interface{} `json:"params"`
	Image   string                 `json:"image"`
	Version string                 `json:"version"`
	Rebuild bool                   `json:"rebuild"`
	Source  string                 `json:"source"`
	CodeDir string                 `json:"codeDir"`
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
