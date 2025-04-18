package model

type AppDetail struct {
	BaseModel
	AppId               uint   `json:"appId" gorm:"not null"`
	Version             string `json:"version" gorm:"not null"`
	Params              string `json:"-"`
	DockerCompose       string `json:"dockerCompose"`
	Status              string `json:"status" gorm:"not null"`
	LastVersion         string `json:"lastVersion"`
	LastModified        int    `json:"lastModified"`
	DownloadUrl         string `json:"downloadUrl"`
	DownloadCallBackUrl string `json:"downloadCallBackUrl"`
	Update              bool   `json:"update"`
}
