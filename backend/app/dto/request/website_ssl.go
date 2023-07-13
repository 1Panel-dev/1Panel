package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type WebsiteSSLSearch struct {
	dto.PageInfo
	AcmeAccountID string `json:"acmeAccountID"`
}

type WebsiteSSLCreate struct {
	PrimaryDomain string `json:"primaryDomain" validate:"required"`
	OtherDomains  string `json:"otherDomains"`
	Provider      string `json:"provider" validate:"required"`
	AcmeAccountID uint   `json:"acmeAccountId" validate:"required"`
	DnsAccountID  uint   `json:"dnsAccountId"`
	AutoRenew     bool   `json:"autoRenew" validate:"required"`
}

type WebsiteDNSReq struct {
	Domains       []string `json:"domains" validate:"required"`
	AcmeAccountID uint     `json:"acmeAccountId"  validate:"required"`
}

type WebsiteSSLRenew struct {
	SSLID uint `json:"SSLId" validate:"required"`
}

type WebsiteAcmeAccountCreate struct {
	Email string `json:"email" validate:"required"`
}

type WebsiteDnsAccountCreate struct {
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Authorization map[string]string `json:"authorization" validate:"required"`
}

type WebsiteDnsAccountUpdate struct {
	ID            uint              `json:"id" validate:"required"`
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Authorization map[string]string `json:"authorization" validate:"required"`
}

type WebsiteResourceReq struct {
	ID uint `json:"id" validate:"required"`
}

type WebsiteSSLUpdate struct {
	ID        uint `json:"id" validate:"required"`
	AutoRenew bool `json:"autoRenew" validate:"required"`
}
