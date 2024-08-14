package model

type ImageRepo struct {
	BaseModel

	Name        string `gorm:"not null" json:"name"`
	DownloadUrl string `json:"downloadUrl"`
	Protocol    string `json:"protocol"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Auth        bool   `json:"auth"`

	Status  string `json:"status"`
	Message string `json:"message"`
}
