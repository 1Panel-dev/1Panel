package model

type Runtime struct {
	BaseModel
	Name          string `gorm:"type:varchar;not null" json:"name"`
	AppDetailID   uint   `gorm:"type:integer" json:"appDetailId"`
	Image         string `gorm:"type:varchar" json:"image"`
	WorkDir       string `gorm:"type:varchar" json:"workDir"`
	DockerCompose string `gorm:"type:varchar" json:"dockerCompose"`
	Env           string `gorm:"type:varchar" json:"env"`
	Params        string `gorm:"type:varchar" json:"params"`
	Version       string `gorm:"type:varchar;not null" json:"version"`
	Type          string `gorm:"type:varchar;not null" json:"type"`
	Status        string `gorm:"type:varchar;not null" json:"status"`
	Resource      string `gorm:"type:varchar;not null" json:"resource"`
	Message       string `gorm:"type:longtext;" json:"message"`
}
