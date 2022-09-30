package model

type AppDetail struct {
	BaseModel
	AppId         uint   `json:"appId" gorm:"type:integer;not null"`
	Version       string `json:"version" gorm:"type:varchar(64);not null"`
	Params        string `json:"-" gorm:"type:longtext;"`
	DockerCompose string `json:"-"  gorm:"type:longtext;not null"`
	Readme        string `json:"readme"  gorm:"type:longtext;not null"`
	Status        string `json:"status" gorm:"type:varchar(64);not null"`
}
