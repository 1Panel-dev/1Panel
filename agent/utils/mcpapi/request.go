package mcpapi

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/google/uuid"
)

const (
	InvocationHeader = "X-Panel-MCP-Invocation"
	NodeHeader       = "X-Panel-MCP-Node"
	MethodHeader     = "X-Panel-MCP-Method"
	PathHeader       = "X-Panel-MCP-Path"
	CapabilitiesPath = "/api/v2/internal/mcp/capabilities"
	ContractVersion  = 4
)

type verifiedNodeTransportKey struct{}

func BindVerifiedNodeTransport(request *http.Request) *http.Request {
	if request.TLS == nil || !request.TLS.HandshakeComplete || len(request.TLS.VerifiedChains) == 0 {
		return request
	}
	return request.WithContext(context.WithValue(request.Context(), verifiedNodeTransportKey{}, request.TLS))
}

func validNodeID(id string) bool {
	if id == "local" {
		return true
	}
	value, err := strconv.ParseUint(id, 10, 64)
	return err == nil && value != 0 && strconv.FormatUint(value, 10) == id
}

func TrustedTransport(request *http.Request) bool {
	if global.IsMaster {
		local, ok := request.Context().Value(http.LocalAddrContextKey).(net.Addr)
		return ok && local.Network() == "unix"
	}
	verified := request.Context().Value(verifiedNodeTransportKey{})
	return request.TLS != nil && verified == request.TLS
}

func Check(request *http.Request) (present, trusted bool) {
	present = false
	for _, name := range []string{InvocationHeader, NodeHeader, MethodHeader, PathHeader} {
		if len(request.Header.Values(name)) != 0 {
			present = true
		}
	}
	if !present {
		return false, false
	}
	for _, name := range []string{InvocationHeader, NodeHeader, MethodHeader, PathHeader} {
		if len(request.Header.Values(name)) != 1 {
			return true, false
		}
	}
	if _, err := uuid.Parse(request.Header.Get(InvocationHeader)); err != nil {
		return true, false
	}
	nodeID := request.Header.Get(NodeHeader)
	if !validNodeID(nodeID) || request.Header.Get(MethodHeader) != request.Method || request.Header.Get(PathHeader) != request.URL.Path || request.URL.RawQuery != "" {
		return true, false
	}
	return true, TrustedTransport(request) && ((nodeID == "local") == global.IsMaster)
}

func NodeID(request *http.Request) (string, bool) {
	_, trusted := Check(request)
	if !trusted {
		return "", false
	}
	return request.Header.Get(NodeHeader), true
}
