package dto

type WebPage struct {
	PageInfo
}

type WebSiteCreate struct {
	PrimaryDomain  string   `json:"primaryDomain"`
	Type           string   `json:"type"`
	Alias          string   `json:"alias"`
	Remark         string   `json:"remark"`
	Domains        []string `json:"domains"`
	AppType        string   `json:"appType"`
	AppInstallID   uint     `json:"appInstallID"`
	WebSiteGroupID uint     `json:"webSiteGroupID"`
}
