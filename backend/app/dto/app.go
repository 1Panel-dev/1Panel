package dto

import (
	"encoding/json"
)

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

type AppResource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

var AppToolMap = map[string]string{
	"mysql": "phpmyadmin",
	"redis": "redis-commander",
}
