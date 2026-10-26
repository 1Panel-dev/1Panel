package service

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"

	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/1Panel-dev/1Panel/core/utils/mcpapi"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/1Panel-dev/1Panel/core/utils/xpack/providers"
)

func executeMCPAPI(ctx context.Context, req contract.APIRequest) (any, error) {
	result, err := mcpapi.Dispatch(ctx, mcpapi.Request{InvocationID: req.InvocationID, NodeID: req.NodeID, Owner: req.Owner, Method: req.Method, Path: req.Path, Arguments: req.Arguments})
	if err != nil {
		if message := mcpapi.SafeReadErrorMessage(err); message != "" {
			if mcpapi.IsSystemReadTransportFailure(err) && ctx.Err() == nil {
				if provider, ok := xpack.AuthProvider.(providers.MCPNodeProvider); ok {
					node, nodeErr := provider.ResolveMCPNode(req.Owner, req.NodeID)
					if nodeErr == nil && node.ID == req.NodeID && node.Name != "" && node.IsLocal == (req.NodeID == "local") && node.IsLocal == (node.Name == "local") {
						switch strings.ToLower(node.Status) {
						case "unhealthy", "lost", "failed":
							message = mcpapi.NodeAbnormalMessage
						}
					}
				}
			}
			return nil, contract.Fail("MCP_EXECUTION_FAILED", 502, message)
		}
		if message := mcpapi.SafeBusinessErrorMessage(err); message != "" {
			return nil, contract.Fail("MCP_EXECUTION_FAILED", 502, message)
		}
		return nil, contract.Fail("MCP_EXECUTION_FAILED", 502, "Panel API execution failed; inspect the invocation before retrying")
	}
	projected, err := projectMCPAPIResult(req, result)
	if err != nil {
		return nil, contract.Fail("MCP_EXECUTION_FAILED", 502, "Panel API returned an incompatible result")
	}
	return projected, nil
}

func projectMCPAPIResult(req contract.APIRequest, result any) (any, error) {
	if req.Method == http.MethodGet && req.Path == "/api/v2/settings/website/dir" {
		directory, ok := result.(string)
		if !ok || !path.IsAbs(directory) || len(directory) > 4096 || strings.ContainsAny(directory, "\x00\r\n") {
			return nil, errors.New("unexpected website directory result")
		}
		result = map[string]any{"websiteDir": directory}
	}
	return contract.ProjectResult(result, req.ResponseFields)
}
