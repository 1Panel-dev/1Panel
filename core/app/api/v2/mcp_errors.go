package v2

import "github.com/1Panel-dev/1Panel/core/i18n"

func mcpConflictMessageKey(reason string) string {
	switch reason {
	case "settings_revision_changed":
		return "ErrMCPSettingsChanged"
	case "credential_revision_changed":
		return "ErrMCPCredentialChanged"
	case "credential_revoked":
		return "ErrMCPCredentialRevoked"
	case "credential_name_exists":
		return "ErrMCPCredentialNameExists"
	case "request_reused":
		return "ErrMCPRequestReused"
	case "installation_running":
		return "ErrMCPInstallationRunning"
	case "service_operation_running":
		return "ErrMCPServiceOperationRunning"
	case "endpoint_managed_by_website":
		return "ErrMCPEndpointManagedByWebsite"
	case "website_configuration_exists":
		return "ErrMCPWebsiteConfigurationExists"
	case "website_already_bound":
		return "ErrMCPWebsiteAlreadyBound"
	case "website_not_bound":
		return "ErrMCPWebsiteNotBound"
	case "website_missing":
		return "ErrMCPWebsiteMissing"
	case "website_operation_pending":
		return "ErrMCPWebsiteOperationPending"
	case "website_domain_conflict":
		return "ErrMCPWebsiteDomainConflict"
	case "website_ownership_changed":
		return "ErrMCPWebsiteOwnershipChanged"
	case "website_domain_changed":
		return "ErrMCPWebsiteDomainChanged"
	case "website_directory_changed":
		return "ErrMCPWebsiteDirectoryChanged"
	case "website_proxy_modified":
		return "ErrMCPWebsiteProxyModified"
	case "website_modified":
		return "ErrMCPWebsiteModified"
	case "website_request_changed":
		return "ErrMCPWebsiteRequestChanged"
	case "website_proxy_not_loaded":
		return "ErrMCPWebsiteProxyNotLoaded"
	case "website_state_unconfirmed":
		return "ErrMCPWebsiteStateUnconfirmed"
	default:
		return ""
	}
}

func mcpManagementMessage(code string, reasons ...string) string {
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
		if len(reasons) > 0 {
			if specific := mcpConflictMessageKey(reasons[0]); specific != "" {
				key = specific
			}
		}
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
	case "MCP_WEBSITE_PENDING":
		key = "ErrMCPWebsitePending"
	case "MCP_WEBSITE_UNAVAILABLE":
		key = "ErrMCPWebsiteUnavailable"
	case "MCP_WEBSITE_SERVICE_REQUIRED":
		key = "ErrMCPWebsiteServiceRequired"
	case "MCP_WEBSITE_HTTP_PORT_REQUIRED":
		key = "ErrMCPWebsiteHTTPPortRequired"
	}
	return i18n.GetMsgWithMap(key, nil)
}
