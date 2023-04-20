package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type NginxConfigFileUpdate struct {
	Content  string `json:"content" validate:"required"`
	FilePath string `json:"filePath" validate:"required"`
	Backup   bool   `json:"backup" validate:"required"`
}

type NginxScopeReq struct {
	Scope     dto.NginxKey `json:"scope" validate:"required"`
	WebsiteID uint         `json:"websiteId"`
}

type NginxConfigUpdate struct {
	Scope     dto.NginxKey `json:"scope"`
	Operate   string       `json:"operate" validate:"required;oneof=add update delete"`
	WebsiteID uint         `json:"websiteId" validate:"required"`
	Params    interface{}  `json:"params"`
}

type NginxRewriteReq struct {
	WebsiteID uint   `json:"websiteId" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

type NginxRewriteUpdate struct {
	WebsiteID uint   `json:"websiteId" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Content   string `json:"content" validate:"required"`
}
