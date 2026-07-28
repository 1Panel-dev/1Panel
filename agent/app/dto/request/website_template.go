package request

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

type WebsiteTemplateSearch struct {
	dto.PageInfo
	Name string `json:"name"`
	Type string `json:"type"`
}

type WebsiteTemplateCreate struct {
	Name      string `json:"name" validate:"required"`
	Type      string `json:"type" validate:"required,oneof=single multi"`
	Content   string `json:"content"`
	FilePath  string `json:"filePath"`
	Variables string `json:"variables"`
	Remark    string `json:"remark"`
}

type WebsiteTemplateUpdate struct {
	ID        uint   `json:"id" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Type      string `json:"type" validate:"required,oneof=single multi"`
	Content   string `json:"content"`
	FilePath  string `json:"filePath"`
	Variables string `json:"variables"`
	Remark    string `json:"remark"`
}

type WebsiteTemplateOutputSearch struct {
	dto.PageInfo
	TemplateID uint `json:"templateID"`
}

type WebsiteTemplateOutputCreate struct {
	TemplateID     uint              `json:"templateID" validate:"required"`
	Name           string            `json:"name" validate:"required"`
	VariableValues map[string]string `json:"variableValues"`
}

type WebsitePreviewReq struct {
	TemplateID     uint              `json:"templateID" validate:"required"`
	VariableValues map[string]string `json:"variableValues"`
}

type WebsiteTemplateImportReq struct {
	OutputID uint   `json:"outputID" validate:"required"`
	SiteDir  string `json:"siteDir" validate:"required"`
}
