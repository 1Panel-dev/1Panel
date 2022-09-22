package model

type AppConfig struct {
	BaseModel
	Version   string `json:"version"`
	OssPath   string `json:"ossPath"`
	CanUpdate bool   `json:"canUpdate"`
}
