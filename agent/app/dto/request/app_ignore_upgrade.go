package request

type AppIgnoreUpgradeReq struct {
	AppID       uint   `json:"appID" validate:"required"`
	AppDetailID uint   `json:"appDetailID"`
	Scope       string `json:"scope" validate:"required,oneof=all version"`
}
