package response

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type McpServersRes struct {
	Items []McpServerDTO `json:"items"`
	Total int64          `json:"total"`
}

type McpServerDTO struct {
	model.McpServer
	Environments []request.Environment `json:"environments"`
	Volumes      []request.Volume      `json:"volumes"`
}

type McpBindDomainRes struct {
	Domain        string   `json:"domain"`
	SSLID         uint     `json:"sslID"`
	AcmeAccountID uint     `json:"acmeAccountID"`
	AllowIPs      []string `json:"allowIPs"`
	WebsiteID     uint     `json:"websiteID"`
	ConnUrl       string   `json:"connUrl"`
}
