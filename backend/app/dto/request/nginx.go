package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type NginxConfigFileUpdate struct {
	Content string `json:"content" validate:"required"`
	Backup  bool   `json:"backup"`
}

type NginxScopeReq struct {
	Scope     dto.NginxKey `json:"scope" validate:"required"`
	WebsiteID uint         `json:"websiteId"`
}

type NginxConfigUpdate struct {
	Scope     dto.NginxKey `json:"scope"`
	Operate   string       `json:"operate" validate:"required,oneof=add update delete"`
	WebsiteID uint         `json:"websiteId"`
	Params    interface{}  `json:"params"`
}

type NginxRewriteReq struct {
	WebsiteID uint   `json:"websiteId" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

type NginxRewriteUpdate struct {
	WebsiteID uint   `json:"websiteId" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Content   string `json:"content"`
}

type NginxProxyUpdate struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Content   string `json:"content" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

type NginxAuthUpdate struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Operate   string `json:"operate" validate:"required"`
	Username  string `json:"username"`
	Password  string `json:"password"`
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
	Enable      bool     `json:"enable" `
	ServerNames []string `json:"serverNames"`
	Cache       bool     `json:"cache"`
	CacheTime   int      `json:"cacheTime"`
	CacheUint   string   `json:"cacheUint"`
	NoneRef     bool     `json:"noneRef"`
	LogEnable   bool     `json:"logEnable"`
	Blocked     bool     `json:"blocked"`
}

type NginxRedirectReq struct {
	Name         string   `json:"name" validate:"required"`
	WebsiteID    uint     `json:"websiteID" validate:"required"`
	Domains      []string `json:"domains"`
	KeepPath     bool     `json:"keepPath"`
	Enable       bool     `json:"enable"`
	Type         string   `json:"type" validate:"required"`
	Redirect     string   `json:"redirect" validate:"required"`
	Path         string   `json:"path"`
	Target       string   `json:"target" validate:"required"`
	Operate      string   `json:"operate" validate:"required"`
	RedirectRoot bool     `json:"redirectRoot"`
}

type NginxRedirectUpdate struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Content   string `json:"content" validate:"required"`
	Name      string `json:"name" validate:"required"`
}
