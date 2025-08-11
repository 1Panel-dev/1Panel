package request

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

type AppSearch struct {
	dto.PageInfo
	Name            string   `json:"name"`
	Tags            []string `json:"tags"`
	Type            string   `json:"type"`
	Recommend       bool     `json:"recommend"`
	Resource        string   `json:"resource"`
	ShowCurrentArch bool     `json:"showCurrentArch"`
}

type AppInstallCreate struct {
	AppDetailId uint                   `json:"appDetailId" validate:"required"`
	Params      map[string]interface{} `json:"params"`
	Name        string                 `json:"name" validate:"required"`
	Services    map[string]string      `json:"services"`
	TaskID      string                 `json:"taskID"`
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
	HostMode      bool    `json:"hostMode"`
	PullImage     bool    `json:"pullImage"`
	GpuConfig     bool    `json:"gpuConfig"`
	WebUI         string  `json:"webUI"`
	Type          string  `json:"type"`
	SpecifyIP     string  `json:"specifyIP"`
	RestartPolicy string  `json:"restartPolicy" validate:"omitempty,oneof=always unless-stopped no on-failure"`
}

type AppInstalledSearch struct {
	dto.PageInfo
	Type   string   `json:"type"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
	Update bool     `json:"update"`
	Unused bool     `json:"unused"`
	All    bool     `json:"all"`
	Sync   bool     `json:"sync"`
}

type AppInstalledInfo struct {
	Key  string `json:"key" validate:"required"`
	Name string `json:"name"`
}

type AppBackupSearch struct {
	dto.PageInfo
	AppInstallID uint `json:"appInstallID"`
}

type AppBackupDelete struct {
	Ids []uint `json:"ids"`
}

type AppInstalledOperate struct {
	InstallId     uint                `json:"installId" validate:"required"`
	BackupId      uint                `json:"backupId"`
	DetailId      uint                `json:"detailId"`
	Operate       constant.AppOperate `json:"operate" validate:"required"`
	ForceDelete   bool                `json:"forceDelete"`
	DeleteBackup  bool                `json:"deleteBackup"`
	DeleteDB      bool                `json:"deleteDB"`
	Backup        bool                `json:"backup"`
	PullImage     bool                `json:"pullImage"`
	DockerCompose string              `json:"dockerCompose"`
	TaskID        string              `json:"taskID"`
	DeleteImage   bool                `json:"deleteImage"`
	Favorite      bool                `json:"favorite"`
}

type AppInstallUpgrade struct {
	InstallID     uint   `json:"installId"`
	DetailID      uint   `json:"detailId"`
	Backup        bool   `json:"backup"`
	PullImage     bool   `json:"pullImage"`
	DockerCompose string `json:"dockerCompose"`
	TaskID        string `json:"taskID"`
}

type AppInstallDelete struct {
	Install      model.AppInstall
	DeleteBackup bool   `json:"deleteBackup"`
	ForceDelete  bool   `json:"forceDelete"`
	DeleteDB     bool   `json:"deleteDB"`
	DeleteImage  bool   `json:"deleteImage"`
	TaskID       string `json:"taskID"`
}

type AppInstalledUpdate struct {
	InstallId uint                   `json:"installId" validate:"required"`
	Params    map[string]interface{} `json:"params" validate:"required"`
	AppContainerConfig
}

type AppConfigUpdate struct {
	InstallID uint   `json:"installID" validate:"required"`
	WebUI     string `json:"webUI"`
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

type AppUpdateVersion struct {
	AppInstallID  uint   `json:"appInstallID" validate:"required"`
	UpdateVersion string `json:"updateVersion"`
}
