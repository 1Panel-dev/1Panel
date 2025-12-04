package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
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
	LocalAgentProxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "unix"
		},
		Transport: transport,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			rw.WriteHeader(http.StatusBadGateway)
			_, _ = rw.Write([]byte("Bad Gateway: " + err.Error()))
		},
	}
}
