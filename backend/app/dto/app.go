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
	ShortDescZh        string   `json:"shortDescZh"`
	ShortDescEn        string   `json:"shortDescEn"`
	Type               string   `json:"type"`
	Required           []string `json:"Required"`
	CrossVersionUpdate bool     `json:"crossVersionUpdate"`
	Limit              int      `json:"limit"`
	Recommend          int      `json:"recommend"`
	Website            string   `json:"website"`
	Github             string   `json:"github"`
	Document           string   `json:"document"`
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
	FormFields []AppFormFields `json:"formFields"`
}

type AppFormFields struct {
	Type     string      `json:"type"`
	LabelZh  string      `json:"labelZh"`
	LabelEn  string      `json:"labelEn"`
	Required bool        `json:"required"`
	Default  interface{} `json:"default"`
	EnvKey   string      `json:"envKey"`
	Disabled bool        `json:"disabled"`
}

type AppResource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

var AppToolMap = map[string]string{
	"mysql": "phpmyadmin",
	"redis": "redis-commander",
}
