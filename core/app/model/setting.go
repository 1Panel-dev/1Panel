package model

type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"not null;"`
	Value string `json:"value"`
	About string `json:"about"`
}
