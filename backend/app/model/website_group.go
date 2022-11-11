package model

type WebSiteGroup struct {
	BaseModel
	Name    string `gorm:"type:varchar(64);not null" json:"name"`
	Default bool   `json:"default"`
}

func (w WebSiteGroup) TableName() string {
	return "website_groups"
}
