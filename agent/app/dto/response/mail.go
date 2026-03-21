package response

import "github.com/1Panel-dev/1Panel/agent/app/model"

type MailDomainRes struct {
	model.MailDomain
	AccountCount int `json:"accountCount"`
	AliasCount   int `json:"aliasCount"`
}

type MailAccountRes struct {
	model.MailAccount
	DomainName string `json:"domainName"`
}

type MailAliasRes struct {
	model.MailAlias
	DomainName string `json:"domainName"`
}

type MailStatus struct {
	Enabled           bool   `json:"enabled"`
	MailContainerUp   bool   `json:"mailContainerUp"`
	MailContainerName string `json:"mailContainerName"`
	RoundcubeUp       bool   `json:"roundcubeUp"`
	RoundcubeURL      string `json:"roundcubeURL"`
	Version           string `json:"version"`
}
