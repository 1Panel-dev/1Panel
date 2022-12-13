package model

type WebsiteGroup struct {
	BaseModel
	Name    string `gorm:"type:varchar(64);not null" json:"name"`
	Default bool   `json:"default"`
}

func (w WebsiteGroup) TableName() string {
	return "website_groups"
}
