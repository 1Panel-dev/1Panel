package request

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
)

type AppSearch struct {
	dto.PageInfo
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Type      string   `json:"type"`
	Recommend bool     `json:"recommend"`
}

type AppInstallCreate struct {
	AppDetailId uint                   `json:"appDetailId" validate:"required"`
	Params      map[string]interface{} `json:"params"`
	Name        string                 `json:"name" validate:"required"`
	Services    map[string]string      `json:"services"`
	AppContainerConfig
}

type AppContainerConfig struct {
	Advanced      bool    `json:"advanced"`
	CpuQuota      float64 `json:"cpuQuota"`
	MemoryLimit   float64 `json:"memoryLimit"`
	MemoryUnit    string  `json:"memoryUnit"`
	ContainerName string  `json:"containerName"`
	AllowPort     bool    `json:"allowPort"`
	EditCompose   bool    `json:"editCompose"`
	DockerCompose string  `json:"dockerCompose"`
}

type AppInstalledSearch struct {
	dto.PageInfo
	Type   string   `json:"type"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
	Update bool     `json:"update"`
	Unused bool     `json:"unused"`
}

type AppBackupSearch struct {
	dto.PageInfo
	AppInstallID uint `json:"appInstallID"`
}

type AppBackupDelete struct {
	Ids []uint `json:"ids"`
}

type AppInstalledOperate struct {
	InstallId    uint                `json:"installId" validate:"required"`
	BackupId     uint                `json:"backupId"`
	DetailId     uint                `json:"detailId"`
	Operate      constant.AppOperate `json:"operate" validate:"required"`
	ForceDelete  bool                `json:"forceDelete"`
	DeleteBackup bool                `json:"deleteBackup"`
	DeleteDB     bool                `json:"deleteDB"`
}

type AppInstalledUpdate struct {
	InstallId uint                   `json:"installId" validate:"required"`
	Params    map[string]interface{} `json:"params" validate:"required"`
	AppContainerConfig
}

type AppInstalledIgnoreUpgrade struct {
	DetailID uint   `json:"detailID"  validate:"required"`
	Operate  string `json:"operate"   validate:"required,oneof=cancel ignore"`
}

type PortUpdate struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Port int64  `json:"port"`
}
