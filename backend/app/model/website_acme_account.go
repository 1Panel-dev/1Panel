package model

type WebsiteAcmeAccount struct {
	BaseModel
	Email      string `gorm:"type:varchar(256);not null" json:"email"`
	URL        string `gorm:"type:varchar(256);not null" json:"url"`
	PrivateKey string `gorm:"type:longtext;not null" json:"-"`
	Type       string `gorm:"type:varchar(64);not null;default:letsencrypt" json:"type"`
	EabKid     string `gorm:"type:varchar(256);" json:"eabKid"`
	EabHmacKey string `gorm:"type:varchar(256);" json:"eabHmacKey"`
}

func (w WebsiteAcmeAccount) TableName() string {
	return "website_acme_accounts"
}
