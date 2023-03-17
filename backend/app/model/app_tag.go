package model

type AppTag struct {
	BaseModel
	AppId uint `json:"appId" gorm:"type:integer;not null"`
	TagId uint `json:"tagId" gorm:"type:integer;not null"`
}
