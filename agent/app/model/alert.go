package model

type Alert struct {
	BaseModel

	Title     string `gorm:"type:varchar(256);not null" json:"title"`
	Type      string `gorm:"type:varchar(64);not null" json:"type"`
	Cycle     uint   `gorm:"type:integer;not null" json:"cycle"`
	Count     uint   `gorm:"type:integer;not null" json:"count"`
	Project   string `gorm:"type:varchar(64)" json:"project"`
	Status    string `gorm:"type:varchar(64);not null" json:"status"`
	Method    string `gorm:"type:varchar(64);not null" json:"method"`
	SendCount uint   `gorm:"type:integer" json:"sendCount"`
}

type AlertTask struct {
	BaseModel
	Type      string `gorm:"type:varchar(64);not null" json:"type"`
	Quota     string `gorm:"type:varchar(64)" json:"quota"`
	QuotaType string `gorm:"type:varchar(64)" json:"quotaType"`
	Method    string `gorm:"type:varchar(64);not null;default:'sms" json:"method"`
}

type AlertLog struct {
	BaseModel

	Type        string `gorm:"type:varchar(64);not null" json:"type"`
	Status      string `gorm:"type:varchar(64);not null" json:"status"`
	AlertId     uint   `gorm:"type:integer;not null" json:"alertId"`
	AlertDetail string `gorm:"type:varchar(256);not null" json:"alertDetail"`
	AlertRule   string `gorm:"type:varchar(256);not null" json:"alertRule"`
	Count       uint   `gorm:"type:integer;not null" json:"count"`
	Message     string `gorm:"type:varchar(256);" json:"message"`
	RecordId    uint   `gorm:"type:integer;" json:"recordId"`
	LicenseId   string `gorm:"type:varchar(256);not null;" json:"licenseId" `
	Method      string `gorm:"type:varchar(64);not null;default:'sms" json:"method"`
}

type AlertConfig struct {
	BaseModel
	Type   string `gorm:"type:varchar(64);not null" json:"type"`
	Title  string `gorm:"type:varchar(64);not null" json:"title"`
	Status string `gorm:"type:varchar(64);not null" json:"status"`
	Config string `gorm:"type:varchar(256);not null" json:"config"`
}
