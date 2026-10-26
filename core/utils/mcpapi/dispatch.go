// Package mcpapi sends an already-authorized MCP operation through the existing
// Core HTTP application without exposing an HTTP header-based identity bridge.
package mcpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/utils/mcpapi/metadata"
	"github.com/1Panel-dev/1Panel/core/utils/xpack/providers"
)

const (
	CapabilityContractVersion = 4
	InvocationHeader          = "X-Panel-MCP-Invocation"
	NodeHeader                = "X-Panel-MCP-Node"
	MethodHeader              = "X-Panel-MCP-Method"
	PathHeader                = "X-Panel-MCP-Path"
)
const maxResponseBytes = 8 << 20

type Request struct {
	InvocationID string
	Owner        auth.APIKeyOwner
	NodeID       string
	Method       string
	Path         string
	Arguments    json.RawMessage
}
type identityKey struct{}
type identity struct {
	InvocationID string
	Owner        auth.APIKeyOwner
	NodeID       string
	Method, Path string
}

var application struct {
	sync.RWMutex
	handler http.Handler
}

func SetHandler(handler http.Handler) {
	application.Lock()
	application.handler = handler
	application.Unlock()
}

func Identity(req *http.Request) (string, auth.APIKeyOwner, bool) {
	value, ok := req.Context().Value(identityKey{}).(identity)
	if !ok || value.Method != req.Method || value.Path != req.URL.Path || req.URL.RawQuery != "" || value.InvocationID == "" || value.Owner.ID == "" || providers.ValidateMCPAPIBoundary(value.NodeID, value.Method, value.Path) != nil {
		return "", auth.APIKeyOwner{}, false
	}
	return value.InvocationID, value.Owner, true
}

func IsRequest(req *http.Request) bool { _, _, ok := Identity(req); return ok }

func HasIdentity(req *http.Request) bool {
	_, ok := req.Context().Value(identityKey{}).(identity)
	return ok
}

func NodeID(req *http.Request) (string, bool) {
	if !IsRequest(req) {
		return "", false
	}
	return req.Context().Value(identityKey{}).(identity).NodeID, true
}

func ClearForwardedHeaders(header http.Header) {
	for _, name := range []string{InvocationHeader, NodeHeader, MethodHeader, PathHeader} {
		header.Del(name)
	}
}

func SetForwardedHeaders(header http.Header, req *http.Request) bool {
	ClearForwardedHeaders(header)
	invocation, _, ok := Identity(req)
	if !ok {
		return false
	}
	nodeID, _ := NodeID(req)
	header.Set(InvocationHeader, invocation)
	header.Set(NodeHeader, nodeID)
	header.Set(MethodHeader, req.Method)
	header.Set(PathHeader, req.URL.Path)
	return true
}

func Dispatch(ctx context.Context, input Request) (result any, err error) {
	if err := providers.ValidateMCPAPIBoundary(input.NodeID, input.Method, input.Path); err != nil {
		return nil, err
	}
	if input.InvocationID == "" || input.Owner.ID == "" || len(input.Arguments) > 64<<10 {
		return nil, errors.New("invalid MCP API request")
	}
	application.RLock()
	handler := application.handler
	application.RUnlock()
	if handler == nil {
		return nil, errors.New("MCP API dispatcher unavailable")
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	request, err := http.NewRequestWithContext(context.WithValue(ctx, identityKey{}, identity{InvocationID: input.InvocationID, Owner: input.Owner, NodeID: input.NodeID, Method: input.Method, Path: input.Path}), input.Method, "http://mcp.internal"+input.Path, bytes.NewReader(input.Arguments))
	if err != nil {
		return nil, errors.New("invalid MCP API request")
	}
	request = metadata.BindRequest(request, IsRequest)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if input.NodeID == "local" {
		request.Header.Set("CurrentNode", "local")
	}
	response := &boundedResponse{header: make(http.Header)}
	defer func() {
		if recover() != nil {
			result, err = nil, errors.New("MCP API execution failed")
		}
	}()
	handler.ServeHTTP(response, request)
	if response.overflow {
		return nil, errors.New("MCP API execution failed")
	}
	if !strings.Contains(strings.ToLower(response.header.Get("Content-Type")), "application/json") {
		if json.Valid(response.body.Bytes()) {
			return nil, errors.New("MCP API execution failed")
		}
		if failure := classifyReadResponseFailure(input, response.status, 0, "", false); failure != nil {
			return nil, failure
		}
		return nil, errors.New("MCP API execution failed")
	}
	var envelope struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	decoder := json.NewDecoder(bytes.NewReader(response.body.Bytes()))
	if decoder.Decode(&envelope) != nil || decoder.Decode(new(any)) != io.EOF {
		return nil, errors.New("MCP API execution failed")
	}
	if response.status < 200 || response.status >= 300 || envelope.Code != 200 {
		businessCode := responseBusinessCode(response.header)
		if failure := classifyReadResponseFailure(input, response.status, envelope.Code, businessCode, true); failure != nil {
			return nil, failure
		}
		if (response.status == 200 || response.status == 400 || response.status == 409 || response.status >= 500 && response.status <= 599) &&
			(envelope.Code == 400 || envelope.Code == 409 || envelope.Code == 500) {
			if hint := BusinessErrorHint(input.Method, input.Path, businessCode); hint != "" {
				return nil, &businessHintError{hint: hint}
			}
		}
		return nil, errors.New("MCP API execution failed")
	}
	if len(envelope.Data) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Data), []byte("null")) {
		return map[string]any{}, nil
	}
	decoder = json.NewDecoder(bytes.NewReader(envelope.Data))
	decoder.UseNumber()
	if decoder.Decode(&result) != nil {
		return nil, errors.New("MCP API execution failed")
	}
	return result, nil
}

type boundedResponse struct {
	header   http.Header
	body     bytes.Buffer
	status   int
	overflow bool
}

func (r *boundedResponse) Header() http.Header { return r.header }
func (r *boundedResponse) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
}
func (r *boundedResponse) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	if r.overflow || len(data) > maxResponseBytes-r.body.Len() {
		r.overflow = true
		return 0, io.ErrShortBuffer
	}
	return r.body.Write(data)
}
