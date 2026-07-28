package response

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type WebsiteTemplateDTO struct {
	model.WebsiteTemplate
}

type WebsiteTemplateOutputDTO struct {
	model.WebsiteTemplateOutput
	TemplateName string `json:"templateName"`
}

type WebsitePreviewDTO struct {
	HTML string `json:"html"`
}
