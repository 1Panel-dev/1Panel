package dto

import (
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

type FirewallSubsystemStatus struct {
	IPv6Enabled     bool                        `json:"ipv6Enabled"`
	Name            string                      `json:"name"`
	Backend         string                      `json:"backend"`
	ConflictBackend string                      `json:"conflictBackend,omitempty"`
	IsExist         bool                        `json:"isExist"`
	IsActive        bool                        `json:"isActive"`
	IsInit          bool                        `json:"isInit"`
	IsBind          bool                        `json:"isBind"`
	Version         string                      `json:"version"`
	PingStatus      string                      `json:"pingStatus"`
	Message         string                      `json:"message,omitempty"`
	Reason          string                      `json:"reason,omitempty"`
	LifecycleTaskID string                      `json:"lifecycleTaskID,omitempty"`
	IPv4            FirewallBackendFamilyStatus `json:"ipv4"`
	IPv6            FirewallBackendFamilyStatus `json:"ipv6"`
}

type FirewallLifecycleOperation struct {
	Operation         string `json:"operation" validate:"required,oneof=start stop restart disableBanPing enableBanPing"`
	WithDockerRestart bool   `json:"withDockerRestart"`
}

type FirewallLifecycleOperationResponse struct {
	TaskID string `json:"taskID,omitempty"`
	Queued bool   `json:"queued"`
}

type FirewallBackendOption struct {
	Name           string                      `json:"name"`
	Installed      bool                        `json:"installed"`
	Active         bool                        `json:"active"`
	Initialized    bool                        `json:"initialized"`
	Bound          bool                        `json:"bound"`
	Supported      bool                        `json:"supported"`
	SupportReason  string                      `json:"supportReason,omitempty"`
	Implementation string                      `json:"implementation,omitempty"`
	Message        string                      `json:"message,omitempty"`
	IPv4           FirewallBackendFamilyStatus `json:"ipv4"`
	IPv6           FirewallBackendFamilyStatus `json:"ipv6"`
}

type FirewallBackendFamilyStatus struct {
	Partial       bool     `json:"partial"`
	Available     bool     `json:"available"`
	Initialized   bool     `json:"initialized"`
	Bound         bool     `json:"bound"`
	Reason        string   `json:"reason,omitempty"`
	ForwardPolicy string   `json:"forwardPolicy,omitempty"`
	RAInterfaces  []string `json:"raInterfaces,omitempty"`
}

type FirewallBackendGroup struct {
	Selected string                  `json:"selected"`
	Current  string                  `json:"current,omitempty"`
	Options  []FirewallBackendOption `json:"options"`
}

type FirewallSettings struct {
	IPv6Enabled   bool                   `json:"ipv6Enabled"`
	System        FirewallBackendGroup   `json:"system"`
	Forwarding    FirewallBackendGroup   `json:"forwarding"`
	Docker        FirewallBackendGroup   `json:"docker"`
	PingStatus    string                 `json:"pingStatus"`
	PortWhitelist []filter.PortWhitelist `json:"portWhiteList"`
	PanelPort     string                 `json:"panelPort"`
	SSHPort       string                 `json:"sshPort"`
}

type FirewallPortWhitelistCreate struct {
	Rule filter.PortWhitelist `json:"rule" validate:"required"`
}

type FirewallPortWhitelistUpdate struct {
	OldRule filter.PortWhitelist `json:"oldRule" validate:"required"`
	Rule    filter.PortWhitelist `json:"rule" validate:"required"`
}

type FirewallPortWhitelistDelete struct {
	Rule *filter.PortWhitelist `json:"rule" validate:"required"`
}

type FirewallBackendOperation struct {
	Subsystem string `json:"subsystem" validate:"required,oneof=system forwarding docker"`
	Backend   string `json:"backend" validate:"required,oneof=firewalld ufw iptables nftables"`
	Operation string `json:"operation" validate:"required,oneof=select initialize cleanup"`
}

type FirewallIPv6Operation struct {
	Status string `json:"status" validate:"required,oneof=Enable Disable"`
}

type FirewallFamilyOperation struct {
	Subsystem string `json:"subsystem" validate:"required,oneof=system forwarding docker"`
	Backend   string `json:"backend" validate:"required,oneof=iptables nftables"`
	Family    string `json:"family" validate:"required,oneof=ipv4 ipv6"`
	Operation string `json:"operation" validate:"required,oneof=initialize repair bind"`
}

type FilterChainOperation struct {
	Name    string `json:"name" validate:"required,eq=1PANEL_BASIC"`
	Operate string `json:"operate" validate:"required,oneof=init-base bind-base unbind-base"`
	TaskID  string `json:"taskID,omitempty" validate:"omitempty,max=64"`
}

type FilterChainOperationResponse struct {
	TaskID string `json:"taskID"`
	Queued bool   `json:"queued"`
}

type FirewallInitializationTask struct {
	BackupFile string `json:"backupFile,omitempty" validate:"omitempty,max=255"`
	TaskID     string `json:"taskID,omitempty" validate:"omitempty,max=64"`
}

type FirewallSystemPort = firewall.SystemPort

type FirewallRuleInventoryResponse struct {
	IPv4Range filter.PositionRange   `json:"ipv4Range"`
	IPv6Range filter.PositionRange   `json:"ipv6Range"`
	Total     int64                  `json:"total"`
	AllTotal  int64                  `json:"allTotal"`
	Items     []filter.InventoryItem `json:"items"`
	Notices   []filter.ScopeNotice   `json:"notices,omitempty"`
}

type FirewallRuleBackup struct {
	Name       string          `json:"name"`
	Provider   filter.Provider `json:"provider"`
	RuleCount  int             `json:"ruleCount"`
	ModifiedAt int64           `json:"modifiedAt"`
}

type FirewallRuleBackups struct {
	Directory string               `json:"directory"`
	Files     []FirewallRuleBackup `json:"files"`
}

type FirewallRuleResetResponse struct {
	BackupPath string `json:"backupPath"`
	Removed    int    `json:"removed"`
	Disabled   bool   `json:"disabled"`
}

type FirewallRuleReset struct {
	Subsystem         string          `json:"subsystem,omitempty" validate:"omitempty,oneof=system forwarding docker"`
	Backup            *bool           `json:"backup,omitempty" default:"true"`
	Provider          filter.Provider `json:"provider,omitempty" validate:"omitempty,oneof=firewalld ufw iptables nftables"`
	WithDockerRestart bool            `json:"withDockerRestart"`
}

type FirewallRuleInventory struct {
	PageInfo
	Scope         filter.Scope    `json:"scope,omitempty"`
	Scopes        []filter.Scope  `json:"scopes,omitempty" validate:"max=16"`
	All           bool            `json:"all,omitempty"`
	Info          string          `json:"info"`
	Families      []filter.Family `json:"families,omitempty" validate:"omitempty,dive,oneof=ipv4 ipv6"`
	Actions       []string        `json:"actions,omitempty" validate:"omitempty,dive,oneof=accept deny"`
	ExcludeChains []string        `json:"excludeChains,omitempty" validate:"omitempty,dive,oneof=1PANEL_BASIC_BEFORE 1PANEL_BASIC 1PANEL_BASIC_AFTER"`
}

type FirewallNativeDetail struct {
	Provider   filter.Provider   `json:"provider" validate:"required,oneof=firewalld ufw"`
	NativeKind filter.NativeKind `json:"nativeKind" validate:"required,oneof=zone_service ufw_application"`
	Name       string            `json:"name" validate:"required"`
	Permanent  bool              `json:"permanent"`
}

type DockerPortGuardBase struct {
	IPv6Enabled bool                        `json:"ipv6Enabled"`
	Name        string                      `json:"name"`
	Version     string                      `json:"version"`
	IsExist     bool                        `json:"isExist"`
	Initialized bool                        `json:"initialized"`
	Bound       bool                        `json:"bound"`
	IPv4        DockerPortGuardFamilyStatus `json:"ipv4"`
	IPv6        DockerPortGuardFamilyStatus `json:"ipv6"`
	Backend     string                      `json:"backend"`
	Message     string                      `json:"message,omitempty"`
}

type DockerPortGuardFamilyStatus struct {
	Partial     bool   `json:"partial"`
	State       string `json:"state"`
	Reason      string `json:"reason,omitempty"`
	Initialized bool   `json:"initialized"`
	Bound       bool   `json:"bound"`
	Effective   bool   `json:"effective"`
}

type DockerPortGuardEndpoint struct {
	Family           string   `json:"family"`
	HostIP           string   `json:"hostIP"`
	HostPort         uint16   `json:"hostPort"`
	Protocol         string   `json:"protocol"`
	ContainerID      string   `json:"containerID"`
	ContainerName    string   `json:"containerName"`
	ContainerState   string   `json:"containerState,omitempty"`
	ContainerPort    uint16   `json:"containerPort"`
	Compose          string   `json:"compose,omitempty"`
	Application      string   `json:"application,omitempty"`
	PolicyUUID       string   `json:"policyUUID,omitempty"`
	Mode             string   `json:"mode,omitempty"`
	Sources          []string `json:"sources"`
	Effective        bool     `json:"effective"`
	TrafficPath      string   `json:"trafficPath"`
	ManagementTarget string   `json:"managementTarget"`
	ManagementReason string   `json:"managementReason,omitempty"`
}

type DockerPortGuardPortGroup struct {
	Key       string                    `json:"key"`
	Label     string                    `json:"label"`
	Endpoint  DockerPortGuardEndpoint   `json:"endpoint"`
	Endpoints []DockerPortGuardEndpoint `json:"endpoints"`
}

type DockerPortGuardContainer struct {
	Key         string                     `json:"key"`
	Name        string                     `json:"name"`
	Compose     string                     `json:"compose,omitempty"`
	Application string                     `json:"application,omitempty"`
	Endpoints   []DockerPortGuardEndpoint  `json:"endpoints"`
	PortGroups  []DockerPortGuardPortGroup `json:"portGroups"`
}

type DockerPortGuardList struct {
	Base           DockerPortGuardBase        `json:"base"`
	Containers     []DockerPortGuardContainer `json:"containers"`
	OrphanPolicies []DockerPortGuardEndpoint  `json:"orphanPolicies"`
}

type DockerPortGuardEndpointIdentity struct {
	Family   string `json:"family" validate:"required,oneof=ipv4 ipv6"`
	HostIP   string `json:"hostIP" validate:"required,max=45"`
	HostPort uint16 `json:"hostPort" validate:"required,min=1"`
	Protocol string `json:"protocol" validate:"required,oneof=tcp udp"`
}

type DockerPortGuardPolicyBatch struct {
	Policies []DockerPortGuardPolicy `json:"policies" validate:"required,min=1,dive"`
	Import   bool                    `json:"import"`
}

type DockerPortGuardPolicyBatchDelete struct {
	UUIDs []string `json:"uuids" validate:"required,min=1,dive,required,max=64"`
}

type DockerPortGuardPolicy struct {
	DockerPortGuardEndpointIdentity
	Mode    string   `json:"mode" validate:"required,oneof=deny_sources allow_sources deny_all accept_sources accept_all"`
	Sources []string `json:"sources" validate:"dive,required,max=64"`
}

type DockerPortGuardOperation struct {
	BackupFile string `json:"backupFile,omitempty" validate:"omitempty,max=255"`
	Operation  string `json:"operation" validate:"required,oneof=initialize bind unbind"`
	TaskID     string `json:"taskID,omitempty" validate:"omitempty,max=64"`
}

type FirewallRuleCreateItem struct {
	Raw         string              `json:"raw,omitempty"`
	ParseStatus filter.ParseStatus  `json:"parseStatus,omitempty"`
	Rule        filter.FirewallRule `json:"rule" validate:"required"`
	SourceKind  string              `json:"sourceKind" validate:"omitempty,oneof=user imported"`
}

type FirewallRuleCreate struct {
	BackupFile string                   `json:"backupFile,omitempty" validate:"omitempty,max=255"`
	Initialize bool                     `json:"initialize"`
	Items      []FirewallRuleCreateItem `json:"items" validate:"dive"`
}

type FirewallRuleCreateResponse struct {
	TaskID    string                      `json:"taskID,omitempty"`
	Queued    bool                        `json:"queued,omitempty"`
	Succeeded int                         `json:"succeeded"`
	Failed    int                         `json:"failed"`
	Skipped   int                         `json:"skipped"`
	Errors    []FirewallRuleCreateFailure `json:"errors,omitempty"`
}

type FirewallRuleCreateFailure struct {
	Index  int                 `json:"index"`
	Status string              `json:"status"`
	Rule   filter.FirewallRule `json:"rule"`
	Error  string              `json:"error,omitempty"`
}

type FirewallRuleDelete struct {
	Targets []FirewallRuleDeleteItem `json:"targets" validate:"required,min=1,dive"`
}

type FirewallRuleDeleteItem struct {
	FirewallRuleDeleteTarget
	Observed filter.ObservedRule `json:"observed" validate:"required"`
}

type FirewallRuleDeleteTarget struct {
	Scope       filter.Scope `json:"scope" validate:"required"`
	InstanceKey string       `json:"instanceKey" validate:"required,max=128"`
}

type FirewallRuleDeleteResponse struct {
	TaskID    string                      `json:"taskID,omitempty"`
	Queued    bool                        `json:"queued,omitempty"`
	Succeeded int                         `json:"succeeded"`
	Failed    int                         `json:"failed"`
	Errors    []FirewallRuleDeleteFailure `json:"errors,omitempty"`
}

type FirewallRuleDeleteFailure struct {
	Index       int    `json:"index"`
	InstanceKey string `json:"instanceKey"`
	Error       string `json:"error"`
}

type FirewallRuleUpdate struct {
	FirewallRuleDeleteTarget
	Rule        *filter.FirewallRule `json:"rule,omitempty" validate:"required_without_all=Description OrderIndex Priority,excluded_with=Description OrderIndex Priority"`
	Description *string              `json:"description,omitempty" validate:"excluded_with=Rule"`
	OrderIndex  *int64               `json:"orderIndex,omitempty" validate:"excluded_with=Rule Priority"`
	Priority    *int                 `json:"priority,omitempty" validate:"excluded_with=Rule OrderIndex"`
}

type FirewallRuleReorder struct {
	FirewallRuleDeleteTarget
	TargetPosition *int64 `json:"targetPosition"`
	Priority       *int   `json:"priority"`
}

type FirewallRuleExportItem struct {
	filter.FirewallRule
	Raw         string             `json:"raw,omitempty"`
	ParseStatus filter.ParseStatus `json:"parseStatus,omitempty"`
}

type FirewallSubsystemBackup struct {
	Families   []string                        `json:"families,omitempty"`
	Subsystem  string                          `json:"subsystem"`
	Provider   filter.Provider                 `json:"provider"`
	Forwarding []forwarding.Rule               `json:"forwarding"`
	Docker     *dockerfirewall.PolicyInventory `json:"docker,omitempty"`
}
