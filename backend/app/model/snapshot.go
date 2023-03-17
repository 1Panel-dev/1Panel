package model

type Snapshot struct {
	BaseModel
	Name        string `json:"name" gorm:"type:varchar(64);not null;unique"`
	Description string `json:"description" gorm:"type:varchar(256)"`
	From        string `json:"from"`
	Status      string `json:"status" gorm:"type:varchar(64)"`
	Message     string `json:"message" gorm:"type:varchar(256)"`
	Version     string `json:"version" gorm:"type:varchar(256)"`

	InterruptStep    string `json:"interruptStep" gorm:"type:varchar(64)"`
	RecoverStatus    string `json:"recoverStatus" gorm:"type:varchar(64)"`
	RecoverMessage   string `json:"recoverMessage" gorm:"type:varchar(256)"`
	LastRecoveredAt  string `json:"lastRecoveredAt" gorm:"type:varchar(64)"`
	RollbackStatus   string `json:"rollbackStatus" gorm:"type:varchar(64)"`
	RollbackMessage  string `json:"rollbackMessage" gorm:"type:varchar(256)"`
	LastRollbackedAt string `json:"lastRollbackedAt" gorm:"type:varchar(64)"`
}
