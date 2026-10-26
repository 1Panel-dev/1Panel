package service

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
)

func (s *MCPService) Nodes(owner auth.APIKeyOwner) ([]dto.MCPNodeItem, error) {
	if s.NodeProvider == nil {
		return nil, contract.Forbidden()
	}
	nodes, err := s.NodeProvider.ListMCPNodes(owner)
	if err != nil {
		return nil, contract.Forbidden()
	}
	items := make([]dto.MCPNodeItem, 0, len(nodes))
	seen := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if !contract.ValidateNodeID(node.ID) || node.IsLocal != (node.ID == "local") || seen[node.ID] {
			return nil, contract.Forbidden()
		}
		seen[node.ID] = true
		items = append(items, dto.MCPNodeItem{ID: node.ID, Name: node.Name, Status: node.Status, IsLocal: node.IsLocal})
	}
	return items, nil
}

func mcpNodeItemIDs(nodes []dto.MCPNodeItem) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	return ids
}

func mcpEncodeNodeIDs(ids []string) string {
	data, _ := json.Marshal(ids)
	return string(data)
}

func mcpCredentialNodeIDs(key model.MCPCredential) ([]string, error) {
	var ids []string
	if json.Unmarshal([]byte(key.NodeIDs), &ids) != nil || len(ids) == 0 || len(ids) > 256 {
		return nil, contract.Forbidden()
	}
	for _, id := range ids {
		if !contract.ValidateNodeID(id) {
			return nil, contract.Forbidden()
		}
	}
	return ids, nil
}

func (s *MCPService) validateNodeIDs(owner auth.APIKeyOwner, ids *[]string) error {
	if len(*ids) == 0 || len(*ids) > 256 {
		return contract.Invalid()
	}
	for _, id := range *ids {
		if !contract.ValidateNodeID(id) {
			return contract.Invalid()
		}
	}
	slices.Sort(*ids)
	*ids = slices.Compact(*ids)
	nodes, err := s.Nodes(owner)
	if err != nil {
		return err
	}
	available := mcpNodeItemIDs(nodes)
	for _, id := range *ids {
		if !slices.Contains(available, id) {
			return contract.Forbidden()
		}
	}
	return nil
}

func (s *MCPService) credentialNodes(key model.MCPCredential, owner auth.APIKeyOwner) ([]dto.MCPNodeItem, error) {
	ids, err := mcpCredentialNodeIDs(key)
	if err != nil {
		return nil, err
	}
	nodes, err := s.Nodes(owner)
	if err != nil {
		return nil, err
	}
	allowed := make([]dto.MCPNodeItem, 0, len(nodes))
	for _, node := range nodes {
		if slices.Contains(ids, node.ID) {
			allowed = append(allowed, node)
		}
	}
	return allowed, nil
}

func mcpIsNodeDiscovery(tool contract.Tool) bool {
	return tool.ToolID == "list_nodes" && tool.ReadOnly && len(tool.Operations) == 1 &&
		tool.Operations[0].ID == "nodes" && tool.Operations[0].Method == http.MethodGet &&
		tool.Operations[0].Path == contract.NodesPath
}

func (s *MCPService) authorizeToolOnNodes(owner auth.APIKeyOwner, nodeIDs []string, tool contract.Tool) error {
	if mcpIsNodeDiscovery(tool) {
		return nil
	}
	for _, nodeID := range nodeIDs {
		if s.authorizeToolOnNode(owner, nodeID, tool) == nil {
			return nil
		}
	}
	return contract.Forbidden()
}

func (s *MCPService) authorizeInvocationNode(key model.MCPCredential, owner auth.APIKeyOwner, nodeID string, tool contract.Tool) error {
	if mcpIsNodeDiscovery(tool) {
		if nodeID != "local" {
			return contract.Forbidden()
		}
		_, err := s.credentialNodes(key, owner)
		return err
	}
	nodes, err := s.credentialNodes(key, owner)
	if err != nil {
		return err
	}
	if !slices.Contains(mcpNodeItemIDs(nodes), nodeID) {
		return contract.Forbidden()
	}
	return s.authorizeToolOnNode(owner, nodeID, tool)
}
