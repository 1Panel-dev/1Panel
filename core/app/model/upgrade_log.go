package model

type UpgradeLog struct {
	BaseModel
	NodeID     uint   `json:"nodeID"`
	OldVersion string `json:"oldVersion"`
	NewVersion string `json:"newVersion"`
	BackupFile string `json:"backupFile"`
}
