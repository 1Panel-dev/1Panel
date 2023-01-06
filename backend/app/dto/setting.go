package dto

import "time"

type SettingInfo struct {
	UserName string `json:"userName"`
	Email    string `json:"email"`

	SessionTimeout string `json:"sessionTimeout"`
	LocalTime      string `json:"localTime"`

	PanelName string `json:"panelName"`
	Theme     string `json:"theme"`
	Language  string `json:"language"`

	ServerPort             string `json:"serverPort"`
	SecurityEntrance       string `json:"securityEntrance"`
	ExpirationDays         string `json:"expirationDays"`
	ExpirationTime         string `json:"expirationTime"`
	ComplexityVerification string `json:"complexityVerification"`
	MFAStatus              string `json:"mfaStatus"`
	MFASecret              string `json:"mfaSecret"`

	MonitorStatus    string `json:"monitorStatus"`
	MonitorStoreDays string `json:"monitorStoreDays"`

	MessageType string `json:"messageType"`
	EmailVars   string `json:"emailVars"`
	WeChatVars  string `json:"weChatVars"`
	DingVars    string `json:"dingVars"`
}

type SettingUpdate struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type PasswordUpdate struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

type SnapshotCreate struct {
	BackupType  string `json:"backupType" validate:"required,oneof=OSS S3 SFTP MINIO"`
	Description string `json:"description"`
}
type SnapshotInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BackupType  string    `json:"backupType"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"createdAt"`
}
