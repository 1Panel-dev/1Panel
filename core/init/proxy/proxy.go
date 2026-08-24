package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"slices"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/utils/clientip"
	"github.com/1Panel-dev/1Panel/core/utils/publicshare"
)

const SockPath = "/etc/1panel/agent.sock"

var (
	LocalAgentProxy *httputil.ReverseProxy
)

func Init() {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	dialUnix := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, "unix", SockPath)
	}
	transport := &http.Transport{
		DialContext:         dialUnix,
		ForceAttemptHTTP2:   false,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     30 * time.Second,
	}
	LocalAgentProxy = newLocalAgentProxy(transport)
}

func newLocalAgentProxy(transport http.RoundTripper) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(proxyReq *httputil.ProxyRequest) {
			if proxyReq.In.Form == nil {
				proxyReq.Out.URL.RawQuery = proxyReq.In.URL.RawQuery
			}
			if publicshare.IsAPI(proxyReq.In.URL.Path) {
				proxyReq.SetXForwarded()
				clientip.ReplaceForwardingHeaders(
					proxyReq.Out.Header,
					clientip.FromRemoteAddr(proxyReq.In.RemoteAddr),
				)
			} else {
				restoreLegacyForwardingHeaders(proxyReq)
			}
			proxyReq.Out.URL.Scheme = "http"
			proxyReq.Out.URL.Host = "unix"
		},
		Transport: transport,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			rw.WriteHeader(http.StatusBadGateway)
			_, _ = rw.Write([]byte("Bad Gateway: " + err.Error()))
		},
	}
}

func restoreLegacyForwardingHeaders(proxyReq *httputil.ProxyRequest) {
	restoreInboundHeader(proxyReq, "Forwarded")
	restoreLegacyForwardedFor(proxyReq)
	restoreLegacyForwardedMetadata(proxyReq, "X-Forwarded-Host", proxyReq.In.Host)

	forwardedProto := "http"
	if proxyReq.In.TLS != nil {
		forwardedProto = "https"
	}
	restoreLegacyForwardedMetadata(proxyReq, "X-Forwarded-Proto", forwardedProto)
}

func restoreLegacyForwardedFor(proxyReq *httputil.ProxyRequest) {
	restoreInboundHeader(proxyReq, "X-Forwarded-For")
	clientIP, _, err := net.SplitHostPort(proxyReq.In.RemoteAddr)
	if err != nil {
		return
	}
	prior, ok := proxyReq.Out.Header["X-Forwarded-For"]
	omit := ok && prior == nil
	if len(prior) > 0 {
		clientIP = strings.Join(prior, ", ") + ", " + clientIP
	}
	if !omit {
		proxyReq.Out.Header.Set("X-Forwarded-For", clientIP)
	}
}

func restoreLegacyForwardedMetadata(proxyReq *httputil.ProxyRequest, name, fallback string) {
	if !isConnectionHeader(proxyReq.In.Header, name) && proxyReq.In.Header.Get(name) != "" {
		restoreInboundHeader(proxyReq, name)
		return
	}
	proxyReq.Out.Header.Del(name)
	if fallback != "" {
		proxyReq.Out.Header.Set(name, fallback)
	}
}

func restoreInboundHeader(proxyReq *httputil.ProxyRequest, name string) {
	proxyReq.Out.Header.Del(name)
	if isConnectionHeader(proxyReq.In.Header, name) {
		return
	}
	canonicalName := http.CanonicalHeaderKey(name)
	if values, ok := proxyReq.In.Header[canonicalName]; ok {
		proxyReq.Out.Header[canonicalName] = slices.Clone(values)
	}
}

func isConnectionHeader(header http.Header, name string) bool {
	for _, value := range header.Values("Connection") {
		for _, token := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), name) {
				return true
			}
		}
	}
	return false
}
