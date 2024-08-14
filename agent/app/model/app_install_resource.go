package model

type AppInstallResource struct {
	BaseModel
	AppInstallId uint   `json:"appInstallId" gorm:"not null;"`
	LinkId       uint   `json:"linkId"  gorm:"not null;"`
	ResourceId   uint   `json:"resourceId"`
	Key          string `json:"key" gorm:"not null"`
	From         string `json:"from" gorm:"not null;default:local"`
}
