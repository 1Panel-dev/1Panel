package v2

import "github.com/1Panel-dev/1Panel/core/i18n"

// Management responses use panel localization while runtime errors retain their
// stable machine contract. Never render raw errors or credential data here.
func mcpManagementMessage(code string) string {
	key := "ErrInternalServer"
	switch code {
	case "MCP_INVALID_ARGUMENTS":
		key = "ErrInvalidParams"
	case "MCP_UNAUTHENTICATED":
		key = "ErrNotLogin"
	case "MCP_FORBIDDEN":
		key = "ErrRBAC"
	case "MCP_CONFLICT":
		key = "ErrMCPConflict"
	case "MCP_POLICY_CHANGED", "MCP_SCHEMA_MISMATCH":
		key = "ErrMCPPolicyChanged"
	case "MCP_CATALOG_UNAVAILABLE":
		key = "ErrMCPCatalogUnavailable"
	case "MCP_DISABLED":
		key = "ErrMCPDisabled"
	case "MCP_SECRET_KEY_UNAVAILABLE":
		key = "ErrMCPSecretKeyUnavailable"
	case "MCP_SECRET_INVALID":
		key = "ErrMCPSecretInvalid"
	case "MCP_CREDENTIAL_INACTIVE":
		key = "ErrMCPCredentialInactive"
	case "MCP_AUDIT_UNAVAILABLE":
		key = "ErrMCPAuditUnavailable"
	case "MCP_ENDPOINT_UNAVAILABLE":
		key = "ErrMCPEndpointUnavailable"
	}
	return i18n.GetMsgWithMap(key, nil)
}
