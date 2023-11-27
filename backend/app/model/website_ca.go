package model

type WebsiteCA struct {
	BaseModel
	CSR        string `gorm:"not null;" json:"csr"`
	Name       string `gorm:"not null;" json:"name"`
	PrivateKey string `gorm:"not null" json:"privateKey"`
	KeyType    string `gorm:"not null;default:2048" json:"keyType"`
}
