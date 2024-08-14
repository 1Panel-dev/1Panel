package model

type WebsiteDnsAccount struct {
	BaseModel
	Name          string `gorm:"not null" json:"name"`
	Type          string `gorm:"not null" json:"type"`
	Authorization string `gorm:"not null" json:"-"`
}

func (w WebsiteDnsAccount) TableName() string {
	return "website_dns_accounts"
}
