package model

type ScriptLibrary struct {
	BaseModel
	Name        string `json:"name" gorm:"not null;"`
	Lable       string `json:"lable"`
	Script      string `json:"script" gorm:"not null;"`
	Groups      string `json:"groups"`
	IsSystem    bool   `json:"isSystem"`
	Description string `json:"description"`
}
