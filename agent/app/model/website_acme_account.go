package model

type WebsiteAcmeAccount struct {
	BaseModel
	Email      string `gorm:"not null" json:"email"`
	URL        string `gorm:"not null" json:"url"`
	PrivateKey string `gorm:"not null" json:"-"`
	Type       string `gorm:"not null;default:letsencrypt" json:"type"`
	EabKid     string `gorm:"default:null;" json:"eabKid"`
	EabHmacKey string `gorm:"default:null" json:"eabHmacKey"`
	KeyType    string `gorm:"not null;default:2048" json:"keyType"`
}

func (w WebsiteAcmeAccount) TableName() string {
	return "website_acme_accounts"
}
