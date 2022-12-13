package model

type WebsiteDomain struct {
	BaseModel
	WebsiteID uint   `gorm:"column:website_id;type:varchar(64);not null;" json:"websiteId"`
	Domain    string `gorm:"type:varchar(256);not null" json:"domain"`
	Port      int    `gorm:"type:integer" json:"port"`
}

func (w WebsiteDomain) TableName() string {
	return "website_domains"
}
