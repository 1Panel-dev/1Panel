package helper

import (
	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/utils/xpack/providers"
)

var _ providers.MCPNodeProvider = (*authHelper)(nil)

func (a *authHelper) ResolveMCPNode(owner auth.APIKeyOwner, id string) (providers.MCPNode, error) {
	if id != "local" {
		return providers.MCPNode{}, buserr.New("ErrNoSuchNode")
	}
	if _, err := a.ResolveAPIKeyOwner(nil, owner.Type, owner.ID); err != nil {
		return providers.MCPNode{}, err
	}
	return providers.MCPNode{ID: "local", Name: "local", Status: "Healthy", IsLocal: true}, nil
}

func (a *authHelper) ListMCPNodes(owner auth.APIKeyOwner) ([]providers.MCPNode, error) {
	node, err := a.ResolveMCPNode(owner, "local")
	if err != nil {
		return nil, err
	}
	return []providers.MCPNode{node}, nil
}
