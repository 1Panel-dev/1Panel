package model

type DockerPortGuardPolicy struct {
	BaseModel

	UUID         string `gorm:"uniqueIndex" json:"uuid"`
	ReadOnly     bool   `gorm:"default:false;uniqueIndex:idx_docker_port_guard_endpoint" json:"-"`
	Family       string `gorm:"uniqueIndex:idx_docker_port_guard_endpoint" json:"family"`
	HostIP       string `gorm:"uniqueIndex:idx_docker_port_guard_endpoint" json:"hostIP"`
	HostPort     uint16 `gorm:"uniqueIndex:idx_docker_port_guard_endpoint" json:"hostPort"`
	Protocol     string `gorm:"uniqueIndex:idx_docker_port_guard_endpoint" json:"protocol"`
	Mode         string `json:"mode"`
	Sources      string `gorm:"type:text" json:"-"`
	Description  string `gorm:"type:text" json:"description"`
	NativeAction string `gorm:"default:''" json:"-"`
	NativeRules  string `gorm:"type:text" json:"-"`
	Sequence     int64  `gorm:"default:0" json:"-"`
}

type ForwardingRule struct {
	BaseModel

	Family     string `gorm:"uniqueIndex:idx_forwarding_rule_identity" json:"family"`
	Protocol   string `gorm:"uniqueIndex:idx_forwarding_rule_identity" json:"protocol"`
	Port       string `gorm:"uniqueIndex:idx_forwarding_rule_identity" json:"port"`
	TargetIP   string `gorm:"uniqueIndex:idx_forwarding_rule_identity" json:"targetIP"`
	TargetPort string `gorm:"uniqueIndex:idx_forwarding_rule_identity" json:"targetPort"`
	Interface  string `gorm:"default:'';uniqueIndex:idx_forwarding_rule_identity" json:"interface"`
}

type FirewallRule struct {
	UUID   string `gorm:"primaryKey" json:"uuid"`
	Family string `json:"family"`

	Protocol           string `json:"protocol"`
	SourceAddress      string `json:"sourceAddress"`
	SourcePort         string `json:"sourcePort"`
	DestinationAddress string `json:"destinationAddress"`
	DestinationPort    string `json:"destinationPort"`
	Interface          string `json:"interface"`
	ConnectionStates   string `gorm:"type:text" json:"connectionStates"`
	Action             string `json:"action"`
	Description        string `gorm:"type:text" json:"description"`
	CompatibilityError string `gorm:"type:text" json:"compatibilityError,omitempty"`
	Priority           *int   `json:"priority,omitempty"`
	Sequence           *int64 `gorm:"index" json:"sequence,omitempty"`

	Origin   string `json:"origin"`
	Owner    string `json:"owner"`
	Revision uint   `gorm:"default:1" json:"revision"`
}
