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
