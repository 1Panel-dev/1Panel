package model

type Database struct {
	BaseModel
	AppInstallID uint   `json:"appInstallID"`
	Name         string `json:"name" gorm:"not null;unique"`
	Type         string `json:"type" gorm:"not null"`
	Version      string `json:"version" gorm:"not null"`
	From         string `json:"from" gorm:"not null"`
	Address      string `json:"address" gorm:"not null"`
	Port         uint   `json:"port" gorm:"not null"`
	Username     string `json:"username"`
	Password     string `json:"password"`

	SSL        bool   `json:"ssl"`
	RootCert   string `json:"rootCert"`
	ClientKey  string `json:"clientKey"`
	ClientCert string `json:"clientCert"`
	SkipVerify bool   `json:"skipVerify"`

	Timeout     uint   `json:"timeout"`
	Description string `json:"description"`
}
