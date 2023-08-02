package response

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type NginxStatus struct {
	Active   string `json:"active"`
	Accepts  string `json:"accepts"`
	Handled  string `json:"handled"`
	Requests string `json:"requests"`
	Reading  string `json:"reading"`
	Writing  string `json:"writing"`
	Waiting  string `json:"waiting"`
}

type NginxParam struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

type NginxAuthRes struct {
	Enable bool            `json:"enable"`
	Items  []dto.NginxAuth `json:"items"`
}

type NginxAntiLeechRes struct {
	Enable      bool     `json:"enable"`
	Extends     string   `json:"extends"`
	Return      string   `json:"return"`
	ServerNames []string `json:"serverNames"`
	Cache       bool     `json:"cache"`
	CacheTime   int      `json:"cacheTime"`
	CacheUint   string   `json:"cacheUint"`
	NoneRef     bool     `json:"noneRef"`
	LogEnable   bool     `json:"logEnable"`
	Blocked     bool     `json:"blocked"`
}

type NginxRedirectConfig struct {
	WebsiteID    uint     `json:"websiteID"`
	Name         string   `json:"name"`
	Domains      []string `json:"domains"`
	KeepPath     bool     `json:"keepPath"`
	Enable       bool     `json:"enable"`
	Type         string   `json:"type"`
	Redirect     string   `json:"redirect"`
	Path         string   `json:"path"`
	Target       string   `json:"target"`
	FilePath     string   `json:"filePath"`
	Content      string   `json:"content"`
	RedirectRoot bool     `json:"redirectRoot"`
}

type NginxFile struct {
	Content string `json:"content"`
}
