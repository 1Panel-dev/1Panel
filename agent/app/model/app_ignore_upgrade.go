package model

type AppIgnoreUpgrade struct {
	BaseModel
	AppID       uint   `json:"appID"`
	AppDetailID uint   `json:"appDetailID"`
	Scope       string `json:"scope"`
}
