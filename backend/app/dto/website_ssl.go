package dto

import "github.com/1Panel-dev/1Panel/backend/app/model"

type WebsiteSSLDTO struct {
	model.WebSiteSSL
}

type SSLProvider string

const (
	DNSAccount = "dnsAccount"
	DnsCommon  = "dnsCommon"
	Http       = "http"
)

type WebsiteSSLCreate struct {
	Domains       []string    `json:"domains" validate:"required"`
	Provider      SSLProvider `json:"provider" validate:"required"`
	WebsiteID     uint        `json:"websiteId" validate:"required"`
	AcmeAccountID uint        `json:"acmeAccountId"`
	DnsAccountID  uint        `json:"dnsAccountId"`
}

type WebsiteSSLApply struct {
	WebsiteID uint `json:"websiteId" validate:"required"`
	SSLID     uint `json:"SSLId" validate:"required"`
}

type WebsiteDNSReq struct {
	Domains       []string `json:"domains" validate:"required"`
	AcmeAccountID uint     `json:"acmeAccountId"  validate:"required"`
}

type WebsiteDNSRes struct {
	Key   string `json:"resolve"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
