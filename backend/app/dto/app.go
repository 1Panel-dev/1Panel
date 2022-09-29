package dto

import "github.com/1Panel-dev/1Panel/app/model"

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

type AppList struct {
	Version string      `json:"version"`
	Tags    []Tag       `json:"tags"`
	Items   []AppDefine `json:"items"`
}

type AppDefine struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Versions  []string `json:"versions"`
	Icon      string   `json:"icon"`
	Author    string   `json:"author"`
	Source    string   `json:"source"`
	ShortDesc string   `json:"short_desc"`
	Type      string   `json:"type"`
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

type AppRequest struct {
	PageInfo
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type AppInstallRequest struct {
	AppDetailId uint                   `json:"appDetailId" validate:"required"`
	Params      map[string]interface{} `json:"params"`
	Name        string                 `json:"name" validate:"required"`
}

type AppInstalled struct {
	model.AppInstall
	Total   int    `json:"total"`
	Ready   int    `json:"ready"`
	AppName string `json:"appName"`
	Icon    string `json:"icon"`
}

type AppInstalledRequest struct {
	PageInfo
}

type AppOperate string

var (
	Up      AppOperate = "up"
	Down    AppOperate = "down"
	Restart AppOperate = "restart"
	Delete  AppOperate = "delete"
	Sync    AppOperate = "sync"
)

type AppInstallOperate struct {
	InstallId uint       `json:"installId" validate:"required"`
	Operate   AppOperate `json:"operate" validate:"required"`
}

//type AppContainer struct {
//	Names  []string `json:"names"`
//	Image  string   `json:"image"`
//	Ports  string   `json:"ports"`
//	Status string   `json:"status"`
//	State  string   `json:"state"`
//}
