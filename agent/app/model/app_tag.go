package model

type AppTag struct {
	BaseModel
	AppId uint `json:"appId" gorm:"not null"`
	TagId uint `json:"tagId" gorm:"not null"`
}
