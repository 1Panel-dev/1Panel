package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const mcpServiceCallsPerMinute = 240
const mcpAdmissionLifetime = 350 * time.Second
const mcpAdmissionMaxOperations = 64

type mcpAdmittedCall struct {
	credentialID, tokenHash, toolID, nodeID, policyHash, requestID string
	schemaVersion                                                  int
	expires                                                        time.Time
	operations                                                     int
	active, closed                                                 bool
	timer                                                          *time.Timer
	// Used only when the adapter fails after successful API sub-operations.
	outcome model.MCPInvocation
}

var mcpAdmissions = struct {
	sync.Mutex
	calls map[string]*mcpAdmittedCall
}{calls: map[string]*mcpAdmittedCall{}}
var mcpGlobalCalls = struct {
	sync.Mutex
	start int64
	calls int
}{}

func mcpRateError(message string, retry int) *contract.Error {
	err := contract.Fail("MCP_RATE_LIMITED", 429, message)
	if retry < 1 {
		retry = 1
	}
	err.RetryAfterSeconds = retry
	return err
}

func (s *MCPService) Admit(c *gin.Context, token string, req contract.Admission) (*contract.Admitted, error) {
	now := s.Now()
	nodeID, err := contract.NormalizeNodeID(req.NodeID)
	if err != nil || len(req.RequestID) > 128 || len(req.Arguments) > 64<<10 {
		return nil, contract.Invalid()
	}
	key, owner, err := s.Authenticate(c, token)
	if err != nil {
		return nil, err
	}
	if !admitMCPCall("admit:"+key.ID, now) {
		return nil, mcpRateError("Too many MCP admission attempts", 60)
	}
	_, err = contract.VerifyAccessProof(c.GetHeader(contract.AccessProofHeader), token, time.Now())
	if err != nil {
		return nil, err
	}
	catalog, err := s.catalog()
	if err != nil {
		return nil, err
	}
	tool, found := catalogIndex(catalog)[req.ToolID]
	if !found {
		return nil, contract.Fail("MCP_UNKNOWN_TOOL", 400, "Unknown managed tool")
	}
	var args map[string]any
	if json.Unmarshal(req.Arguments, &args) != nil || args == nil || !contract.ValidateInput(tool.InputSchema, args) {
		return nil, contract.Invalid()
	}
	requestedNode := "local"
	if value, exists := args["nodeId"]; exists {
		text, ok := value.(string)
		if !ok {
			return nil, contract.Invalid()
		}
		requestedNode, err = contract.NormalizeNodeID(text)
		if err != nil {
			return nil, err
		}
	}
	if requestedNode != nodeID {
		return nil, contract.Forbidden()
	}
	check := contract.Invoke{ToolID: req.ToolID, NodeID: nodeID, SchemaVersion: req.SchemaVersion, PolicyHash: req.PolicyHash, OperationID: tool.Operations[0].ID}
	if _, err := s.authorizeInvocation(key, owner, check); err != nil {
		record := model.MCPInvocation{ID: newMCPAdmissionID(), CredentialID: key.ID, CredentialName: key.Name, OwnerType: key.OwnerType, OwnerID: key.OwnerID, OwnerName: key.OwnerName, ToolID: safeMCPAuditID(req.ToolID), NodeID: nodeID, RequestID: safeMCPAuditID(req.RequestID), PolicyHash: safeMCPPolicyHash(req.PolicyHash), Revision: key.Revision, Status: "denied", Code: "MCP_FORBIDDEN", StartedAt: now, FinishedAt: &now}
		if known, ok := err.(*contract.Error); ok {
			record.Code = known.Code
		}
		if auditErr := s.DB.Create(&record).Error; auditErr != nil {
			return nil, contract.Fail("MCP_EXECUTION_FAILED", 503, "Admission audit is unavailable")
		}
		if known, ok := err.(*contract.Error); ok {
			copy := *known
			copy.InvocationID = record.ID
			return nil, &copy
		}
		return nil, err
	}
	if !mcpIsNodeDiscovery(tool) {
		if err := s.authorizeToolOnNode(owner, nodeID, tool); err != nil {
			return nil, err
		}
	}
	select {
	case mcpExecutionSlots <- struct{}{}:
	default:
		return nil, mcpRateError("Too many concurrent MCP calls", 1)
	}
	release := true
	defer func() {
		if release {
			<-mcpExecutionSlots
		}
	}()
	mcpMutationMu.Lock()
	defer mcpMutationMu.Unlock()
	mcpGlobalCalls.Lock()
	defer mcpGlobalCalls.Unlock()
	now = s.Now()
	if now.Unix() >= mcpGlobalCalls.start+60 {
		mcpGlobalCalls.start = now.Unix()
		mcpGlobalCalls.calls = 0
	}
	if mcpGlobalCalls.calls >= mcpServiceCallsPerMinute {
		return nil, mcpRateError("MCP global call limit exceeded", int(mcpGlobalCalls.start+60-now.Unix()))
	}
	id, err := randomMCPAdmissionID()
	if err != nil {
		return nil, contract.Fail("MCP_EXECUTION_FAILED", 503, "Admission could not be created")
	}
	if err := s.validateMCPAdmission(key, now); err != nil {
		return nil, err
	}
	entry := &mcpAdmittedCall{credentialID: key.ID, tokenHash: tokenDigest(token), toolID: req.ToolID, nodeID: nodeID, policyHash: req.PolicyHash, requestID: req.RequestID, schemaVersion: req.SchemaVersion, expires: time.Now().Add(mcpAdmissionLifetime)}
	entry.outcome = model.MCPInvocation{ID: newMCPAdmissionID(), CredentialID: key.ID, CredentialName: key.Name, OwnerType: key.OwnerType, OwnerID: key.OwnerID, OwnerName: key.OwnerName, ToolID: safeMCPAuditID(req.ToolID), NodeID: nodeID, RequestID: safeMCPAuditID(req.RequestID), OperationID: "tool.result", PolicyHash: safeMCPPolicyHash(req.PolicyHash), Revision: key.Revision, StartedAt: now}
	mcpAdmissions.Lock()
	mcpAdmissions.calls[id] = entry
	entry.timer = time.AfterFunc(mcpAdmissionLifetime, func() { closeMCPAdmission(id) })
	mcpAdmissions.Unlock()
	mcpGlobalCalls.calls++
	release = false
	return &contract.Admitted{AdmissionID: id, ExpiresAt: entry.expires}, nil
}

func randomMCPAdmissionID() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}
func newMCPAdmissionID() string { return uuid.NewString() }

func (s *MCPService) validateMCPAdmission(key model.MCPCredential, now time.Time) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var latest model.MCPCredential
		if err := tx.Where("id = ? AND revision = ? AND status = ? AND revoked_at IS NULL", key.ID, key.Revision, "Enable").Take(&latest).Error; err != nil {
			return contract.Forbidden()
		}
		if latest.ExpiresAt != nil && !now.Before(*latest.ExpiresAt) {
			return contract.Forbidden()
		}
		var setting model.MCPSetting
		if err := tx.Select("enabled").Take(&setting, 1).Error; err != nil || !setting.Enabled {
			return contract.Fail("MCP_DISABLED", 503, "MCP access is disabled")
		}

		return nil
	})
}

func closeMCPAdmission(id string) {
	mcpAdmissions.Lock()
	defer mcpAdmissions.Unlock()
	entry := mcpAdmissions.calls[id]
	if entry == nil {
		return
	}
	closeMCPAdmissionLocked(id, entry)
}

func closeMCPAdmissionLocked(id string, entry *mcpAdmittedCall) {
	entry.closed = true
	if entry.active {
		return
	}
	if entry.timer != nil {
		entry.timer.Stop()
	}
	delete(mcpAdmissions.calls, id)
	<-mcpExecutionSlots
}

func (s *MCPService) Complete(token string, req contract.Completion) error {
	_, err := s.CompleteWithOutcome(token, req)
	return err
}

func (s *MCPService) CompleteWithOutcome(token string, req contract.Completion) (*contract.Completed, error) {
	if len(req.AdmissionID) != 64 || len(req.ErrorCode) > 64 {
		return nil, contract.Invalid()
	}
	mcpAdmissions.Lock()
	entry := mcpAdmissions.calls[req.AdmissionID]
	if entry != nil && entry.tokenHash != tokenDigest(token) {
		mcpAdmissions.Unlock()
		return nil, contract.Forbidden()
	}
	var outcome *model.MCPInvocation
	if entry != nil && req.ErrorCode != "" {
		if entry.active {
			closeMCPAdmissionLocked(req.AdmissionID, entry)
			mcpAdmissions.Unlock()
			return nil, contract.Fail("MCP_CONFLICT", 409, "An API operation is still active")
		}
		record := entry.outcome
		outcome = &record
	}
	if entry != nil {
		closeMCPAdmissionLocked(req.AdmissionID, entry)
	}
	mcpAdmissions.Unlock()
	result := &contract.Completed{}
	if outcome != nil {
		now := s.Now()
		outcome.Status, outcome.Code = "failed", mcpCompletionErrorCode(req.ErrorCode)
		outcome.FinishedAt, outcome.DurationMs = &now, now.Sub(outcome.StartedAt).Milliseconds()
		if err := s.DB.Create(outcome).Error; err != nil {
			return nil, contract.Fail("MCP_AUDIT_UNAVAILABLE", 503, "Invocation outcome could not be recorded; do not retry a write automatically")
		}
		result.InvocationID = outcome.ID
	}
	return result, nil
}

func mcpCompletionErrorCode(code string) string {
	switch code {
	case "MCP_CONFLICT", "MCP_INVALID_ARGUMENTS", "MCP_SCHEMA_MISMATCH", "MCP_FORBIDDEN", "MCP_EXECUTION_FAILED", "MCP_NODE_UNAVAILABLE", "MCP_DISABLED", "MCP_POLICY_CHANGED", "MCP_UNKNOWN_TOOL", "MCP_RATE_LIMITED", "MCP_CATALOG_UNAVAILABLE", "MCP_UNAUTHENTICATED":
		return code
	default:
		return "MCP_EXECUTION_FAILED"
	}
}

func useMCPAdmission(token string, req contract.Invoke) (func(), error) {
	mcpAdmissions.Lock()
	entry := mcpAdmissions.calls[req.AdmissionID]
	if entry == nil || entry.closed || !time.Now().Before(entry.expires) || entry.active || entry.operations >= mcpAdmissionMaxOperations || entry.tokenHash != tokenDigest(token) || entry.toolID != req.ToolID || entry.nodeID != req.NodeID || entry.policyHash != req.PolicyHash || entry.schemaVersion != req.SchemaVersion || entry.requestID != req.RequestID {
		mcpAdmissions.Unlock()
		return nil, contract.Forbidden()
	}
	entry.active = true
	entry.operations++
	mcpAdmissions.Unlock()
	return func() {
		mcpAdmissions.Lock()
		entry.active = false
		if entry.closed {
			closeMCPAdmissionLocked(req.AdmissionID, entry)
		}
		mcpAdmissions.Unlock()
	}, nil
}
