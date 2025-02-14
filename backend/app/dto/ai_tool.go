package dto

type OllamaModelInfo struct {
	Name     string `json:"name"`
	Size     string `json:"size"`
	Modified string `json:"modified"`
}

type OllamaModelName struct {
	Name string `json:"name"`
}

type OllamaBindDomain struct {
	Domain       string   `json:"domain" validate:"required"`
	AppInstallID uint     `json:"appInstallID" validate:"required"`
	SSLID        uint     `json:"sslID"`
	AllowIPs     []string `json:"allowIPs"`
	WebsiteID    uint     `json:"websiteID"`
}

type OllamaBindDomainReq struct {
	AppInstallID uint `json:"appInstallID" validate:"required"`
}

type OllamaBindDomainRes struct {
	Domain    string   `json:"domain"`
	SSLID     uint     `json:"sslID"`
	AllowIPs  []string `json:"allowIPs"`
	WebsiteID uint     `json:"websiteID"`
}
