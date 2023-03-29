package model

type Runtime struct {
	BaseModel
	Name          string `gorm:"type:varchar;not null" json:"name"`
	AppDetailID   uint   `gorm:"type:integer" json:"appDetailId"`
	Image         string `gorm:"type:varchar;not null" json:"image"`
	WorkDir       string `gorm:"type:varchar;not null" json:"workDir"`
	DockerCompose string `gorm:"type:varchar;not null" json:"dockerCompose"`
	Env           string `gorm:"type:varchar;not null" json:"env"`
	Params        string `gorm:"type:varchar;not null" json:"params"`
	Type          string `gorm:"type:varchar;not null" json:"type"`
}
