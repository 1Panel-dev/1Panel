package model

type WebsiteTemplate struct {
	BaseModel
	Name      string `gorm:"not null" json:"name"`
	Type      string `gorm:"not null" json:"type"` // single | multi
	Content   string `gorm:"type:longtext" json:"content"`
	FilePath  string `json:"filePath"`
	Variables string `gorm:"type:text" json:"variables"`
	Remark    string `json:"remark"`
}

func (w WebsiteTemplate) TableName() string {
	return "website_templates"
}

type WebsiteTemplateOutput struct {
	BaseModel
	Name           string `gorm:"not null" json:"name"`
	TemplateID     uint   `gorm:"not null" json:"templateID"`
	TemplateType   string `json:"templateType"`
	VariableValues string `gorm:"type:text" json:"variableValues"`
	OutputPath     string `json:"outputPath"`
}

func (w WebsiteTemplateOutput) TableName() string {
	return "website_template_outputs"
}
