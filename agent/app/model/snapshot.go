package model

type Snapshot struct {
	BaseModel
	Name              string `json:"name" gorm:"not null;unique"`
	Secret            string `json:"secret"`
	Description       string `json:"description"`
	SourceAccountIDs  string `json:"sourceAccountIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`
	Status            string `json:"status"`
	Message           string `json:"message"`
	Version           string `json:"version"`

	TaskID         string `json:"taskID"`
	TaskRecoverID  string `json:"taskRecoverID"`
	TaskRollbackID string `json:"taskRollbackID"`

	AppData          string `json:"appData"`
	PanelData        string `json:"panelData"`
	BackupData       string `json:"backupData"`
	WithMonitorData  bool   `json:"withMonitorData"`
	WithLoginLog     bool   `json:"withLoginLog"`
	WithOperationLog bool   `json:"withOperationLog"`
	WithSystemLog    bool   `json:"withSystemLog"`
	WithTaskLog      bool   `json:"withTaskLog"`
	IgnoreFiles      string `json:"ignoreFiles"`

	InterruptStep   string `json:"interruptStep"`
	RecoverStatus   string `json:"recoverStatus"`
	RecoverMessage  string `json:"recoverMessage"`
	RollbackStatus  string `json:"rollbackStatus"`
	RollbackMessage string `json:"rollbackMessage"`
}
