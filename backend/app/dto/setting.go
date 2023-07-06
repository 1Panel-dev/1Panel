package dto

import "time"

type SettingInfo struct {
	UserName      string `json:"userName"`
	Email         string `json:"email"`
	SystemIP      string `json:"systemIP"`
	SystemVersion string `json:"systemVersion"`

	SessionTimeout string `json:"sessionTimeout"`
	LocalTime      string `json:"localTime"`
	TimeZone       string `json:"timeZone"`
	NtpSite        string `json:"ntpSite"`

	Port      string `json:"port"`
	PanelName string `json:"panelName"`
	Theme     string `json:"theme"`
	Language  string `json:"language"`

	ServerPort             string `json:"serverPort"`
	SSL                    string `json:"ssl"`
	SSLType                string `json:"sslType"`
	BindDomain             string `json:"bindDomain"`
	AllowIPs               string `json:"allowIPs"`
	SecurityEntrance       string `json:"securityEntrance"`
	ExpirationDays         string `json:"expirationDays"`
	ExpirationTime         string `json:"expirationTime"`
	ComplexityVerification string `json:"complexityVerification"`
	MFAStatus              string `json:"mfaStatus"`
	MFASecret              string `json:"mfaSecret"`
	MFAInterval            string `json:"mfaInterval"`

	MonitorStatus    string `json:"monitorStatus"`
	MonitorInterval  string `json:"monitorInterval"`
	MonitorStoreDays string `json:"monitorStoreDays"`

	MessageType string `json:"messageType"`
	EmailVars   string `json:"emailVars"`
	WeChatVars  string `json:"weChatVars"`
	DingVars    string `json:"dingVars"`

	AppStoreVersion      string `json:"appStoreVersion"`
	AppStoreLastModified string `json:"appStoreLastModified"`
}

type SettingUpdate struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type SSLUpdate struct {
	SSLType string `json:"sslType"`
	Domain  string `json:"domain"`
	SSL     string `json:"ssl" validate:"required,oneof=enable disable"`
	Cert    string `json:"cert"`
	Key     string `json:"key"`
	SSLID   uint   `json:"sslID"`
}
type SSLInfo struct {
	Domain   string `json:"domain"`
	Timeout  string `json:"timeout"`
	RootPath string `json:"rootPath"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	SSLID    uint   `json:"sslID"`
}

type PasswordUpdate struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

type PortUpdate struct {
	ServerPort uint `json:"serverPort" validate:"required,number,max=65535,min=1"`
}

type SnapshotCreate struct {
	From        string `json:"from" validate:"required,oneof=OSS S3 SFTP MINIO COS KODO OneDrive"`
	Description string `json:"description" validate:"max=256"`
}
type SnapshotRecover struct {
	IsNew      bool `json:"isNew"`
	ReDownload bool `json:"reDownload"`
	ID         uint `json:"id" validate:"required"`
}
type SnapshotImport struct {
	From        string   `json:"from"`
	Names       []string `json:"names"`
	Description string   `json:"description" validate:"max=256"`
}
type SnapshotInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description" validate:"max=256"`
	From        string    `json:"from"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"createdAt"`
	Version     string    `json:"version"`

	InterruptStep    string `json:"interruptStep"`
	RecoverStatus    string `json:"recoverStatus"`
	RecoverMessage   string `json:"recoverMessage"`
	LastRecoveredAt  string `json:"lastRecoveredAt"`
	RollbackStatus   string `json:"rollbackStatus"`
	RollbackMessage  string `json:"rollbackMessage"`
	LastRollbackedAt string `json:"lastRollbackedAt"`
}

type UpgradeInfo struct {
	NewVersion    string `json:"newVersion"`
	LatestVersion string `json:"latestVersion"`
	ReleaseNote   string `json:"releaseNote"`
}

type SyncTime struct {
	NtpSite string `json:"ntpSite"`
}

type Upgrade struct {
	Version string `json:"version"`
}
