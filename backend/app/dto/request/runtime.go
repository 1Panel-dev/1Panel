package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

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
