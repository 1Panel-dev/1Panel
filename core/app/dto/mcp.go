package dto

import "time"

type MCPCredentialFields struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	ExpiresAt    *time.Time        `json:"expiresAt"`
	Tools        []string          `json:"tools"`
	ToolPolicies map[string]string `json:"toolPolicies"`
	ToolReadOnly map[string]bool   `json:"toolReadOnly"`
	NodeIDs      []string          `json:"nodeIds"`
}
type MCPNodeItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	IsLocal bool   `json:"isLocal"`
}
type MCPCredentialCreate struct {
	RequestID string `json:"requestID"`
	MCPCredentialFields
}
type MCPCredentialMutation struct {
	ID       string `json:"id"`
	Revision uint64 `json:"revision"`
}
type MCPCredentialUpdate struct {
	MCPCredentialMutation
	MCPCredentialFields
}
type MCPCredentialStatus struct {
	MCPCredentialMutation
	Status string `json:"status"`
}
type MCPCredentialItem struct {
	ID string `json:"id"`
	MCPCredentialFields
	KeyHint    string    `json:"keyHint"`
	Status     string    `json:"status"`
	Revision   uint64    `json:"revision"`
	CreatedAt  time.Time `json:"createdAt"`
	OwnerName  string    `json:"ownerName"`
	StaleTools []string  `json:"staleTools"`
}
type MCPCredentialCreated struct {
	Item           MCPCredentialItem `json:"item"`
	Token          string            `json:"token"`
	AlreadyCreated bool              `json:"alreadyCreated,omitempty"`
}
type MCPCredentialSearch struct {
	Page           int  `json:"page"`
	PageSize       int  `json:"pageSize"`
	ExcludeRevoked bool `json:"excludeRevoked"`
}
type MCPInvocationSearch struct {
	Page         int    `json:"page"`
	PageSize     int    `json:"pageSize"`
	CredentialID string `json:"credentialId"`
	ToolID       string `json:"toolId"`
	Status       string `json:"status"`
}
