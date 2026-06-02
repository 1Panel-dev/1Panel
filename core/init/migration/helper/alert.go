package helper

type AlertAuditAlert struct {
	ID         uint   `gorm:"primaryKey"`
	CreateUser string `gorm:"type:varchar(256)"`
	UpdateUser string `gorm:"type:varchar(256)"`
}

func (AlertAuditAlert) TableName() string {
	return "alerts"
}

type AlertAuditConfig struct {
	ID         uint   `gorm:"primaryKey"`
	CreateUser string `gorm:"type:varchar(256)"`
	UpdateUser string `gorm:"type:varchar(256)"`
}

func (AlertAuditConfig) TableName() string {
	return "alert_configs"
}
