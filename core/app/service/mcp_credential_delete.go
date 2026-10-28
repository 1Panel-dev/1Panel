package service

import (
	"errors"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *MCPService) Delete(owner auth.APIKeyOwner, req dto.MCPCredentialMutation) (resultErr error) {
	defer func() {
		if resultErr != nil {
			if err := s.auditEnrollment(s.DB, owner, "delete", req, resultErr); err != nil {
				resultErr = err
			}
		}
	}()
	if err := s.authorize(owner, "mcp_access_manage"); err != nil {
		return err
	}
	if _, err := uuid.Parse(req.ID); err != nil || req.Revision == 0 {
		return contract.Invalid()
	}
	mcpMutationMu.Lock()
	defer mcpMutationMu.Unlock()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.authorize(owner, "mcp_access_manage"); err != nil {
			return err
		}
		data := repo.MCPRepo{DB: tx}
		key, err := data.Credential(owner.Type, owner.ID, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.Forbidden()
			}
			return err
		}
		if key.Revision != req.Revision {
			return contract.Fail("MCP_CONFLICT", 409, "Credential changed; refresh and retry")
		}
		if err := data.DeleteGrants(key.ID); err != nil {
			return err
		}
		deleted, err := data.DeleteCredential(owner.Type, owner.ID, key.ID, req.Revision)
		if err != nil {
			return err
		}
		if deleted != 1 {
			return contract.Fail("MCP_CONFLICT", 409, "Credential changed; refresh and retry")
		}
		return s.auditEnrollment(tx, owner, "delete", req, nil)
	})
}
