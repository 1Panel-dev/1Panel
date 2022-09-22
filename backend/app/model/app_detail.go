package model

type AppDetail struct {
	BaseModel
	AppId         uint   `json:"appId" gorm:"type:integer;not null"`
	Version       string `json:"version" gorm:"type:varchar(64);not null"`
	FormFields    string `json:"formFields" gorm:"type:longtext;"`
	DockerCompose string `json:"dockerCompose"  gorm:"type:longtext;not null"`
	Readme        string `json:"readme"  gorm:"type:longtext;not null"`
}
