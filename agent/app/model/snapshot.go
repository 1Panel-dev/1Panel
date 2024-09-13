package model

type Snapshot struct {
	BaseModel
	Name              string `json:"name" gorm:"not null;unique"`
	Description       string `json:"description"`
	SourceAccountIDs  string `json:"sourceAccountIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`
	Status            string `json:"status"`
	Message           string `json:"message"`
	Version           string `json:"version"`

	AppData          string `json:"appData"`
	PanelData        string `json:"panelData"`
	BackupData       string `json:"backupData"`
	WithMonitorData  bool   `json:"withMonitorData"`
	WithLoginLog     bool   `json:"withLoginLog"`
	WithOperationLog bool   `json:"withOperationLog"`

	InterruptStep   string `json:"interruptStep"`
	RecoverStatus   string `json:"recoverStatus"`
	RecoverMessage  string `json:"recoverMessage"`
	LastRecoveredAt string `json:"lastRecoveredAt"`
	RollbackStatus  string `json:"rollbackStatus"`
	RollbackMessage string `json:"rollbackMessage"`
	LastRollbackAt  string `json:"lastRollbackAt"`
}

type SnapshotStatus struct {
	BaseModel
	SnapID uint `json:"snapID"`

	BaseData   string `json:"baseData" gorm:"default:Running"`
	AppImage   string `json:"appImage" gorm:"default:Running"`
	PanelData  string `json:"panelData" gorm:"default:Running"`
	BackupData string `json:"backupData" gorm:"default:Running"`

	Compress string `json:"compress" gorm:"default:Waiting"`
	Size     string `json:"size" `
	Upload   string `json:"upload" gorm:"default:Waiting"`
}
