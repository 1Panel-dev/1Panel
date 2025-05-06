package dto

import (
	"time"
)

type SettingInfo struct {
	UserName      string `json:"userName"`
	SystemVersion string `json:"systemVersion"`
	DeveloperMode string `json:"developerMode"`

	SessionTimeout string `json:"sessionTimeout"`
	Port           string `json:"port"`
	Ipv6           string `json:"ipv6"`
	BindAddress    string `json:"bindAddress"`
	PanelName      string `json:"panelName"`
	Theme          string `json:"theme"`
	MenuTabs       string `json:"menuTabs"`
	Language       string `json:"language"`
	SystemIP       string `json:"systemIP"`

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

	AppStoreVersion      string `json:"appStoreVersion"`
	AppStoreLastModified string `json:"appStoreLastModified"`
	AppStoreSyncStatus   string `json:"appStoreSyncStatus"`

	HideMenu      string `json:"hideMenu"`
	NoAuthSetting string `json:"noAuthSetting"`

	ProxyUrl        string `json:"proxyUrl"`
	ProxyType       string `json:"proxyType"`
	ProxyPort       string `json:"proxyPort"`
	ProxyUser       string `json:"proxyUser"`
	ProxyPasswd     string `json:"proxyPasswd"`
	ProxyPasswdKeep string `json:"proxyPasswdKeep"`

	ApiInterfaceStatus string `json:"apiInterfaceStatus"`
	ApiKey             string `json:"apiKey"`
	IpWhiteList        string `json:"ipWhiteList"`
	ApiKeyValidityTime string `json:"apiKeyValidityTime"`
}

type SettingUpdate struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type SSLUpdate struct {
	SSLType string `json:"sslType" validate:"required,oneof=self select import import-paste import-local"`
	Domain  string `json:"domain"`
	SSL     string `json:"ssl" validate:"required,oneof=Enable Disable"`
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
	ID              uint   `json:"id"`
	From            string `json:"from" validate:"required"`
	DefaultDownload string `json:"defaultDownload" validate:"required"`
	Description     string `json:"description" validate:"max=256"`
	Secret          string `json:"secret"`
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
	From        string   `json:"from"`
	Names       []string `json:"names"`
	Description string   `json:"description" validate:"max=256"`
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

type UpgradeInfo struct {
	TestVersion   string `json:"testVersion"`
	NewVersion    string `json:"newVersion"`
	LatestVersion string `json:"latestVersion"`
	ReleaseNote   string `json:"releaseNote"`
}

type SyncTime struct {
	NtpSite string `json:"ntpSite" validate:"required"`
}

type BindInfo struct {
	Ipv6        string `json:"ipv6" validate:"required,oneof=Enable Disable"`
	BindAddress string `json:"bindAddress" validate:"required"`
}

type Upgrade struct {
	Version string `json:"version" validate:"required"`
}

type ProxyUpdate struct {
	ProxyUrl        string `json:"proxyUrl"`
	ProxyType       string `json:"proxyType"`
	ProxyPort       string `json:"proxyPort"`
	ProxyUser       string `json:"proxyUser"`
	ProxyPasswd     string `json:"proxyPasswd"`
	ProxyPasswdKeep string `json:"proxyPasswdKeep"`
	ProxyDocker     bool   `json:"proxyDocker"`
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

type ShowMenu struct {
	ID       string     `json:"id"`
	Label    string     `json:"label"`
	Disabled bool       `json:"disabled"`
	IsShow   bool       `json:"isShow"`
	Title    string     `json:"title"`
	Path     string     `json:"path,omitempty"`
	Children []ShowMenu `json:"children,omitempty"`
}

type ApiInterfaceConfig struct {
	ApiInterfaceStatus string `json:"apiInterfaceStatus"`
	ApiKey             string `json:"apiKey"`
	IpWhiteList        string `json:"ipWhiteList"`
	ApiKeyValidityTime string `json:"apiKeyValidityTime"`
}

type TerminalInfo struct {
	LineHeight        string `json:"lineHeight"`
	LetterSpacing     string `json:"letterSpacing"`
	FontSize          string `json:"fontSize"`
	CursorBlink       string `json:"cursorBlink"`
	CursorStyle       string `json:"cursorStyle"`
	Scrollback        string `json:"scrollback"`
	ScrollSensitivity string `json:"scrollSensitivity"`
}

type AppstoreUpdate struct {
	Scope  string `json:"scope" validate:"required,oneof=UninstallDeleteImage UpgradeBackup UninstallDeleteBackup"`
	Status string `json:"status"  validate:"required,oneof=Disable Enable"`
}
type AppstoreConfig struct {
	UninstallDeleteImage  string `json:"uninstallDeleteImage"`
	UpgradeBackup         string `json:"upgradeBackup"`
	UninstallDeleteBackup string `json:"uninstallDeleteBackup"`
}
