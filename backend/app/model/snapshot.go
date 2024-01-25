package model

type Snapshot struct {
	BaseModel
	Name            string `json:"name" gorm:"type:varchar(64);not null;unique"`
	Description     string `json:"description" gorm:"type:varchar(256)"`
	From            string `json:"from"`
	DefaultDownload string `json:"defaultDownload" gorm:"type:varchar(64)"`
	Status          string `json:"status" gorm:"type:varchar(64)"`
	Message         string `json:"message" gorm:"type:varchar(256)"`
	Version         string `json:"version" gorm:"type:varchar(256)"`

	InterruptStep    string `json:"interruptStep" gorm:"type:varchar(64)"`
	RecoverStatus    string `json:"recoverStatus" gorm:"type:varchar(64)"`
	RecoverMessage   string `json:"recoverMessage" gorm:"type:varchar(256)"`
	LastRecoveredAt  string `json:"lastRecoveredAt" gorm:"type:varchar(64)"`
	RollbackStatus   string `json:"rollbackStatus" gorm:"type:varchar(64)"`
	RollbackMessage  string `json:"rollbackMessage" gorm:"type:varchar(256)"`
	LastRollbackedAt string `json:"lastRollbackedAt" gorm:"type:varchar(64)"`
}

type SnapshotStatus struct {
	BaseModel
	SnapID     uint   `gorm:"type:decimal" json:"snapID"`
	Panel      string `json:"panel" gorm:"type:varchar(64);default:Running"`
	PanelInfo  string `json:"panelInfo" gorm:"type:varchar(64);default:Running"`
	DaemonJson string `json:"daemonJson" gorm:"type:varchar(64);default:Running"`
	AppData    string `json:"appData" gorm:"type:varchar(64);default:Running"`
	PanelData  string `json:"panelData" gorm:"type:varchar(64);default:Running"`
	BackupData string `json:"backupData" gorm:"type:varchar(64);default:Running"`

	Compress string `json:"compress" gorm:"type:varchar(64);default:Waiting"`
	Size     string `json:"size" gorm:"type:varchar(64)"`
	Upload   string `json:"upload" gorm:"type:varchar(64);default:Waiting"`
}
