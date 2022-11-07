package dto

import "github.com/1Panel-dev/1Panel/backend/utils/nginx/components"

type NginxConfig struct {
	FilePath      string             `json:"filePath"`
	ContainerName string             `json:"containerName"`
	Config        *components.Config `json:"config"`
	OldContent    string             `json:"oldContent"`
}

type NginxConfigReq struct {
	Scope     NginxScope        `json:"scope"`
	WebSiteID uint              `json:"webSiteId" validate:"required"`
	Params    map[string]string `json:"params"`
}

type NginxScope string

const (
	Index NginxScope = "index"
)

var ScopeKeyMap = map[NginxScope][]string{
	Index: {"index"},
}
