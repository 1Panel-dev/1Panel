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
	WithSystemLog    bool   `json:"withSystemLog"`
	WithTaskLog      bool   `json:"withTaskLog"`

	InterruptStep   string `json:"interruptStep"`
	RecoverStatus   string `json:"recoverStatus"`
	RecoverMessage  string `json:"recoverMessage"`
	LastRecoveredAt string `json:"lastRecoveredAt"`
	RollbackStatus  string `json:"rollbackStatus"`
	RollbackMessage string `json:"rollbackMessage"`
	LastRollbackAt  string `json:"lastRollbackAt"`
}
