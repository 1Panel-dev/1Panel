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

type NginxProxyUpdate struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Content   string `json:"content" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

type NginxAuthUpdate struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Operate   string `json:"operate" validate:"required"`
	Username  string `json:"username"  validate:"required"`
	Password  string `json:"password" validate:"required"`
	Remark    string `json:"remark"`
}

type NginxAuthReq struct {
	WebsiteID uint `json:"websiteID" validate:"required"`
}

type NginxCommonReq struct {
	WebsiteID uint `json:"websiteID" validate:"required"`
}

type NginxAntiLeechUpdate struct {
	WebsiteID   uint     `json:"websiteID" validate:"required"`
	Extends     string   `json:"extends" validate:"required"`
	Return      string   `json:"return" validate:"required"`
	Enable      bool     `json:"enable"  validate:"required"`
	ServerNames []string `json:"serverNames"`
	Cache       bool     `json:"cache"`
	CacheTime   int      `json:"cacheTime"`
	CacheUint   string   `json:"cacheUint"`
	NoneRef     bool     `json:"noneRef"`
	LogEnable   bool     `json:"logEnable"`
	Blocked     bool     `json:"blocked"`
}
