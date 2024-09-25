package dto

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
