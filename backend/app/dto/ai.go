package dto

import "time"

type OllamaModelInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Size         string `json:"size"`
	From         string `json:"from"`
	LogFileExist bool   `json:"logFileExist"`

	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type OllamaModelDropList struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type OllamaModelName struct {
	Name string `json:"name"`
}

type OllamaBindDomain struct {
	Domain       string `json:"domain" validate:"required"`
	AppInstallID uint   `json:"appInstallID" validate:"required"`
	SSLID        uint   `json:"sslID"`
	WebsiteID    uint   `json:"websiteID"`
	IPList       string `json:"ipList"`
}

type OllamaBindDomainReq struct {
	AppInstallID uint `json:"appInstallID" validate:"required"`
}

type OllamaBindDomainRes struct {
	Domain        string   `json:"domain"`
	SSLID         uint     `json:"sslID"`
	AcmeAccountID uint     `json:"acmeAccountID"`
	AllowIPs      []string `json:"allowIPs"`
	WebsiteID     uint     `json:"websiteID"`
	ConnUrl       string   `json:"connUrl"`
}
