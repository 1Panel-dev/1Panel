package mcp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/core/app/auth"
)

const ContractVersion = 5

const ToolPolicyVersion = 4

const DockerSocketDir = "/etc/1panel/mcp-runtime"
const DockerCoreSocket = DockerSocketDir + "/core.sock"
const RuntimePrefix = "/api/v2/core/mcp/runtime"
const NodesPath = RuntimePrefix + "/nodes"

type Error struct {
	Code              string `json:"code"`
	Message           string `json:"message"`
	InvocationID      string `json:"invocationId,omitempty"`
	Status            int    `json:"-"`
	RetryAfterSeconds int    `json:"retryAfterSeconds,omitempty"`
}

func (e *Error) Error() string { return e.Code }
func Fail(code string, status int, message string) *Error {
	return &Error{Code: code, Status: status, Message: message}
}
func Invalid() *Error { return Fail("MCP_INVALID_ARGUMENTS", 400, "Invalid tool arguments") }
func Forbidden() *Error {
	return Fail("MCP_FORBIDDEN", 403, "This credential or its owner cannot perform this operation")
}

type Operation struct {
	ID             string         `json:"id"`
	Method         string         `json:"method"`
	Path           string         `json:"path"`
	BodySchema     map[string]any `json:"bodySchema"`
	ResponseFields []string       `json:"responseFields"`
}
type Tool struct {
	ToolID        string         `json:"toolId"`
	CategoryID    string         `json:"categoryId"`
	CategoryName  string         `json:"categoryName"`
	CategoryOrder int            `json:"categoryOrder"`
	SchemaVersion int            `json:"schemaVersion"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	InputSchema   map[string]any `json:"inputSchema"`
	OutputSchema  map[string]any `json:"outputSchema"`
	ReadOnly      bool           `json:"readOnly"`
	Risk          string         `json:"risk"`
	Operations    []Operation    `json:"operations"`
	PolicyHash    string         `json:"policyHash"`
}
type Invoke struct {
	NodeID        string            `json:"nodeId"`
	ToolID        string            `json:"toolId"`
	SchemaVersion int               `json:"schemaVersion"`
	OperationID   string            `json:"operationId"`
	PolicyHash    string            `json:"policyHash"`
	Arguments     json.RawMessage   `json:"arguments"`
	PathParams    map[string]string `json:"pathParams,omitempty"`
	RequestID     string            `json:"requestId,omitempty"`
	AdmissionID   string            `json:"admissionId,omitempty"`
}
type Admission struct {
	NodeID        string          `json:"nodeId"`
	ToolID        string          `json:"toolId"`
	SchemaVersion int             `json:"schemaVersion"`
	PolicyHash    string          `json:"policyHash"`
	Arguments     json.RawMessage `json:"arguments"`
	RequestID     string          `json:"requestId,omitempty"`
}
type Admitted struct {
	AdmissionID string    `json:"admissionId"`
	ExpiresAt   time.Time `json:"expiresAt"`
}
type Completion struct {
	AdmissionID string `json:"admissionId"`
	ErrorCode   string `json:"errorCode,omitempty"`
}
type Completed struct {
	InvocationID string `json:"invocationId,omitempty"`
}
type APIRequest struct {
	NodeID         string
	InvocationID   string
	Owner          auth.APIKeyOwner
	Method         string
	Path           string
	Arguments      json.RawMessage
	ResponseFields []string
	ReadOnly       bool
}
type InvocationResult struct {
	InvocationID string `json:"invocationId"`
	NodeID       string `json:"nodeId"`
	Result       any    `json:"result"`
}

var toolIdentifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var categoryIdentifier = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
var pathPart = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)
var fieldPart = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,127}$`)

func NormalizeNodeID(id string) (string, error) {
	if id == "local" {
		return "local", nil
	}
	if len(id) == 0 || len(id) > 20 || id[0] == '0' {
		return "", Invalid()
	}
	for _, ch := range id {
		if ch < '0' || ch > '9' {
			return "", Invalid()
		}
	}
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return "", Invalid()
	}
	return id, nil
}

func ValidateNodeID(id string) bool {
	normalized, err := NormalizeNodeID(id)
	return err == nil && normalized == id
}

func ToolPolicyHash(tool Tool) string {
	ops := append([]Operation(nil), tool.Operations...)
	slices.SortFunc(ops, func(a, b Operation) int { return strings.Compare(a.ID, b.ID) })
	for i := range ops {
		ops[i].ResponseFields = append([]string(nil), ops[i].ResponseFields...)
		slices.Sort(ops[i].ResponseFields)
	}
	encoded, _ := json.Marshal(struct {
		Version       int            `json:"version"`
		SchemaVersion int            `json:"schemaVersion"`
		InputSchema   map[string]any `json:"inputSchema"`
		Operations    []Operation    `json:"operations"`
	}{ToolPolicyVersion, tool.SchemaVersion, tool.InputSchema, ops})
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func ValidateAPIPath(path string, template bool) bool {
	if path == NodesPath {
		return true
	}
	if !strings.HasPrefix(path, "/api/v2/") || len(path) > 1024 || strings.HasSuffix(path, "/") {
		return false
	}
	seen := map[string]bool{}
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if template && strings.HasPrefix(part, ":") {
			name := strings.TrimPrefix(part, ":")
			if !fieldPart.MatchString(name) || seen[name] {
				return false
			}
			seen[name] = true
			continue
		}
		if !pathPart.MatchString(part) || part == "." || part == ".." {
			return false
		}
	}
	for _, part := range strings.Split(path, "/") {
		if part == "internal" || part == "core" || part == "terminal" || part == "ws" || part == "websocket" {
			return false
		}
	}
	return true
}
func ValidateCatalog(input []Tool) ([]Tool, error) {
	if len(input) > 256 {
		return nil, errors.New("too many MCP tools")
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var result []Tool
	if err = json.Unmarshal(encoded, &result); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	type categoryMetadata struct {
		name  string
		order int
	}
	categories := map[string]categoryMetadata{}
	for i := range result {
		t := &result[i]
		if !toolIdentifier.MatchString(t.ToolID) || seen[t.ToolID] || t.SchemaVersion < 1 || t.SchemaVersion > 1000000 || len(t.Operations) < 1 || len(t.Operations) > 32 || len(t.Name) > 256 || len(t.Description) > 8192 || t.InputSchema["type"] != "object" {
			return nil, errors.New("invalid MCP tool")
		}
		if !categoryIdentifier.MatchString(t.CategoryID) || t.CategoryOrder < 1 || t.CategoryOrder > 1000000 || len(t.CategoryName) == 0 || len(t.CategoryName) > 128 || !utf8.ValidString(t.CategoryName) || strings.TrimSpace(t.CategoryName) != t.CategoryName || strings.ContainsFunc(t.CategoryName, unicode.IsControl) {
			return nil, errors.New("invalid MCP category")
		}
		metadata := categoryMetadata{t.CategoryName, t.CategoryOrder}
		if previous, exists := categories[t.CategoryID]; exists && previous != metadata {
			return nil, errors.New("inconsistent MCP category")
		}
		categories[t.CategoryID] = metadata
		seen[t.ToolID] = true
		ops := map[string]bool{}
		for j := range t.Operations {
			op := &t.Operations[j]
			if op.Path == NodesPath && (!t.ReadOnly || op.Method != http.MethodGet) {
				return nil, errors.New("invalid MCP node discovery operation")
			}
			if !toolIdentifier.MatchString(op.ID) || ops[op.ID] || !slices.Contains([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}, op.Method) || !ValidateAPIPath(op.Path, true) || op.BodySchema["type"] != "object" {
				return nil, errors.New("invalid MCP API operation")
			}
			ops[op.ID] = true
			if err = validateSchema(op.BodySchema, 0); err != nil {
				return nil, err
			}
			if _, err = projectionTree(op.ResponseFields); err != nil {
				return nil, err
			}
		}
		t.Risk = "write"
		if t.ReadOnly {
			t.Risk = "read"
		}
		if t.Name == "" {
			t.Name = t.ToolID
		}
		t.PolicyHash = ToolPolicyHash(*t)
	}
	return result, nil
}
func OperationPath(op Operation, params map[string]string) (string, error) {
	parts := strings.Split(op.Path, "/")
	used := 0
	for i, part := range parts {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		value, ok := params[part[1:]]
		if !ok || !pathPart.MatchString(value) || value == "." || value == ".." {
			return "", Invalid()
		}
		parts[i] = value
		used++
	}
	if used != len(params) {
		return "", Invalid()
	}
	path := strings.Join(parts, "/")
	if !ValidateAPIPath(path, false) {
		return "", Invalid()
	}
	return path, nil
}

func PermissionPath(op Operation) string {
	params := map[string]string{}
	for _, part := range strings.Split(op.Path, "/") {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = "1"
		}
	}
	p, _ := OperationPath(op, params)
	return p
}
func NormalizeOperation(tool Tool, req Invoke) (Operation, string, json.RawMessage, error) {
	if _, err := NormalizeNodeID(req.NodeID); err != nil {
		return Operation{}, "", nil, err
	}
	if len(req.RequestID) > 128 || strings.ContainsAny(req.RequestID, "\r\n\x00") {
		return Operation{}, "", nil, Invalid()
	}
	for _, op := range tool.Operations {
		if op.ID == req.OperationID {
			path, err := OperationPath(op, req.PathParams)
			if err != nil {
				return Operation{}, "", nil, err
			}
			var args map[string]any
			if err = DecodeObject(req.Arguments, &args); err != nil {
				return Operation{}, "", nil, err
			}
			if err = validateValue(op.BodySchema, args, 0); err != nil {
				return Operation{}, "", nil, Invalid()
			}
			data, err := json.Marshal(args)
			return op, path, data, err
		}
	}
	return Operation{}, "", nil, Forbidden()
}

func DecodeObject(data []byte, value any) error {
	return DecodeObjectLimit(data, value, 64<<10)
}

func DecodeObjectLimit(data []byte, value any, limit int) error {
	if len(data) == 0 || len(data) > limit {
		return Invalid()
	}
	d := json.NewDecoder(bytes.NewReader(data))
	if err := checkJSON(d); err != nil {
		return Invalid()
	}
	if _, err := d.Token(); err != io.EOF {
		return Invalid()
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return Invalid()
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return Invalid()
	}
	return nil
}
func checkJSON(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := key.(string)
			if !ok || seen[s] {
				return errors.New("duplicate JSON key")
			}
			seen[s] = true
			if err = checkJSON(d); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := checkJSON(d); err != nil {
				return err
			}
		}
	default:
		return errors.New("unexpected delimiter")
	}
	_, err = d.Token()
	return err
}
