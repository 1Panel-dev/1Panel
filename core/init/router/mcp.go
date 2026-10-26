package router

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/gin-gonic/gin"
)

func registerMCPRuntime(r *gin.Engine) {
	r.GET(contract.RuntimePrefix+"/tools", v2.MCPRuntimeTools)
	r.POST(contract.RuntimePrefix+"/invoke", v2.MCPRuntimeInvoke)
	r.POST(contract.RuntimePrefix+"/admit", v2.MCPRuntimeAdmit)
	r.POST(contract.RuntimePrefix+"/complete", v2.MCPRuntimeComplete)
}
func registerMCPTransport(r *gin.Engine) {
	for _, method := range []string{http.MethodPost, http.MethodGet, http.MethodDelete} {
		r.Handle(method, "/mcp", proxyMCPTransport)
	}
}
func proxyMCPTransport(c *gin.Context) {
	if err := contract.AdmitIngress(time.Now()); err != nil {
		v2.MCPRuntimeError(c, err)
		return
	}
	if c.GetHeader("Origin") != "" {
		v2.MCPRuntimeError(c, contract.Forbidden())
		return
	}
	token, err := service.MCPBearer(c)
	if err != nil {
		v2.MCPRuntimeError(c, err)
		return
	}
	if _, _, err = service.NewMCPService().Authenticate(c, token); err != nil {
		v2.MCPRuntimeError(c, err)
		return
	}
	source := c.GetString("MCP_CLIENT_IP")
	if source == "" {
		source = common.GetRealClientIP(c)
	}
	proof, err := contract.IssueAccessProof(token, source, time.Now())
	if err != nil {
		v2.MCPRuntimeError(c, contract.Forbidden())
		return
	}
	settings, err := service.NewMCPRuntimeService().Settings()
	if err != nil {
		v2.MCPRuntimeError(c, err)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	transport := &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: 3 * time.Second}).DialContext, ResponseHeaderTimeout: 355 * time.Second}
	defer transport.CloseIdleConnections()
	proxy := &httputil.ReverseProxy{Transport: transport, Rewrite: func(p *httputil.ProxyRequest) {
		p.Out.URL.Scheme = "http"
		p.Out.URL.Host = net.JoinHostPort("127.0.0.1", strconv.Itoa(settings.ListenPort))
		p.Out.URL.Path = "/mcp"
		p.Out.URL.RawPath = ""
		p.Out.URL.RawQuery = ""
		p.Out.Host = p.Out.URL.Host
		p.Out.Header = make(http.Header)
		for _, key := range []string{"Content-Type", "Accept", "MCP-Protocol-Version", "Accept-Language"} {
			if value := p.In.Header.Get(key); value != "" {
				p.Out.Header.Set(key, value)
			}
		}
		p.Out.Header.Set("Authorization", "Bearer "+token)
		p.Out.Header.Set(contract.AccessProofHeader, proof)
	}, ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"code":"MCP_NODE_UNAVAILABLE","message":"MCP adapter is unavailable"}`))
	}}
	c.Header("Cache-Control", "no-store")
	proxy.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}

func StartMCPBridge() error {
	if err := prepareMCPDockerSocketDir(contract.DockerSocketDir); err != nil {
		return err
	}
	return startMCPBridge(contract.DockerCoreSocket)
}

func prepareMCPDockerSocketDir(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() {
		return errors.New("MCP container bridge directory must be a real directory")
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() != "core.sock" || entry.Type()&os.ModeSocket == 0 {
			return errors.New("MCP container bridge directory must contain only its socket")
		}
	}
	return nil
}
func startMCPBridge(socket string) error {
	if err := os.MkdirAll(filepath.Dir(socket), 0700); err != nil {
		return err
	}
	if info, err := os.Lstat(socket); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return errors.New("MCP bridge path is not a socket")
		}
		conn, dialErr := net.DialTimeout("unix", socket, time.Second)
		if dialErr == nil {
			conn.Close()
			return errors.New("MCP bridge is already running")
		}
		if err = os.Remove(socket); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	if err = os.Chmod(socket, 0600); err != nil {
		listener.Close()
		return err
	}
	r := gin.New()
	r.Use(gin.Recovery())
	registerMCPRuntime(r)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "contractVersion": contract.ContractVersion}) })
	server := &http.Server{Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 340 * time.Second, WriteTimeout: 340 * time.Second, IdleTimeout: 30 * time.Second, BaseContext: func(net.Listener) context.Context { return context.Background() }}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) && global.LOG != nil {
			global.LOG.Error("MCP runtime bridge stopped")
		}
	}()
	return nil
}
