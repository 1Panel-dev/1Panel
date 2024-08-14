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
	SnapID     uint   `json:"snapID"`
	Panel      string `json:"panel" gorm:"default:Running"`
	PanelInfo  string `json:"panelInfo" gorm:"default:Running"`
	DaemonJson string `json:"daemonJson" gorm:"default:Running"`
	AppData    string `json:"appData" gorm:"default:Running"`
	PanelData  string `json:"panelData" gorm:"default:Running"`
	BackupData string `json:"backupData" gorm:"default:Running"`

	Compress string `json:"compress" gorm:"default:Waiting"`
	Size     string `json:"size" `
	Upload   string `json:"upload" gorm:"default:Waiting"`
}
