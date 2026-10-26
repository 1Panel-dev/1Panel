package model

import "time"

type MCPCredential struct {
	ID               string     `gorm:"type:varchar(36);primaryKey" json:"-"`
	OwnerType        string     `gorm:"not null;uniqueIndex:idx_mcp_request;uniqueIndex:idx_mcp_name;index:idx_mcp_owner" json:"-"`
	OwnerID          string     `gorm:"not null;uniqueIndex:idx_mcp_request;uniqueIndex:idx_mcp_name;index:idx_mcp_owner" json:"-"`
	OwnerName        string     `json:"-"`
	Name             string     `gorm:"not null" json:"-"`
	ActiveName       *string    `gorm:"uniqueIndex:idx_mcp_name" json:"-"`
	Description      string     `json:"-"`
	NodeIDs          string     `gorm:"type:text;not null" json:"-"`
	RequestID        string     `gorm:"not null;uniqueIndex:idx_mcp_request" json:"-"`
	RequestHash      string     `json:"-"`
	TokenHash        string     `gorm:"uniqueIndex;not null" json:"-"`
	SecretCiphertext string     `gorm:"type:text" json:"-"`
	SecretVersion    int        `gorm:"not null;default:0" json:"-"`
	KeyHint          string     `json:"-"`
	Status           string     `gorm:"not null" json:"-"`
	ExpiresAt        *time.Time `json:"-"`
	Revision         uint64     `gorm:"not null" json:"-"`
	CreatedAt        time.Time  `json:"-"`
	UpdatedAt        time.Time  `json:"-"`
	RevokedAt        *time.Time `json:"-"`
}
type MCPGrant struct {
	CredentialID          string `gorm:"primaryKey"`
	ToolID                string `gorm:"primaryKey"`
	PolicyHash            string `gorm:"not null"`
	AuthorizationSnapshot string `gorm:"type:text;not null" json:"-"`
}

type MCPInvocation struct {
	ID             string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	RequestID      string     `json:"requestId"`
	CredentialID   string     `gorm:"index" json:"credentialId"`
	CredentialName string     `json:"credentialName"`
	OwnerType      string     `gorm:"index:idx_mcp_invocation_owner" json:"-"`
	OwnerID        string     `gorm:"index:idx_mcp_invocation_owner" json:"-"`
	OwnerName      string     `json:"ownerName"`
	ToolID         string     `gorm:"index" json:"toolId"`
	OperationID    string     `json:"operationId"`
	PolicyHash     string     `json:"policyHash"`
	NodeID         string     `json:"nodeId"`
	Revision       uint64     `json:"revision"`
	Status         string     `gorm:"index" json:"status"`
	Code           string     `json:"code"`
	StartedAt      time.Time  `gorm:"index" json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	DurationMs     int64      `json:"durationMs"`
}

type MCPSetting struct {
	RuntimeType  string `gorm:"not null;default:docker"`
	DockerImage  string `gorm:"not null;default:1panel/1panel-mcp-server:v0.0.1"`
	ID           uint   `gorm:"primaryKey"`
	Enabled      bool   `gorm:"not null;default:false"`
	ListenPort   int    `gorm:"not null;default:18081"`
	ExternalURL  string
	AccessPolicy string `gorm:"not null;default:inherit"`
	AllowedIPs   string `gorm:"type:text;not null;default:'[]'"`
	Revision     uint64 `gorm:"not null;default:1"`
}
