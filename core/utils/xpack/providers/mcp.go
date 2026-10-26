package providers

import (
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/buserr"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/gin-gonic/gin"
)

type MCPManagementAuthorizer interface {
	AuthorizeMCPManagement(owner auth.APIKeyOwner, nodeName string, requiredCodes []string) error
}

type MCPAPIAuthorizer interface {
	AuthorizeMCPAPI(owner auth.APIKeyOwner, nodeID, method, requestPath string) error
	BindMCPAPIIdentity(c *gin.Context, owner auth.APIKeyOwner) error
}

type MCPNode struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	IsLocal bool   `json:"isLocal"`
}

type MCPNodeProvider interface {
	ListMCPNodes(owner auth.APIKeyOwner) ([]MCPNode, error)
	ResolveMCPNode(owner auth.APIKeyOwner, id string) (MCPNode, error)
}

func ValidateMCPNodeID(id string) error {
	if id == "local" {
		return nil
	}
	value, err := strconv.ParseUint(id, 10, 64)
	if err != nil || value == 0 || strconv.FormatUint(value, 10) != id {
		return buserr.New("ErrRBAC")
	}
	return nil
}

func ValidateMCPAPIBoundary(nodeID, method, requestPath string) error {
	if ValidateMCPNodeID(nodeID) != nil || !strings.HasPrefix(requestPath, "/api/v2/") || path.Clean(requestPath) != requestPath || strings.ContainsAny(requestPath, "%?#\\\x00\r\n") {
		return buserr.New("ErrRBAC")
	}
	if nodeID != "local" && strings.HasPrefix(requestPath, "/api/v2/core/") {
		return buserr.New("ErrRBAC")
	}
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return buserr.New("ErrRBAC")
	}
	for _, prefix := range []string{"/api/v2/internal", "/api/v2/core/mcp", "/api/v2/core/auth", "/api/v2/hosts/terminal", "/api/v2/xpack/alert/offline", "/api/v2/hosts/firewall/port"} {
		if requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") {
			if requestPath == contract.NodesPath && nodeID == "local" && method == http.MethodGet {
				return nil
			}
			return buserr.New("ErrRBAC")
		}
	}
	return nil
}

func BindMCPResolvedOwner(c *gin.Context, owner auth.APIKeyOwner) {
	c.Set("API_AUTH", true)
	c.Set("MCP_API_AUTH", true)
	c.Set("API_AUTH_USERNAME", owner.Name)
	c.Set("API_AUTH_USER_ID", owner.ID)
	c.Set("API_AUTH_IS_SUPER_ADMIN", owner.IsSuperAdmin)
}

func ValidateMCPManagementBoundary(nodeName string, requiredCodes []string) error {
	if nodeName != "local" || len(requiredCodes) == 0 {
		return buserr.New("ErrRBAC")
	}
	for _, code := range requiredCodes {
		switch code {
		case "mcp_access_view", "mcp_access_manage":
		default:
			return buserr.New("ErrRBAC")
		}
	}
	return nil
}
