package dto

import "time"

type SettingInfo struct {
	SystemIP       string `json:"systemIP"`
	DockerSockPath string `json:"dockerSockPath"`
	SystemVersion  string `json:"systemVersion"`

	LocalTime string `json:"localTime"`
	TimeZone  string `json:"timeZone"`
	NtpSite   string `json:"ntpSite"`

	DefaultNetwork string `json:"defaultNetwork"`
	LastCleanTime  string `json:"lastCleanTime"`
	LastCleanSize  string `json:"lastCleanSize"`
	LastCleanData  string `json:"lastCleanData"`

	MonitorStatus    string `json:"monitorStatus"`
	MonitorInterval  string `json:"monitorInterval"`
	MonitorStoreDays string `json:"monitorStoreDays"`

	AppStoreVersion      string `json:"appStoreVersion"`
	AppStoreLastModified string `json:"appStoreLastModified"`
	AppStoreSyncStatus   string `json:"appStoreSyncStatus"`

	FileRecycleBin string `json:"fileRecycleBin"`

	SnapshotIgnore string `json:"snapshotIgnore"`
}

type SettingUpdate struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type SnapshotStatus struct {
	Panel      string `json:"panel"`
	PanelInfo  string `json:"panelInfo"`
	DaemonJson string `json:"daemonJson"`
	AppData    string `json:"appData"`
	PanelData  string `json:"panelData"`
	BackupData string `json:"backupData"`

	Compress string `json:"compress"`
	Size     string `json:"size"`
	Upload   string `json:"upload"`
}

type SnapshotCreate struct {
	ID                uint   `json:"id"`
	SourceAccountIDs  string `json:"sourceAccountIDs" validate:"required"`
	DownloadAccountID uint   `json:"downloadAccountID" validate:"required"`
	Description       string `json:"description" validate:"max=256"`
	Secret            string `json:"secret"`
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

type SyncTime struct {
	NtpSite string `json:"ntpSite" validate:"required"`
}

type CleanData struct {
	SystemClean    []CleanTree `json:"systemClean"`
	UploadClean    []CleanTree `json:"uploadClean"`
	DownloadClean  []CleanTree `json:"downloadClean"`
	SystemLogClean []CleanTree `json:"systemLogClean"`
	ContainerClean []CleanTree `json:"containerClean"`
}

type CleanTree struct {
	ID       string      `json:"id"`
	Label    string      `json:"label"`
	Children []CleanTree `json:"children"`

	Type string `json:"type"`
	Name string `json:"name"`

	Size        uint64 `json:"size"`
	IsCheck     bool   `json:"isCheck"`
	IsRecommend bool   `json:"isRecommend"`
}

type Clean struct {
	TreeType string `json:"treeType"`
	Name     string `json:"name"`
	Size     uint64 `json:"size"`
}
