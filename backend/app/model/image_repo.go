package model

type ImageRepo struct {
	BaseModel

	Name        string `gorm:"type:varchar(64);not null" json:"name"`
	DownloadUrl string `gorm:"type:varchar(256)" json:"downloadUrl"`
	RepoName    string `gorm:"type:varchar(256)" json:"repoName"`
	Username    string `gorm:"type:varchar(256)" json:"username"`
	Password    string `gorm:"type:varchar(256)" json:"password"`
	Auth        bool   `gorm:"type:varchar(256)" json:"auth"`
}
