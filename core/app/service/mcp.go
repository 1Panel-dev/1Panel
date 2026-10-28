package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/1Panel-dev/1Panel/core/utils/xpack/providers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var mcpMutationMu sync.Mutex
var mcpExecutionSlots = make(chan struct{}, 8)
var mcpCallWindow = struct {
	sync.Mutex
	entries map[string]mcpWindow
}{entries: make(map[string]mcpWindow)}

type mcpWindow struct {
	start time.Time
	calls int
}

func admitMCPCall(id string, now time.Time) bool {
	mcpCallWindow.Lock()
	defer mcpCallWindow.Unlock()
	window, exists := mcpCallWindow.entries[id]
	if !exists || now.Sub(window.start) >= time.Minute {
		if len(mcpCallWindow.entries) >= 4096 {
			for key, value := range mcpCallWindow.entries {
				if now.Sub(value.start) >= time.Minute {
					delete(mcpCallWindow.entries, key)
				}
			}
			if len(mcpCallWindow.entries) >= 4096 && !exists {
				return false
			}
		}
		window = mcpWindow{start: now}
	}
	if window.calls >= 120 {
		return false
	}
	window.calls++
	mcpCallWindow.entries[id] = window
	return true
}

type MCPManagementAuthorizer interface {
	AuthorizeMCPManagement(auth.APIKeyOwner, string, []string) error
}
type MCPAPIAuthorizer interface {
	AuthorizeMCPAPI(auth.APIKeyOwner, string, string, string) error
}
type MCPExecutor func(context.Context, contract.APIRequest) (any, error)
type MCPService struct {
	DB              *gorm.DB
	Provider        auth.APIKeyOwnerProvider
	Authorizer      MCPManagementAuthorizer
	APIAuthorizer   MCPAPIAuthorizer
	NodeProvider    providers.MCPNodeProvider
	Execute         MCPExecutor
	Enabled         func() (bool, error)
	Now             func() time.Time
	LoadCatalog     func() ([]contract.Tool, error)
	SecretKey       func() (string, error)
	EnrollmentAudit func(*gorm.DB, *model.OperationLog) error
}

var NewMCPService = func() *MCPService {
	owner, _ := xpack.AuthProvider.(auth.APIKeyOwnerProvider)
	authorizer, _ := xpack.AuthProvider.(MCPManagementAuthorizer)
	apiAuthorizer, _ := xpack.AuthProvider.(MCPAPIAuthorizer)
	nodeProvider, _ := xpack.AuthProvider.(providers.MCPNodeProvider)
	return &MCPService{DB: global.DB, Provider: owner, Authorizer: authorizer, APIAuthorizer: apiAuthorizer, NodeProvider: nodeProvider, Execute: executeMCPAPI, Now: time.Now, SecretKey: encrypt.APIKeyEncryptionKey, LoadCatalog: func() ([]contract.Tool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return NewMCPRuntimeService().Catalog(ctx)
	}, Enabled: func() (bool, error) {
		enabled, err := (repo.MCPRepo{DB: global.DB}).Enabled()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return enabled, err
	}}
}

func (s *MCPService) authorize(owner auth.APIKeyOwner, codes ...string) error {
	if s.Authorizer == nil || s.Authorizer.AuthorizeMCPManagement(owner, "local", codes) != nil {
		return contract.Forbidden()
	}
	return nil
}

func (s *MCPService) authorizeAPI(owner auth.APIKeyOwner, nodeID, method, path string) error {
	if s.APIAuthorizer == nil || s.APIAuthorizer.AuthorizeMCPAPI(owner, nodeID, method, path) != nil {
		return contract.Forbidden()
	}
	return nil
}
func (s *MCPService) authorizeToolOnNode(owner auth.APIKeyOwner, nodeID string, tool contract.Tool) error {
	for _, op := range tool.Operations {
		if err := s.authorizeAPI(owner, nodeID, op.Method, contract.PermissionPath(op)); err != nil {
			return err
		}
	}
	return nil
}
func (s *MCPService) ManagementOwner(c *gin.Context, write bool) (auth.APIKeyOwner, error) {
	if c.GetHeader("Authorization") != "" {
		return auth.APIKeyOwner{}, contract.Forbidden()
	}
	if _, err := auth.RequireAPIKeySession(c); err != nil {
		return auth.APIKeyOwner{}, contract.Fail("MCP_UNAUTHENTICATED", 401, "A panel session is required")
	}
	if s.Provider == nil {
		return auth.APIKeyOwner{}, contract.Forbidden()
	}
	owner, err := s.Provider.CurrentAPIKeyOwner(c)
	if err != nil {
		return auth.APIKeyOwner{}, contract.Forbidden()
	}
	code := "mcp_access_view"
	if write {
		code = "mcp_access_manage"
	}
	if err = s.authorize(owner, code); err != nil {
		return auth.APIKeyOwner{}, err
	}
	return owner, nil
}
func validMCPPage(page, size int) bool { return page > 0 && page <= 1000000 && size > 0 && size <= 100 }
func (s *MCPService) catalog() ([]contract.Tool, error) {
	unavailable := func() error {
		return contract.Fail("MCP_CATALOG_UNAVAILABLE", 503, "The installed MCP tool catalog is unavailable or incompatible")
	}
	if s.LoadCatalog == nil {
		return nil, unavailable()
	}
	tools, err := s.LoadCatalog()
	if err != nil {
		return nil, unavailable()
	}
	tools, err = contract.ValidateCatalog(tools)
	if err != nil {
		return nil, unavailable()
	}
	return tools, nil
}
func catalogIndex(tools []contract.Tool) map[string]contract.Tool {
	index := make(map[string]contract.Tool, len(tools))
	for _, tool := range tools {
		index[tool.ToolID] = tool
	}
	return index
}
func mcpGrantAllows(grant model.MCPGrant, tool contract.Tool) bool {
	return grant.ToolID == tool.ToolID && grant.PolicyHash != "" &&
		contract.ToolAuthorizationAllows(grant.AuthorizationSnapshot, grant.PolicyHash, tool)
}
func mcpPolicyChanged() error {
	return contract.Fail("MCP_POLICY_CHANGED", 409, "Tool API policy changed; reload the catalog and explicitly save the credential authorization again")
}
func (s *MCPService) Catalog(owner auth.APIKeyOwner) (map[string]any, error) {
	catalog, err := s.catalog()
	if err != nil {
		return nil, err
	}
	nodes, err := s.Nodes(owner)
	if err != nil {
		return nil, err
	}
	tools := []contract.Tool{}
	toolNodeIDs := make(map[string][]string)
	nodeIDs := mcpNodeItemIDs(nodes)
	for _, tool := range catalog {
		allowed := []string{}
		for _, nodeID := range nodeIDs {
			if mcpIsNodeDiscovery(tool) || s.authorizeToolOnNode(owner, nodeID, tool) == nil {
				allowed = append(allowed, nodeID)
			}
		}
		if len(allowed) > 0 || mcpIsNodeDiscovery(tool) {
			tools = append(tools, tool)
			toolNodeIDs[tool.ToolID] = allowed
		}
	}
	return map[string]any{"contractVersion": contract.ContractVersion, "scope": "nodes", "tools": tools, "toolNodeIds": toolNodeIDs}, nil
}
func (s *MCPService) validateFields(owner auth.APIKeyOwner, fields *dto.MCPCredentialFields, checkExpiry bool) ([]model.MCPGrant, error) {
	fields.Name = strings.TrimSpace(fields.Name)
	fields.Description = strings.TrimSpace(fields.Description)
	if fields.Name == "" || utf8.RuneCountInString(fields.Name) > 64 || utf8.RuneCountInString(fields.Description) > 256 || strings.ContainsAny(fields.Name, "\x00\r\n") || len(fields.Tools) > 256 {
		return nil, contract.Invalid()
	}
	if checkExpiry && fields.ExpiresAt != nil && !fields.ExpiresAt.After(s.Now()) {
		return nil, contract.Invalid()
	}
	if err := s.validateNodeIDs(owner, &fields.NodeIDs); err != nil {
		return nil, err
	}
	catalog, err := s.catalog()
	if err != nil {
		return nil, err
	}
	index := catalogIndex(catalog)
	slices.Sort(fields.Tools)
	fields.Tools = slices.Compact(fields.Tools)
	if len(fields.ToolPolicies) != len(fields.Tools) {
		return nil, mcpPolicyChanged()
	}
	if len(fields.ToolReadOnly) != len(fields.Tools) {
		return nil, mcpPolicyChanged()
	}
	grants := make([]model.MCPGrant, 0, len(fields.Tools))
	for _, id := range fields.Tools {
		tool, ok := index[id]
		if !ok || fields.ToolPolicies[id] == "" || fields.ToolPolicies[id] != tool.PolicyHash {
			return nil, mcpPolicyChanged()
		}
		readOnly, reviewed := fields.ToolReadOnly[id]
		if !reviewed || readOnly != tool.ReadOnly {
			return nil, mcpPolicyChanged()
		}
		if err := s.authorizeToolOnNodes(owner, fields.NodeIDs, tool); err != nil {
			return nil, err
		}
		snapshot, err := contract.ToolAuthorizationSnapshot(tool)
		if err != nil {
			return nil, err
		}
		grants = append(grants, model.MCPGrant{ToolID: id, PolicyHash: tool.PolicyHash, AuthorizationSnapshot: snapshot})
	}
	return grants, nil
}
func (s *MCPService) credentialItem(key model.MCPCredential, catalog map[string]contract.Tool) (dto.MCPCredentialItem, error) {
	nodeIDs, err := mcpCredentialNodeIDs(key)
	if err != nil {
		return dto.MCPCredentialItem{}, err
	}
	grants, err := (repo.MCPRepo{DB: s.DB}).Grants(key.ID)
	if err != nil {
		return dto.MCPCredentialItem{}, err
	}
	tools := []string{}
	policies := map[string]string{}
	readOnly := map[string]bool{}
	stale := []string{}
	for _, grant := range grants {
		tools = append(tools, grant.ToolID)
		policies[grant.ToolID] = grant.PolicyHash
		tool, exists := catalog[grant.ToolID]
		if exists {
			readOnly[grant.ToolID] = tool.ReadOnly
		}
		if exists && mcpGrantAllows(grant, tool) {
			// Bind this compatibility decision to the catalog used to calculate
			// it, while leaving the stored consent baseline unchanged.
			policies[grant.ToolID] = tool.PolicyHash
		} else {
			stale = append(stale, grant.ToolID)
		}
	}
	status := key.Status
	if status != "Revoked" && key.ExpiresAt != nil && !s.Now().Before(*key.ExpiresAt) {
		status = "Expired"
	}
	return dto.MCPCredentialItem{ID: key.ID, MCPCredentialFields: dto.MCPCredentialFields{Name: key.Name, Description: key.Description, ExpiresAt: key.ExpiresAt, Tools: tools, ToolPolicies: policies, ToolReadOnly: readOnly, NodeIDs: nodeIDs}, StaleTools: stale, KeyHint: key.KeyHint, Status: status, Revision: key.Revision, CreatedAt: key.CreatedAt, OwnerName: key.OwnerName}, nil
}
func (s *MCPService) Search(owner auth.APIKeyOwner, req dto.MCPCredentialSearch) (map[string]any, error) {
	if !validMCPPage(req.Page, req.PageSize) {
		return nil, contract.Invalid()
	}
	keys, total, err := (repo.MCPRepo{DB: s.DB}).SearchCredentials(owner.Type, owner.ID, req.Page, req.PageSize, req.ExcludeRevoked)
	if err != nil {
		return nil, err
	}
	items := []dto.MCPCredentialItem{}
	catalog, _ := s.catalog()
	index := catalogIndex(catalog)
	for _, key := range keys {
		item, err := s.credentialItem(key, index)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return map[string]any{"items": items, "total": total}, nil
}
func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func (s *MCPService) Create(owner auth.APIKeyOwner, req dto.MCPCredentialCreate) (result *dto.MCPCredentialCreated, resultErr error) {
	defer func() {
		if resultErr != nil {
			if err := s.auditEnrollment(s.DB, owner, "create", dto.MCPCredentialMutation{}, resultErr); err != nil {
				result, resultErr = nil, err
			}
		}
	}()
	if _, err := uuid.Parse(req.RequestID); err != nil {
		return nil, contract.Invalid()
	}
	grants, err := s.validateFields(owner, &req.MCPCredentialFields, true)
	if err != nil {
		return nil, err
	}
	fields, _ := json.Marshal(req.MCPCredentialFields)
	requestHash := tokenDigest(string(fields))
	var key model.MCPCredential
	var token string
	repeated := false
	mcpMutationMu.Lock()
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		store := repo.MCPRepo{DB: tx}
		var err error
		key, err = store.CredentialByRequest(owner.Type, owner.ID, req.RequestID)
		if err == nil {
			if key.RequestHash != requestHash {
				return contract.Fail("MCP_CONFLICT", 409, "Creation request already used")
			}
			repeated = true
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		count, err := store.CountActiveName(owner.Type, owner.ID, req.Name, "")
		if err != nil {
			return err
		}
		if count > 0 {
			return contract.Fail("MCP_CONFLICT", 409, "Credential name already exists")
		}
		secret := make([]byte, 32)
		if _, err = rand.Read(secret); err != nil {
			return err
		}
		token = "mcp_" + hex.EncodeToString(secret)
		key = model.MCPCredential{ID: uuid.NewString(), OwnerType: owner.Type, OwnerID: owner.ID, OwnerName: owner.Name, Name: req.Name, ActiveName: &req.Name, Description: req.Description, NodeIDs: mcpEncodeNodeIDs(req.NodeIDs), RequestID: req.RequestID, RequestHash: requestHash, TokenHash: tokenDigest(token), KeyHint: token[:8] + "…" + token[len(token)-4:], Status: "Enable", ExpiresAt: req.ExpiresAt, Revision: 1}
		if err = s.sealCredentialToken(&key, token); err != nil {
			return err
		}
		if err = store.CreateCredential(&key); err != nil {
			return err
		}
		if err = store.ReplaceTools(key.ID, grants); err != nil {
			return err
		}
		return s.auditEnrollment(tx, owner, "create", dto.MCPCredentialMutation{ID: key.ID, Revision: key.Revision}, nil)
	})
	mcpMutationMu.Unlock()
	if err != nil {
		return nil, err
	}
	catalog, _ := s.catalog()
	item, err := s.credentialItem(key, catalogIndex(catalog))
	if err != nil {
		return nil, err
	}
	return &dto.MCPCredentialCreated{Item: item, Token: token, AlreadyCreated: repeated}, nil
}
func (s *MCPService) mutate(owner auth.APIKeyOwner, req dto.MCPCredentialMutation, change func(*gorm.DB, *model.MCPCredential) error) error {
	if _, err := uuid.Parse(req.ID); err != nil || req.Revision == 0 {
		return contract.Invalid()
	}
	mcpMutationMu.Lock()
	defer mcpMutationMu.Unlock()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		store := repo.MCPRepo{DB: tx}
		key, err := store.Credential(owner.Type, owner.ID, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.Forbidden()
			}
			return err
		}
		if key.Revision != req.Revision || key.Status == "Revoked" {
			return contract.Fail("MCP_CONFLICT", 409, "Credential changed; refresh and retry")
		}
		if err := change(tx, &key); err != nil {
			return err
		}
		if key.Status == "Revoked" {
			key.SecretCiphertext = ""
			key.SecretVersion = 0
		}
		key.Revision++
		affected, err := store.UpdateCredential(&key, req.Revision)
		if err != nil {
			return err
		}
		if affected != 1 {
			return contract.Fail("MCP_CONFLICT", 409, "Credential changed; refresh and retry")
		}
		return nil
	})
}
func (s *MCPService) Update(owner auth.APIKeyOwner, req dto.MCPCredentialUpdate) error {
	if _, err := uuid.Parse(req.ID); err != nil || req.Revision == 0 {
		return contract.Invalid()
	}
	grants, err := s.validateFields(owner, &req.MCPCredentialFields, true)
	if err != nil {
		return err
	}
	return s.mutate(owner, req.MCPCredentialMutation, func(tx *gorm.DB, key *model.MCPCredential) error {
		count, err := (repo.MCPRepo{DB: tx}).CountActiveName(owner.Type, owner.ID, req.Name, key.ID)
		if err != nil {
			return err
		}
		if count > 0 {
			return contract.Fail("MCP_CONFLICT", 409, "Credential name already exists")
		}
		key.Name = req.Name
		key.ActiveName = &req.Name
		key.Description = req.Description
		key.NodeIDs = mcpEncodeNodeIDs(req.NodeIDs)
		key.ExpiresAt = req.ExpiresAt
		return (repo.MCPRepo{DB: tx}).ReplaceTools(key.ID, grants)
	})
}
func (s *MCPService) Status(owner auth.APIKeyOwner, req dto.MCPCredentialStatus) error {
	if req.Status != "Enable" && req.Status != "Disable" {
		return contract.Invalid()
	}
	return s.mutate(owner, req.MCPCredentialMutation, func(tx *gorm.DB, key *model.MCPCredential) error {
		if req.Status == "Enable" {
			catalog, err := s.catalog()
			if err != nil {
				return err
			}
			index := catalogIndex(catalog)
			grants, err := (repo.MCPRepo{DB: tx}).Grants(key.ID)
			if err != nil {
				return err
			}
			nodeIDs, err := mcpCredentialNodeIDs(*key)
			if err != nil {
				return err
			}
			fields := dto.MCPCredentialFields{Name: key.Name, Description: key.Description, ExpiresAt: key.ExpiresAt, ToolPolicies: map[string]string{}, ToolReadOnly: map[string]bool{}, NodeIDs: nodeIDs}
			for _, grant := range grants {
				tool, exists := index[grant.ToolID]
				if !exists || !mcpGrantAllows(grant, tool) {
					return mcpPolicyChanged()
				}
				fields.Tools = append(fields.Tools, grant.ToolID)
				fields.ToolPolicies[grant.ToolID] = tool.PolicyHash
				fields.ToolReadOnly[grant.ToolID] = tool.ReadOnly
			}
			if _, err = s.validateFields(owner, &fields, true); err != nil {
				return err
			}
		}
		key.Status = req.Status
		return nil
	})
}
func (s *MCPService) Revoke(owner auth.APIKeyOwner, req dto.MCPCredentialMutation) error {
	return s.mutate(owner, req, func(tx *gorm.DB, key *model.MCPCredential) error {
		now := s.Now()
		key.Status = "Revoked"
		key.RevokedAt = &now
		key.ActiveName = nil
		return (repo.MCPRepo{DB: tx}).ReplaceTools(key.ID, nil)
	})
}
func (s *MCPService) Authenticate(c *gin.Context, token string) (model.MCPCredential, auth.APIKeyOwner, error) {
	var key model.MCPCredential
	var owner auth.APIKeyOwner
	if len(token) != 68 || !strings.HasPrefix(token, "mcp_") {
		return key, owner, contract.Fail("MCP_UNAUTHENTICATED", 401, "Invalid MCP credential")
	}
	if _, err := hex.DecodeString(token[4:]); err != nil {
		return key, owner, contract.Fail("MCP_UNAUTHENTICATED", 401, "Invalid MCP credential")
	}
	var err error
	key, err = (repo.MCPRepo{DB: s.DB}).Token(tokenDigest(token))
	if err != nil {
		return key, owner, contract.Fail("MCP_UNAUTHENTICATED", 401, "Invalid MCP credential")
	}
	if key.Status != "Enable" || key.ExpiresAt != nil && !s.Now().Before(*key.ExpiresAt) {
		return key, owner, contract.Fail("MCP_UNAUTHENTICATED", 401, "MCP credential is inactive or expired")
	}
	if s.Provider == nil {
		return key, owner, contract.Forbidden()
	}
	owner, err = s.Provider.ResolveAPIKeyOwner(c, key.OwnerType, key.OwnerID)
	if err != nil {
		return key, owner, contract.Forbidden()
	}
	if s.Enabled == nil {
		return key, owner, contract.Fail("MCP_DISABLED", 503, "MCP access is disabled")
	}
	enabled, err := s.Enabled()
	if err != nil || !enabled {
		return key, owner, contract.Fail("MCP_DISABLED", 503, "MCP access is disabled")
	}
	return key, owner, nil
}
func (s *MCPService) RuntimeTools(c *gin.Context, token string) (map[string]any, error) {
	key, _, err := s.Authenticate(c, token)
	if err != nil {
		return nil, err
	}
	catalog, err := s.catalog()
	if err != nil {
		return nil, err
	}
	grants, err := (repo.MCPRepo{DB: s.DB}).Grants(key.ID)
	if err != nil {
		return nil, err
	}
	policies := make(map[string]model.MCPGrant, len(grants))
	for _, grant := range grants {
		policies[grant.ToolID] = grant
	}
	latest, owner, err := s.Authenticate(c, token)
	if err != nil {
		return nil, err
	}
	if latest.Revision != key.Revision {
		return nil, contract.Forbidden()
	}
	nodes, err := s.credentialNodes(latest, owner)
	if err != nil {
		return nil, err
	}
	tools := []contract.Tool{}
	for _, tool := range catalog {
		if grant, exists := policies[tool.ToolID]; exists && mcpGrantAllows(grant, tool) && s.authorizeToolOnNodes(owner, mcpNodeItemIDs(nodes), tool) == nil {
			tools = append(tools, tool)
		}
	}
	return map[string]any{"contractVersion": contract.ContractVersion, "tools": tools}, nil
}

func (s *MCPService) authorizeInvocation(key model.MCPCredential, owner auth.APIKeyOwner, req contract.Invoke) (contract.Tool, error) {
	catalog, err := s.catalog()
	if err != nil {
		return contract.Tool{}, err
	}
	tool, ok := catalogIndex(catalog)[req.ToolID]
	if !ok {
		return contract.Tool{}, contract.Fail("MCP_UNKNOWN_TOOL", 400, "Unknown tool")
	}
	if req.SchemaVersion != tool.SchemaVersion || req.PolicyHash == "" || req.PolicyHash != tool.PolicyHash {
		return contract.Tool{}, mcpPolicyChanged()
	}
	if !slices.ContainsFunc(tool.Operations, func(op contract.Operation) bool { return op.ID == req.OperationID }) {
		return contract.Tool{}, contract.Forbidden()
	}
	grants, err := (repo.MCPRepo{DB: s.DB}).Grants(key.ID)
	if err != nil {
		return contract.Tool{}, err
	}
	granted := false
	for _, grant := range grants {
		if grant.ToolID != tool.ToolID {
			continue
		}
		if !mcpGrantAllows(grant, tool) {
			return contract.Tool{}, mcpPolicyChanged()
		}
		granted = true
		break
	}
	if !granted {
		return contract.Tool{}, contract.Forbidden()
	}
	if err := s.authorizeInvocationNode(key, owner, req.NodeID, tool); err != nil {
		return contract.Tool{}, err
	}
	return tool, nil
}
func (s *MCPService) Invoke(c *gin.Context, token string, req contract.Invoke) (*contract.InvocationResult, error) {
	started := s.Now()
	nodeID, nodeErr := contract.NormalizeNodeID(req.NodeID)
	if nodeErr == nil {
		req.NodeID = nodeID
	}
	done, err := useMCPAdmission(token, req)
	if err != nil {
		return nil, err
	}
	defer done()
	record := model.MCPInvocation{ID: uuid.NewString(), ToolID: safeMCPAuditID(req.ToolID), RequestID: safeMCPAuditID(req.RequestID), OperationID: safeMCPAuditID(req.OperationID), PolicyHash: safeMCPPolicyHash(req.PolicyHash), NodeID: safeMCPAuditID(req.NodeID), StartedAt: started, Status: "denied"}
	key, owner, callErr := s.Authenticate(c, token)
	if key.ID == "" {
		return nil, callErr
	}
	if key.ID != "" {
		record.CredentialID = key.ID
		record.CredentialName = key.Name
		record.OwnerType = key.OwnerType
		record.OwnerID = key.OwnerID
		record.OwnerName = key.OwnerName
		record.Revision = key.Revision
	}
	record.Status = "running"
	if err := (repo.MCPRepo{DB: s.DB}).CreateInvocation(&record); err != nil {
		return nil, contract.Fail("MCP_EXECUTION_FAILED", 503, "Invocation audit is unavailable")
	}
	record.Status = "denied"
	finish := func(err error, result any) (*contract.InvocationResult, error) {
		now := s.Now()
		record.FinishedAt = &now
		record.DurationMs = now.Sub(started).Milliseconds()
		if err == nil {
			record.Status = "success"
		} else {
			var e *contract.Error
			if !errors.As(err, &e) {
				e = contract.Fail("MCP_EXECUTION_FAILED", 502, "Tool execution failed; inspect the panel invocation record before retrying")
			}
			record.Code = e.Code
			e.InvocationID = record.ID
			err = e
		}
		if auditErr := s.finishInvocationAudit(record); auditErr != nil {
			return nil, &contract.Error{Code: "MCP_EXECUTION_FAILED", Status: 503, Message: "Invocation outcome could not be recorded; do not retry a write automatically", InvocationID: record.ID}
		}
		if err != nil {
			return nil, err
		}
		return &contract.InvocationResult{InvocationID: record.ID, NodeID: req.NodeID, Result: result}, nil
	}
	if callErr != nil {
		return finish(callErr, nil)
	}
	if nodeErr != nil {
		return finish(contract.Invalid(), nil)
	}
	tool, err := s.authorizeInvocation(key, owner, req)
	if err != nil {
		return finish(err, nil)
	}
	operation, apiPath, arguments, err := contract.NormalizeOperation(tool, req)
	if err != nil {
		return finish(err, nil)
	}
	latest, latestOwner, err := s.Authenticate(c, token)
	if err != nil {
		return finish(err, nil)
	}
	if latest.Revision != key.Revision {
		return finish(contract.Forbidden(), nil)
	}
	currentTool, err := s.authorizeInvocation(latest, latestOwner, req)
	if err != nil {
		return finish(err, nil)
	}
	latest, latestOwner, err = s.Authenticate(c, token)
	if err != nil {
		return finish(err, nil)
	}
	if latest.Revision != key.Revision || s.authorizeInvocationNode(latest, latestOwner, req.NodeID, currentTool) != nil {
		return finish(contract.Forbidden(), nil)
	}
	discovery := mcpIsNodeDiscovery(currentTool)
	if !discovery && s.authorizeAPI(latestOwner, req.NodeID, operation.Method, apiPath) != nil {
		return finish(contract.Forbidden(), nil)
	}
	if !discovery && s.Execute == nil {
		return finish(contract.Fail("MCP_NODE_UNAVAILABLE", 503, "Selected node is unavailable"), nil)
	}
	requestContext := c.Request.Context()
	if requestContext.Err() != nil {
		return finish(contract.Fail("MCP_EXECUTION_FAILED", 499, "Request ended before execution"), nil)
	}
	if operation.Method != http.MethodGet {
		requestContext = context.WithoutCancel(requestContext)
	}
	ctx, cancel := context.WithTimeout(requestContext, 330*time.Second)
	defer cancel()
	enabled, enabledErr := s.Enabled()
	if enabledErr != nil || !enabled {
		return finish(contract.Fail("MCP_DISABLED", 503, "MCP access is disabled"), nil)
	}
	admitted, admissionErr := (repo.MCPRepo{DB: s.DB}).InvocationCredential(key.ID, key.Revision)
	if errors.Is(admissionErr, gorm.ErrRecordNotFound) {
		return finish(contract.Forbidden(), nil)
	}
	if admissionErr != nil {
		return finish(contract.Fail("MCP_EXECUTION_FAILED", 503, "Credential admission is unavailable"), nil)
	}
	if admitted.ExpiresAt != nil && !s.Now().Before(*admitted.ExpiresAt) {
		return finish(contract.Forbidden(), nil)
	}
	record.Status = "failed"
	if discovery {
		nodes, err := s.credentialNodes(latest, latestOwner)
		if err != nil {
			return finish(err, nil)
		}
		return finish(nil, map[string]any{"items": nodes})
	}
	result, err := s.Execute(ctx, contract.APIRequest{InvocationID: record.ID, NodeID: req.NodeID, Owner: latestOwner, Method: operation.Method, Path: apiPath, Arguments: arguments, ResponseFields: operation.ResponseFields, ReadOnly: operation.Method == http.MethodGet})
	return finish(err, result)
}
func safeMCPPolicyHash(value string) string {
	data, err := hex.DecodeString(value)
	if err != nil || len(data) != sha256.Size {
		return ""
	}
	return strings.ToLower(value)
}
func safeMCPAuditID(value string) string {
	if len(value) > 128 {
		return "invalid"
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("_.:-", r)) {
			return "invalid"
		}
	}
	return value
}
func (s *MCPService) Invocations(owner auth.APIKeyOwner, req dto.MCPInvocationSearch) (map[string]any, error) {
	if !validMCPPage(req.Page, req.PageSize) {
		return nil, contract.Invalid()
	}
	if req.Status != "" {
		if req.Status != "success" && req.Status != "denied" && req.Status != "failed" && req.Status != "running" {
			return nil, contract.Invalid()
		}
	}
	items, total, err := (repo.MCPRepo{DB: s.DB}).SearchInvocations(owner.Type, owner.ID, req.Page, req.PageSize, req.CredentialID, req.ToolID, req.Status)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "total": total}, nil
}

func MCPBearer(c *gin.Context) (string, error) {
	if auth.HasAPICredentials(c) || c.GetHeader("CurrentNode") != "" || c.GetHeader("X-Panel-User") != "" || c.Request.URL.RawQuery != "" {
		return "", contract.Invalid()
	}
	values := c.Request.Header.Values("Authorization")
	if len(values) != 1 {
		return "", contract.Fail("MCP_UNAUTHENTICATED", http.StatusUnauthorized, "MCP Bearer credential required")
	}
	parts := strings.Fields(values[0])
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", contract.Fail("MCP_UNAUTHENTICATED", 401, "MCP Bearer credential required")
	}
	return parts[1], nil
}
