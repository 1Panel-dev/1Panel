package dto

import "time"

type SettingInfo struct {
	UserName       string `json:"userName"`
	Email          string `json:"email"`
	SystemIP       string `json:"systemIP"`
	SystemVersion  string `json:"systemVersion"`
	DockerSockPath string `json:"dockerSockPath"`
	DeveloperMode  string `json:"developerMode"`

	SessionTimeout string `json:"sessionTimeout"`
	LocalTime      string `json:"localTime"`
	TimeZone       string `json:"timeZone"`
	NtpSite        string `json:"ntpSite"`

	Port           string `json:"port"`
	Ipv6           string `json:"ipv6"`
	BindAddress    string `json:"bindAddress"`
	PanelName      string `json:"panelName"`
	Theme          string `json:"theme"`
	MenuTabs       string `json:"menuTabs"`
	Language       string `json:"language"`
	DefaultNetwork string `json:"defaultNetwork"`
	LastCleanTime  string `json:"lastCleanTime"`
	LastCleanSize  string `json:"lastCleanSize"`
	LastCleanData  string `json:"lastCleanData"`

	ServerPort             string `json:"serverPort"`
	SSL                    string `json:"ssl"`
	SSLType                string `json:"sslType"`
	AutoRestart            string `json:"autoRestart"`
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
	AppStoreSyncStatus   string `json:"appStoreSyncStatus"`

	FileRecycleBin string `json:"fileRecycleBin"`

	SnapshotIgnore string `json:"snapshotIgnore"`
	XpackHideMenu  string `json:"xpackHideMenu"`
	NoAuthSetting  string `json:"noAuthSetting"`

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
	LicenseVerify      string `json:"licenseVerify"`
}

type SettingUpdate struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type SSLUpdate struct {
	SSLType     string `json:"sslType" validate:"required,oneof=self select import import-paste import-local"`
	Domain      string `json:"domain"`
	SSL         string `json:"ssl" validate:"required,oneof=enable disable"`
	Cert        string `json:"cert"`
	Key         string `json:"key"`
	SSLID       uint   `json:"sslID"`
	AutoRestart string `json:"autoRestart"`
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

type PageSnapshot struct {
	PageInfo
	Info    string `json:"info"`
	OrderBy string `json:"orderBy" validate:"required,oneof=name created_at"`
	Order   string `json:"order" validate:"required,oneof=null ascending descending"`
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
	CreatedAt       time.Time `json:"created_at"`
	Version         string    `json:"version"`

	InterruptStep    string `json:"interruptStep"`
	RecoverStatus    string `json:"recoverStatus"`
	RecoverMessage   string `json:"recoverMessage"`
	LastRecoveredAt  string `json:"lastRecoveredAt"`
	RollbackStatus   string `json:"rollbackStatus"`
	RollbackMessage  string `json:"rollbackMessage"`
	LastRollbackedAt string `json:"lastRollbackedAt"`
}

type SnapshotFile struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
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
	Ipv6        string `json:"ipv6" validate:"required,oneof=enable disable"`
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

type XpackHideMenu struct {
	ID       string          `json:"id"`
	Label    string          `json:"label"`
	IsCheck  bool            `json:"isCheck"`
	Title    string          `json:"title"`
	Path     string          `json:"path,omitempty"`
	Children []XpackHideMenu `json:"children,omitempty"`
}

type ApiInterfaceConfig struct {
	ApiInterfaceStatus string `json:"apiInterfaceStatus"`
	ApiKey             string `json:"apiKey"`
	IpWhiteList        string `json:"ipWhiteList"`
	ApiKeyValidityTime string `json:"apiKeyValidityTime"`
}
