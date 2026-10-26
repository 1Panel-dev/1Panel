package metadata

import (
	"context"
	"net/http"
)

const BusinessCodeHeader = "X-Panel-MCP-Business-Code"

type verifierKey struct{}

func BindRequest(request *http.Request, verify func(*http.Request) bool) *http.Request {
	return request.WithContext(context.WithValue(request.Context(), verifierKey{}, verify))
}

func TrustedRequest(request *http.Request) bool {
	if request == nil {
		return false
	}
	verify, ok := request.Context().Value(verifierKey{}).(func(*http.Request) bool)
	return ok && verify != nil && verify(request)
}
