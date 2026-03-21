package request

import "github.com/1Panel-dev/1Panel/agent/app/dto"

type MailDomainCreate struct {
	Name       string `json:"name" validate:"required"`
	DnsAutoGen bool   `json:"dnsAutoGen"`
}

type MailDomainDelete struct {
	ID         uint `json:"id" validate:"required"`
	CleanupDns bool `json:"cleanupDns"`
}

type MailAccountSearch struct {
	dto.PageInfo
	DomainID uint   `json:"domainID"`
	Info     string `json:"info"`
}

type MailAccountCreate struct {
	DomainID uint   `json:"domainID" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	Quota    int    `json:"quota"`
}

type MailAccountUpdate struct {
	ID       uint   `json:"id" validate:"required"`
	Password string `json:"password"`
	Quota    int    `json:"quota"`
}

type MailAliasSearch struct {
	dto.PageInfo
	DomainID uint   `json:"domainID"`
	Info     string `json:"info"`
}

type MailAliasCreate struct {
	DomainID uint   `json:"domainID" validate:"required"`
	Source   string `json:"source" validate:"required"`
	Target   string `json:"target" validate:"required"`
}

type MailSetup struct {
	Enable bool `json:"enable"`
}
