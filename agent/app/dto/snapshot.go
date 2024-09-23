package dto

import "time"

type SnapshotStatus struct {
	BaseData   string `json:"baseData"`
	AppImage   string `json:"appImage"`
	PanelData  string `json:"panelData"`
	BackupData string `json:"backupData"`

	Compress string `json:"compress"`
	Size     string `json:"size"`
	Upload   string `json:"upload"`
}

type SnapshotCreate struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	TaskID            string `json:"taskID"`
	SourceAccountIDs  string `json:"sourceAccountIDs" validate:"required"`
	DownloadAccountID uint   `json:"downloadAccountID" validate:"required"`
	Description       string `json:"description" validate:"max=256"`
	Secret            string `json:"secret"`
	InterruptStep     string `json:"interruptStep"`

	AppData    []DataTree `json:"appData"`
	BackupData []DataTree `json:"backupData"`
	PanelData  []DataTree `json:"panelData"`

	WithMonitorData  bool `json:"withMonitorData"`
	WithLoginLog     bool `json:"withLoginLog"`
	WithOperationLog bool `json:"withOperationLog"`
	WithSystemLog    bool `json:"withSystemLog"`
	WithTaskLog      bool `json:"withTaskLog"`
}

type SnapshotData struct {
	AppData    []DataTree `json:"appData"`
	BackupData []DataTree `json:"backupData"`
	PanelData  []DataTree `json:"panelData"`

	WithMonitorData  bool `json:"withMonitorData"`
	WithLoginLog     bool `json:"withLoginLog"`
	WithOperationLog bool `json:"withOperationLog"`
	WithSystemLog    bool `json:"withSystemLog"`
	WithTaskLog      bool `json:"withTaskLog"`
}
type DataTree struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	IsCheck   bool   `json:"isCheck"`
	IsDisable bool   `json:"isDisable"`

	Path string `json:"path"`

	RelationItemID string     `json:"relationItemID"`
	Children       []DataTree `json:"children"`
}
type SnapshotRecover struct {
	IsNew      bool   `json:"isNew"`
	ReDownload bool   `json:"reDownload"`
	ID         uint   `json:"id" validate:"required"`
	Secret     string `json:"secret"`
}
type SnapshotBatchDelete struct {
	DeleteWithFile bool   `json:"deleteWithFile"`
	Ids            []uint `json:"ids" validate:"required"`
}

type SnapshotImport struct {
	BackupAccountID uint     `json:"backupAccountID"`
	Names           []string `json:"names"`
	Description     string   `json:"description" validate:"max=256"`
}

type SnapshotInfo struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	TaskID          string    `json:"taskID"`
	Description     string    `json:"description" validate:"max=256"`
	From            string    `json:"from"`
	DefaultDownload string    `json:"defaultDownload"`
	Status          string    `json:"status"`
	Message         string    `json:"message"`
	CreatedAt       time.Time `json:"createdAt"`
	Version         string    `json:"version"`
	Size            int64     `json:"size"`

	InterruptStep    string `json:"interruptStep"`
	RecoverStatus    string `json:"recoverStatus"`
	RecoverMessage   string `json:"recoverMessage"`
	LastRecoveredAt  string `json:"lastRecoveredAt"`
	RollbackStatus   string `json:"rollbackStatus"`
	RollbackMessage  string `json:"rollbackMessage"`
	LastRollbackedAt string `json:"lastRollbackedAt"`
}
