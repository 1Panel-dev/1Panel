package service

import (
	"encoding/json"
	"errors"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func mcpEnrollmentAuditCode(err error) string {
	if err == nil {
		return "MCP_OK"
	}
	var known *contract.Error
	if errors.As(err, &known) {
		switch known.Code {
		case "MCP_FORBIDDEN", "MCP_CONFLICT", "MCP_INVALID_ARGUMENTS", "MCP_SECRET_INVALID", "MCP_SECRET_KEY_UNAVAILABLE", "MCP_CREDENTIAL_INACTIVE", "MCP_ENDPOINT_UNAVAILABLE", "MCP_POLICY_CHANGED", "MCP_CATALOG_UNAVAILABLE", "MCP_AUDIT_UNAVAILABLE":
			return known.Code
		}
	}
	return "MCP_ENROLLMENT_FAILED"
}

func (s *MCPService) auditEnrollment(db *gorm.DB, owner auth.APIKeyOwner, action string, req dto.MCPCredentialMutation, outcome error) error {
	failed := contract.Fail("MCP_AUDIT_UNAVAILABLE", 503, "Credential access could not be audited; no secret was returned")
	if action != "create" && action != "reveal" && action != "rotate" && action != "export" && action != "delete" {
		return failed
	}
	id := "invalid"
	if parsed, err := uuid.Parse(req.ID); err == nil {
		id = parsed.String()
	}
	status := constant.StatusSuccess
	if outcome != nil {
		status = constant.StatusFailed
	}
	metadata := struct {
		OwnerType    string `json:"ownerType"`
		OwnerID      string `json:"ownerId"`
		Action       string `json:"action"`
		CredentialID string `json:"credentialId"`
		Revision     uint64 `json:"revision"`
		Outcome      string `json:"outcome"`
		Code         string `json:"code"`
	}{owner.Type, owner.ID, action, id, req.Revision, status, mcpEnrollmentAuditCode(outcome)}
	detail, _ := json.Marshal(metadata)
	record := &model.OperationLog{Source: "mcp", User: owner.Name, AuthMethod: "session", Node: "local", Path: "/core/mcp/credentials/" + action, Method: "post", Status: status, Message: metadata.Code, DetailEN: string(detail), DetailZH: string(detail)}
	if db == nil {
		return failed
	}
	var err error
	if s.EnrollmentAudit != nil {
		err = s.EnrollmentAudit(db, record)
	} else {
		err = (repo.MCPRepo{DB: db}).CreateEnrollmentAudit(record)
	}
	if err != nil {
		return failed
	}
	return nil
}
