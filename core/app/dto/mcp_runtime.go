package dto

type MCPRuntimeSettings struct {
	RuntimeType  string   `json:"runtimeType"`
	DockerImage  string   `json:"dockerImage"`
	Enabled      bool     `json:"enabled"`
	ListenPort   int      `json:"listenPort"`
	ExternalURL  string   `json:"externalUrl"`
	Revision     uint64   `json:"revision"`
	AccessPolicy string   `json:"accessPolicy"`
	AllowedIPs   []string `json:"allowedIPs"`
}

type MCPRuntimeSettingsUpdate = MCPRuntimeSettings

type MCPRuntimeInstall struct {
	DockerImage  string                    `json:"dockerImage"`
	ListenPort   int                       `json:"listenPort"`
	ExternalURL  string                    `json:"externalUrl"`
	Revision     uint64                    `json:"revision"`
	AccessPolicy string                    `json:"accessPolicy"`
	AllowedIPs   []string                  `json:"allowedIPs"`
	Website      *MCPRuntimeInstallWebsite `json:"website,omitempty"`
}

type MCPRuntimeInstallWebsite struct {
	Domain    string `json:"domain"`
	EnableSSL bool   `json:"enableSSL"`
	SSLID     uint   `json:"sslID"`
}

type MCPRuntimeInstallTask struct {
	TaskID string `json:"taskID"`
}

type MCPRuntimeStatus struct {
	RuntimeType      string            `json:"runtimeType"`
	DockerImage      string            `json:"dockerImage"`
	ContainerName    string            `json:"containerName"`
	ImageID          string            `json:"imageId"`
	Enabled          bool              `json:"enabled"`
	Installed        bool              `json:"installed"`
	Status           string            `json:"status"`
	Ready            bool              `json:"ready"`
	Reason           string            `json:"reason"`
	AdapterVersion   string            `json:"adapterVersion"`
	ContractVersion  int               `json:"contractVersion"`
	Protocol         string            `json:"protocol"`
	InternalEndpoint string            `json:"internalEndpoint"`
	Endpoint         string            `json:"endpoint"`
	CanManageService bool              `json:"canManageService"`
	InstallTaskID    string            `json:"installTaskID,omitempty"`
	Installing       bool              `json:"installing"`
	Website          *MCPWebsiteStatus `json:"website,omitempty"`
}

type MCPRuntimeLogs struct {
	Content string `json:"content"`
}
