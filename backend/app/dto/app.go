package dto

import (
	"encoding/json"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type AppRes struct {
	Version   string      `json:"version"`
	CanUpdate bool        `json:"canUpdate"`
	Items     []*AppDTO   `json:"items"`
	Tags      []model.Tag `json:"tags"`
	Total     int64       `json:"total"`
}

type AppDTO struct {
	model.App
	Versions []string    `json:"versions"`
	Tags     []model.Tag `json:"tags"`
}

type AppDetailDTO struct {
	model.AppDetail
	Params interface{} `json:"params"`
}

type AppRequest struct {
	PageInfo
	Name string   `json:"name"`
	Tags []string `json:"tags"`
	Type string   `json:"type"`
}

type AppInstallRequest struct {
	AppDetailId uint                   `json:"appDetailId" validate:"required"`
	Params      map[string]interface{} `json:"params"`
	Name        string                 `json:"name" validate:"required"`
	Services    map[string]string      `json:"services"`
}

type CheckInstalled struct {
	IsExist       bool      `json:"isExist"`
	Name          string    `json:"name"`
	App           string    `json:"app"`
	Version       string    `json:"version"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	LastBackupAt  string    `json:"lastBackupAt"`
	AppInstallID  uint      `json:"appInstallId"`
	ContainerName string    `json:"containerName"`
}

type AppInstalled struct {
	model.AppInstall
	Total     int    `json:"total"`
	Ready     int    `json:"ready"`
	AppName   string `json:"appName"`
	Icon      string `json:"icon"`
	CanUpdate bool   `json:"canUpdate"`
}

type AppInstalledRequest struct {
	PageInfo
	Type string `json:"type"`
	Name string `json:"name"`
}

type AppBackupRequest struct {
	PageInfo
	AppInstallID uint `json:"appInstallID"`
}

type AppBackupDeleteRequest struct {
	Ids []uint `json:"ids"`
}

type AppOperate string

var (
	Up      AppOperate = "up"
	Down    AppOperate = "down"
	Restart AppOperate = "restart"
	Delete  AppOperate = "delete"
	Sync    AppOperate = "sync"
	Backup  AppOperate = "backup"
	Restore AppOperate = "restore"
	Update  AppOperate = "update"
)

type AppInstallOperate struct {
	InstallId uint       `json:"installId" validate:"required"`
	BackupId  uint       `json:"backupId"`
	DetailId  uint       `json:"detailId"`
	Operate   AppOperate `json:"operate" validate:"required"`
}

type PortUpdate struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Port int64  `json:"port"`
}

type AppService struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type AppDatabase struct {
	ServiceName string `json:"PANEL_DB_HOST"`
	DbName      string `json:"PANEL_DB_NAME"`
	DbUser      string `json:"PANEL_DB_USER"`
	Password    string `json:"PANEL_DB_USER_PASSWORD"`
}

type AuthParam struct {
	RootPassword string `json:"PANEL_DB_ROOT_PASSWORD"`
}

type ContainerExec struct {
	ContainerName string      `json:"containerName"`
	DbParam       AppDatabase `json:"dbParam"`
	Auth          AuthParam   `json:"auth"`
}

type AppOssConfig struct {
	Version string `json:"version"`
	Package string `json:"package"`
}

type AppVersion struct {
	Version  string `json:"version"`
	DetailId uint   `json:"detailId"`
}

type AppList struct {
	Version string      `json:"version"`
	Tags    []Tag       `json:"tags"`
	Items   []AppDefine `json:"items"`
}

type AppDefine struct {
	Key                string   `json:"key"`
	Name               string   `json:"name"`
	Tags               []string `json:"tags"`
	Versions           []string `json:"versions"`
	Author             string   `json:"author"`
	Source             string   `json:"source"`
	ShortDesc          string   `json:"short_desc"`
	Type               string   `json:"type"`
	Required           []string `json:"Required"`
	CrossVersionUpdate bool     `json:"crossVersionUpdate"`
	Limit              int      `json:"limit"`
}

func (define AppDefine) GetRequired() string {
	by, _ := json.Marshal(define.Required)
	return string(by)
}

type Tag struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type AppForm struct {
	FormFields []AppFormFields `json:"form_fields"`
}

type AppFormFields struct {
	Type     string `json:"type"`
	LabelZh  string `json:"label_zh"`
	LabelEn  string `json:"label_en"`
	Required string `json:"required"`
	Default  string `json:"default"`
	EnvKey   string `json:"env_variable"`
}
