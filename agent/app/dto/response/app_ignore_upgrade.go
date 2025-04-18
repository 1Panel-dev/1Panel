package response

type AppIgnoreUpgradeDTO struct {
	ID          uint   `json:"ID"`
	AppID       uint   `json:"appID"`
	AppDetailID uint   `json:"appDetailID"`
	Scope       string `json:"scope"`
	Version     string `json:"version"`
	Icon        string `json:"icon"`
}
