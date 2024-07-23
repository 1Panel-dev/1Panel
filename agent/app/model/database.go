package model

type Database struct {
	BaseModel
	AppInstallID uint   `json:"appInstallID" gorm:"type:decimal"`
	Name         string `json:"name" gorm:"type:varchar(64);not null;unique"`
	Type         string `json:"type" gorm:"type:varchar(64);not null"`
	Version      string `json:"version" gorm:"type:varchar(64);not null"`
	From         string `json:"from" gorm:"type:varchar(64);not null"`
	Address      string `json:"address" gorm:"type:varchar(64);not null"`
	Port         uint   `json:"port" gorm:"type:decimal;not null"`
	Username     string `json:"username" gorm:"type:varchar(64)"`
	Password     string `json:"password" gorm:"type:varchar(64)"`

	SSL        bool   `json:"ssl"`
	RootCert   string `json:"rootCert" gorm:"type:longText"`
	ClientKey  string `json:"clientKey" gorm:"type:longText"`
	ClientCert string `json:"clientCert" gorm:"type:longText"`
	SkipVerify bool   `json:"skipVerify"`

	Description string `json:"description" gorm:"type:varchar(256);"`
}
