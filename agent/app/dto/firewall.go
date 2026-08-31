package dto

import (
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	firewallsync "github.com/1Panel-dev/1Panel/agent/utils/firewall/sync"
)

type FirewallSubsystemStatus struct {
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
	SyncError       string                      `json:"syncError,omitempty"`
	IPv4            FirewallBackendFamilyStatus `json:"ipv4"`
	IPv6            FirewallBackendFamilyStatus `json:"ipv6"`
}

type FirewallLifecycleOperation struct {
	Operation         string `json:"operation" validate:"required,oneof=start stop restart disableBanPing enableBanPing"`
	WithDockerRestart bool   `json:"withDockerRestart"`
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
	Available   bool   `json:"available"`
	Initialized bool   `json:"initialized"`
	Bound       bool   `json:"bound"`
	Reason      string `json:"reason,omitempty"`
}

type FirewallBackendGroup struct {
	Selected string                  `json:"selected"`
	Current  string                  `json:"current,omitempty"`
	Options  []FirewallBackendOption `json:"options"`
}

type FirewallSettings struct {
	System        FirewallBackendGroup `json:"system"`
	Forwarding    FirewallBackendGroup `json:"forwarding"`
	Docker        FirewallBackendGroup `json:"docker"`
	PingStatus    string               `json:"pingStatus"`
	PortWhitelist string               `json:"portWhiteList"`
}

type FirewallBackendOperation struct {
	Subsystem string `json:"subsystem" validate:"required,oneof=system forwarding docker"`
	Backend   string `json:"backend" validate:"required,oneof=firewalld ufw iptables nftables"`
	Operation string `json:"operation" validate:"required,oneof=select initialize cleanup"`
}

type FilterChainOperation struct {
	Name    string `json:"name" validate:"required,eq=1PANEL_BASIC"`
	Operate string `json:"operate" validate:"required,oneof=init-base bind-base unbind-base"`
}

type FirewallSystemPort struct {
	Family   string
	Port     string
	Protocol string
}

type FirewallRuleInventoryResponse struct {
	Total        int64                  `json:"total"`
	AllTotal     int64                  `json:"allTotal"`
	ManagedTotal int64                  `json:"managedTotal"`
	Items        []filter.InventoryItem `json:"items"`
	Notices      []filter.ScopeNotice   `json:"notices,omitempty"`
}

type FirewallRuleResetResponse struct {
	Removed  int  `json:"removed"`
	Disabled bool `json:"disabled"`
}

type FirewallRuleReset struct {
	Provider filter.Provider `json:"provider,omitempty" validate:"omitempty,oneof=firewalld ufw iptables nftables"`
}

type FirewallRuleCheckResult struct {
	Decision         filter.CheckDecision       `json:"decision"`
	Classification   filter.CheckClassification `json:"classification"`
	Reason           string                     `json:"reason"`
	RequestedRule    filter.FirewallRule        `json:"requestedRule"`
	RequestedRuleKey string                     `json:"requestedRuleKey"`
	ExistingRuleUUID string                     `json:"existingRuleUUID,omitempty"`
	Candidates       []filter.ObservedRule      `json:"candidates,omitempty"`
	AllowedActions   []filter.CheckAction       `json:"allowedActions,omitempty"`
	CheckFlag        string                     `json:"checkFlag"`
}

type FirewallRuleInventory struct {
	PageInfo
	Scope         filter.Scope            `json:"scope,omitempty"`
	Scopes        []filter.Scope          `json:"scopes,omitempty" validate:"max=16"`
	All           bool                    `json:"all,omitempty"`
	Info          string                  `json:"info"`
	Families      []filter.Family         `json:"families,omitempty" validate:"omitempty,dive,oneof=ipv4 ipv6"`
	Actions       []string                `json:"actions,omitempty" validate:"omitempty,dive,oneof=accept deny"`
	States        []filter.InventoryState `json:"states,omitempty" validate:"omitempty,dive,oneof=managed adopted external drifted protected"`
	ExcludeChains []string                `json:"excludeChains,omitempty" validate:"omitempty,dive,oneof=1PANEL_BASIC_BEFORE 1PANEL_BASIC 1PANEL_BASIC_AFTER"`
}

type FirewallNativeDetail struct {
	Provider   filter.Provider   `json:"provider" validate:"required,oneof=firewalld ufw"`
	NativeKind filter.NativeKind `json:"nativeKind" validate:"required,oneof=zone_service ufw_application"`
	Name       string            `json:"name" validate:"required"`
	Permanent  bool              `json:"permanent"`
}

type DockerPortGuardBase struct {
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
	State       string `json:"state"`
	Reason      string `json:"reason,omitempty"`
	Initialized bool   `json:"initialized"`
	Bound       bool   `json:"bound"`
	Effective   bool   `json:"effective"`
}

type DockerPortGuardEndpoint struct {
	Family        string   `json:"family"`
	HostIP        string   `json:"hostIP"`
	HostPort      uint16   `json:"hostPort"`
	Protocol      string   `json:"protocol"`
	ContainerID   string   `json:"containerID"`
	ContainerName string   `json:"containerName"`
	ContainerPort uint16   `json:"containerPort"`
	Compose       string   `json:"compose,omitempty"`
	Application   string   `json:"application,omitempty"`
	PolicyUUID    string   `json:"policyUUID,omitempty"`
	Mode          string   `json:"mode,omitempty"`
	Sources       []string `json:"sources"`
	Effective     bool     `json:"effective"`
	Description   string   `json:"description,omitempty"`
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
	Endpoints   []DockerPortGuardEndpointIdentity `json:"endpoints" validate:"required,min=1,max=256,dive"`
	Mode        string                            `json:"mode" validate:"required,oneof=deny_sources allow_sources deny_all"`
	Sources     []string                          `json:"sources" validate:"max=256,dive,required,max=64"`
	Description string                            `json:"description" validate:"max=256"`
}

type DockerPortGuardPolicyBatchDelete struct {
	UUIDs []string `json:"uuids" validate:"required,min=1,max=256,dive,required,max=64"`
}

type DockerPortGuardOperation struct {
	Operation string `json:"operation" validate:"required,oneof=initialize bind unbind"`
}

type FirewallRuleCheckItem struct {
	UUID string              `json:"uuid" validate:"omitempty,max=64"`
	Rule filter.FirewallRule `json:"rule" validate:"required"`
}

type FirewallRuleCheck struct {
	Items []FirewallRuleCheckItem `json:"items" validate:"required,min=1,max=256,dive"`
}

type FirewallRuleCheckResponse struct {
	Items []FirewallRuleCheckResult `json:"items"`
}

type FirewallRuleCreateItem struct {
	Rule             filter.FirewallRule `json:"rule" validate:"required"`
	CheckFlag        string              `json:"checkFlag"`
	Action           filter.CheckAction  `json:"action"`
	AdoptInstanceKey string              `json:"adoptInstanceKey"`
	SourceKind       string              `json:"sourceKind" validate:"omitempty,oneof=user imported"`
	SourceID         string              `json:"sourceID"`
}

type FirewallRuleCreate struct {
	Items []FirewallRuleCreateItem `json:"items" validate:"required,min=1,max=256,dive"`
}

type FirewallRuleCreateResponse struct {
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

type FirewallRuleSyncRequest struct {
	Subsystem      string          `json:"subsystem" validate:"omitempty,oneof=system forwarding docker"`
	SourceProvider filter.Provider `json:"sourceProvider,omitempty" validate:"omitempty,oneof=firewalld ufw iptables nftables"`
	TargetProvider filter.Provider `json:"targetProvider" validate:"required,oneof=firewalld ufw iptables nftables"`
	ResetSource    bool            `json:"resetSource"`
	TaskID         string          `json:"taskID,omitempty" validate:"omitempty,max=64"`
}

type FirewallRuleSyncItem struct {
	SourceUUID  string                   `json:"sourceUUID"`
	Rule        *filter.FirewallRule     `json:"rule,omitempty"`
	ForwardRule *ForwardRule             `json:"forwardRule,omitempty"`
	DockerRule  *DockerPortGuardEndpoint `json:"dockerRule,omitempty"`
	Status      firewallsync.Status      `json:"status"`
	ReasonCode  firewallsync.ReasonCode  `json:"reasonCode,omitempty"`
	Reason      string                   `json:"reason,omitempty"`
}

type FirewallRuleSyncPreview struct {
	Subsystem      string                 `json:"subsystem"`
	SourceProvider filter.Provider        `json:"sourceProvider,omitempty"`
	TargetProvider filter.Provider        `json:"targetProvider"`
	Total          int                    `json:"total"`
	Ready          int                    `json:"ready"`
	Existing       int                    `json:"existing"`
	Removed        int                    `json:"removed"`
	Blocked        int                    `json:"blocked"`
	Items          []FirewallRuleSyncItem `json:"items"`
}

type FirewallRuleSyncResult struct {
	Subsystem      string                    `json:"subsystem"`
	SourceProvider filter.Provider           `json:"sourceProvider,omitempty"`
	TargetProvider filter.Provider           `json:"targetProvider"`
	Total          int                       `json:"total"`
	Succeeded      int                       `json:"succeeded"`
	Skipped        int                       `json:"skipped"`
	Removed        int                       `json:"removed"`
	Failed         int                       `json:"failed"`
	Errors         []FirewallRuleSyncFailure `json:"errors,omitempty"`
	TaskID         string                    `json:"taskID,omitempty"`
	Queued         bool                      `json:"queued,omitempty"`
}

type FirewallRuleSyncTask struct {
	TaskID    string `json:"taskID,omitempty"`
	Executing bool   `json:"executing"`
}

type FirewallRuleSyncFailure struct {
	SourceUUID  string                   `json:"sourceUUID"`
	Rule        *filter.FirewallRule     `json:"rule,omitempty"`
	ForwardRule *ForwardRule             `json:"forwardRule,omitempty"`
	DockerRule  *DockerPortGuardEndpoint `json:"dockerRule,omitempty"`
	Error       string                   `json:"error"`
}

type FirewallRuleDelete struct {
	UUIDs []string `json:"uuids" validate:"required,min=1,max=256,dive,required,max=64"`
}

type FirewallRuleDeleteResponse struct {
	Succeeded int                         `json:"succeeded"`
	Failed    int                         `json:"failed"`
	Errors    []FirewallRuleDeleteFailure `json:"errors,omitempty"`
}

type FirewallRuleDeleteFailure struct {
	Index int    `json:"index"`
	UUID  string `json:"uuid"`
	Error string `json:"error"`
}

type FirewallRuleUpdate struct {
	UUID string              `json:"uuid" validate:"required,max=64"`
	Rule filter.FirewallRule `json:"rule" validate:"required"`
}

type FirewallRuleReorder struct {
	UUID           string `json:"uuid" validate:"required,max=64"`
	TargetPosition *int64 `json:"targetPosition"`
	Priority       *int   `json:"priority"`
}
