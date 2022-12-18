package model

type ImageRepo struct {
	BaseModel

	Name        string `gorm:"type:varchar(64);not null" json:"name"`
	DownloadUrl string `gorm:"type:varchar(256)" json:"downloadUrl"`
	Protocol    string `gorm:"type:varchar(64)" json:"protocol"`
	Username    string `gorm:"type:varchar(256)" json:"username"`
	Password    string `gorm:"type:varchar(256)" json:"password"`
	Auth        bool   `gorm:"type:varchar(256)" json:"auth"`

	Status  string `gorm:"type:varchar(64)" json:"status"`
	Message string `gorm:"type:varchar(256)" json:"message"`
}
