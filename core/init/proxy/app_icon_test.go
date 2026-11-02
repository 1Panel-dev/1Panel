package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type iconRedirectTransport func(*http.Request) (*http.Response, error)

func (f iconRedirectTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestLocalAppIconRedirectIsPassedToBrowser(t *testing.T) {
	const icon = "https://cdn.example/icon.png?signature=a%2Bb&v=42"
	calls := 0
	proxy := newLocalAgentProxy(iconRedirectTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.URL.Host != "unix" || req.URL.Path != "/api/v2/apps/icon/php" {
			t.Fatalf("proxy attempted to fetch a remote image: %s", req.URL)
		}
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": {icon}, "Cache-Control": {"no-cache"}},
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}))
	rw := httptest.NewRecorder()
	proxy.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/api/v2/apps/icon/php", nil))
	if rw.Code != http.StatusFound || rw.Header().Get("Location") != icon || rw.Header().Get("Cache-Control") != "no-cache" || calls != 1 {
		t.Fatalf("local proxy did not preserve the redirect: status=%d headers=%v calls=%d", rw.Code, rw.Header(), calls)
	}
}
