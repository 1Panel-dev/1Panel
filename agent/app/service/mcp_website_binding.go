package service

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"golang.org/x/net/idna"
	"gorm.io/gorm"
)

const (
	mcpWebsitePrivateDir = ".1panel-mcp"
	mcpWebsiteProxyName  = "proxy.conf"
	mcpWebsiteOwnerName  = "binding.json"
)

var mcpWebsiteMutation sync.Mutex

func init() {
	configureWebsiteManagedConfig(websiteManagedConfigPolicy{
		protectedDirectories: func() []string {
			return []string{filepath.Join(GetWebSiteRootDir(), mcpWebsitePrivateDir)}
		},
		backupIncludes: mcpWebsiteBackupIncludes,
	})
}

type MCPWebsiteBindingService struct {
	createWebsite func(request.WebsiteCreate) error
	updateHTTPS   func(context.Context, request.WebsiteHTTPSOp) error
	applyInclude  func(model.Website) error
	removeInclude func(model.Website) error
	checkReload   func() error
	openresty     func() (model.AppInstall, error)
	now           func() time.Time
}

func NewMCPWebsiteBindingService() *MCPWebsiteBindingService {
	s := &MCPWebsiteBindingService{
		createWebsite: func(req request.WebsiteCreate) error { return NewIWebsiteService().CreateWebsite(req) },
		updateHTTPS: func(ctx context.Context, req request.WebsiteHTTPSOp) error {
			if req.Enable {
				current, err := NewIWebsiteService().GetWebsiteHTTPS(req.WebsiteID)
				if err != nil {
					return err
				}
				defaults := newWebsiteCreateHTTPSOp(req.WebsiteSSLID)
				if current.Enable {
					req.SSLProtocol, req.Algorithm = current.SSLProtocol, current.Algorithm
					req.Hsts, req.HstsIncludeSubDomains, req.Http3 = current.Hsts, current.HstsIncludeSubDomains, current.Http3
					req.HttpConfig = current.HttpConfig
				} else {
					defaults.WebsiteID = req.WebsiteID
					req = defaults
				}
			}
			_, err := NewIWebsiteService().OpWebsiteHTTPS(ctx, req)
			return err
		},
		checkReload: func() error {
			installed, err := getAppInstallByKey(constant.AppOpenresty)
			if err != nil {
				return err
			}
			if installed.Status != constant.StatusRunning {
				return nil
			}
			if err = opNginx(installed.ContainerName, constant.NginxCheck); err != nil {
				return err
			}
			return opNginx(installed.ContainerName, constant.NginxReload)
		},
		openresty: mcpWebsiteOpenresty,
		now:       time.Now,
	}
	s.applyInclude = func(site model.Website) error { return s.setInclude(site, true) }
	s.removeInclude = func(site model.Website) error { return s.setInclude(site, false) }
	return s
}

func mcpWebsiteOpenresty() (model.AppInstall, error) {
	installed, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil || installed.ID == 0 || installed.Status != constant.StatusRunning {
		return installed, err
	}
	client, err := docker.NewDockerClient()
	if err != nil {
		return model.AppInstall{}, errors.New("OpenResty mount is unavailable")
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	inspected, err := client.ContainerInspect(ctx, installed.ContainerName)
	if err != nil || !mcpWebsiteHasWWWMount(inspected.Mounts, GetWebSiteRootDir()) {
		return model.AppInstall{}, errors.New("OpenResty mount is unavailable")
	}
	return installed, nil
}

func mcpWebsiteHasWWWMount(mounts []container.MountPoint, root string) bool {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	for _, mount := range mounts {
		if filepath.Clean(mount.Destination) != "/www" {
			continue
		}
		resolvedSource, sourceErr := filepath.EvalSymlinks(mount.Source)
		if sourceErr == nil && resolvedSource == resolvedRoot {
			return true
		}
	}
	return false
}

type mcpWebsiteOwner struct {
	BindingID   string `json:"bindingId"`
	WebsiteID   uint   `json:"websiteId"`
	KeyHash     string `json:"keyHash"`
	BindHash    string `json:"bindHash"`
	ProxyHash   string `json:"proxyHash"`
	RequestID   string `json:"requestId"`
	RequestHash string `json:"requestHash"`
	Operation   string `json:"operation"`
}

func mcpWebsiteHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func mcpWebsiteAlias(bindingID string) string { return "panel-mcp-" + bindingID }
func mcpWebsitePrivatePath(site model.Website, name string) string {
	return filepath.Join(GetWebSiteRootDir(), mcpWebsitePrivateDir, strings.TrimPrefix(site.Alias, "panel-mcp-"), name)
}

func mcpWebsiteInclude(site model.Website) string {
	return "/www/" + mcpWebsitePrivateDir + "/" + strings.TrimPrefix(site.Alias, "panel-mcp-") + "/" + mcpWebsiteProxyName
}

func mcpWebsiteBackupIncludes(site model.Website) []string {
	bindingID := strings.TrimPrefix(site.Alias, "panel-mcp-")
	if site.Alias != mcpWebsiteAlias(bindingID) || !validMCPWebsiteID(bindingID) {
		return nil
	}
	return []string{mcpWebsiteInclude(site)}
}

func mcpWebsitePrivateDirectory(site model.Website, create bool) error {
	for _, directory := range []string{filepath.Join(GetWebSiteRootDir(), mcpWebsitePrivateDir), filepath.Dir(mcpWebsitePrivatePath(site, mcpWebsiteProxyName))} {
		info, err := os.Lstat(directory)
		if os.IsNotExist(err) && create {
			if err = os.Mkdir(directory, 0700); err != nil {
				return err
			}
			info, err = os.Lstat(directory)
		}
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0700 {
			return errors.New("invalid MCP website private directory")
		}
	}
	return nil
}

func (s *MCPWebsiteBindingService) setInclude(site model.Website, enabled bool) error {
	configPath := GetSitePath(site, SiteConf)
	old, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	config, err := parser.NewStringParser(string(old)).Parse()
	if err != nil || len(config.FindServers()) == 0 {
		return errors.New("invalid website configuration")
	}
	server := config.FindServers()[0]
	include := mcpWebsiteInclude(site)
	for _, directive := range server.FindDirectives("include") {
		parameters := directive.GetParameters()
		if len(parameters) == 1 && strings.Trim(parameters[0], "\"'") == include {
			server.RemoveDirective("include", parameters)
		}
	}
	if enabled {
		server.UpdateDirective("include", []string{include})
	}
	updated := []byte(nginx.DumpConfig(config, nginx.IndentedStyle))
	if string(updated) == string(old) {
		return nil
	}
	if err = mcpWebsiteAtomicWrite(configPath, updated); err != nil {
		return err
	}
	if err = s.checkReload(); err != nil {
		_ = mcpWebsiteAtomicWrite(configPath, old)
		_ = s.checkReload()
		return err
	}
	return nil
}

func mcpWebsiteRemark(req request.MCPWebsiteBinding) string {
	return "1Panel MCP " + req.BindingID + " " + mcpWebsiteHash(req.ProxyKey)
}

func mcpWebsiteFailure(reason string, definitive, conflict bool) response.MCPWebsiteBinding {
	code, message := "MCP_WEBSITE_UNAVAILABLE", "MCP website operation is unavailable"
	if definitive {
		code, message = "MCP_WEBSITE_INVALID_ARGUMENTS", "Invalid MCP website configuration"
	}
	if conflict {
		code, message = "MCP_WEBSITE_CONFLICT", "MCP website configuration changed"
	}
	return response.MCPWebsiteBinding{State: "unavailable", Reason: reason, Code: code, Error: message, Definitive: definitive}
}

func validMCPWebsiteID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func mcpWebsiteDomain(value string) (string, int, error) {
	if value == "" || len(value) > 320 || strings.TrimSpace(value) != value || strings.ContainsAny(value, "/\\?#@\r\n\t ;{}\"'$\x00") {
		return "", 0, errors.New("invalid domain")
	}
	host, port := value, 0
	if parsedHost, parsedPort, err := net.SplitHostPort(value); err == nil {
		host = parsedHost
		port, err = strconv.Atoi(parsedPort)
		if err != nil || port < 1 || port > 65535 {
			return "", 0, errors.New("invalid port")
		}
	} else if strings.ContainsAny(value, ":[]") {
		return "", 0, errors.New("invalid domain")
	}
	host, err := idna.Lookup.ToASCII(strings.ToLower(host))
	if err != nil || len(host) > 253 || net.ParseIP(host) != nil {
		return "", 0, errors.New("invalid domain")
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", 0, errors.New("invalid domain")
		}
		for _, char := range label {
			if char != '-' && (char < 'a' || char > 'z') && (char < '0' || char > '9') {
				return "", 0, errors.New("invalid domain")
			}
		}
	}
	return host, port, nil
}

// Core can select a locally bound panel IP, but an external upstream is never
// accepted even on this private API. Host/SNI is a literal hostname, not Nginx code.
func validateMCPWebsiteUpstream(req request.MCPWebsiteBinding) bool {
	if strings.ContainsAny(req.Upstream, "\r\n\t ;{}\"'$#?") {
		return false
	}
	u, err := url.Parse(req.Upstream)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "/mcp" || u.RawPath != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Port() == "" {
		return false
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil || ip.IsUnspecified() {
		return false
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 {
		return false
	}
	local := ip.IsLoopback()
	if !local {
		addresses, err := net.InterfaceAddrs()
		if err != nil {
			return false
		}
		for _, address := range addresses {
			if network, ok := address.(*net.IPNet); ok && network.IP.Equal(ip) {
				local = true
				break
			}
		}
	}
	if !local || len(req.ProxyKey) != 64 {
		return false
	}
	if _, err = hex.DecodeString(req.ProxyKey); err != nil {
		return false
	}
	if req.CoreHost == "" || len(req.CoreHost) > 320 || strings.ContainsAny(req.CoreHost, "/\\?#@\r\n\t ;{}\"'$\x00") {
		return false
	}
	host := req.CoreHost
	if parsed, _, splitErr := net.SplitHostPort(host); splitErr == nil {
		host = parsed
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = strings.Trim(host, "[]")
	}
	if net.ParseIP(host) == nil {
		if _, _, domainErr := mcpWebsiteDomain(host); domainErr != nil {
			return false
		}
	}
	return true
}

func mcpWebsiteProxy(req request.MCPWebsiteBinding) []byte {
	host := req.CoreHost
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	host = strings.Trim(host, "[]")
	var tlsDirectives string
	if strings.HasPrefix(req.Upstream, "https://") {
		tlsDirectives = "    proxy_ssl_server_name on;\n    proxy_ssl_name " + host + ";\n"
	}
	return []byte("# Managed by 1Panel MCP: " + req.BindingID + "\n" +
		"location = /mcp {\n" +
		"    proxy_pass " + req.Upstream + ";\n" +
		"    proxy_http_version 1.1;\n" +
		"    proxy_set_header Host " + req.CoreHost + ";\n" +
		"    proxy_set_header Connection \"\";\n" +
		"    proxy_set_header X-Panel-MCP-Website-Key " + req.ProxyKey + ";\n" +
		"    proxy_set_header X-Panel-MCP-Website-Peer $realip_remote_addr;\n" +
		"    proxy_set_header X-Forwarded-For $http_x_forwarded_for;\n" +
		"    proxy_set_header X-Real-IP $http_x_real_ip;\n" +
		"    proxy_set_header Forwarded $http_forwarded;\n" +
		"    proxy_set_header Authorization $http_authorization;\n" +
		"    proxy_set_header Origin $http_origin;\n" +
		"    proxy_set_header Accept $http_accept;\n" +
		"    proxy_set_header Content-Type $http_content_type;\n" +
		"    proxy_set_header MCP-Protocol-Version $http_mcp_protocol_version;\n" +
		"    proxy_buffering off;\n    proxy_request_buffering off;\n    proxy_cache off;\n" +
		"    proxy_read_timeout 360s;\n    proxy_send_timeout 360s;\n    proxy_connect_timeout 5s;\n" +
		"    proxy_next_upstream off;\n    client_max_body_size 1m;\n" + tlsDirectives + "}\n")
}

func (s *MCPWebsiteBindingService) validCertificate(id uint, domain string) bool {
	certificate, err := websiteSSLRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return false
	}
	pair, err := tls.X509KeyPair([]byte(certificate.Pem), []byte(certificate.PrivateKey))
	if err != nil || len(pair.Certificate) == 0 {
		return false
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil || s.now().Before(leaf.NotBefore) || !s.now().Before(leaf.NotAfter) || leaf.VerifyHostname(domain) != nil {
		return false
	}
	if len(leaf.ExtKeyUsage) == 0 {
		return true
	}
	for _, usage := range leaf.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth || usage == x509.ExtKeyUsageAny {
			return true
		}
	}
	return false
}

func mcpWebsiteReadOwner(site model.Website) (mcpWebsiteOwner, error) {
	var owner mcpWebsiteOwner
	if err := mcpWebsitePrivateDirectory(site, false); err != nil {
		return owner, err
	}
	path := mcpWebsitePrivatePath(site, mcpWebsiteOwnerName)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 8192 {
		return owner, errors.New("invalid MCP website ownership")
	}
	data, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(data, &owner) != nil {
		return owner, errors.New("invalid MCP website ownership")
	}
	return owner, nil
}

func mcpWebsiteWriteOwner(site model.Website, owner mcpWebsiteOwner) error {
	data, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	return mcpWebsiteAtomicWrite(mcpWebsitePrivatePath(site, mcpWebsiteOwnerName), data)
}

func mcpWebsiteAtomicWrite(path string, data []byte) error {
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return errors.New("managed file is not regular")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".panel-mcp-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err = temp.Chmod(0600); err == nil {
		_, err = temp.Write(data)
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func mcpWebsiteCurrentProxy(site model.Website, owner mcpWebsiteOwner) ([]byte, error) {
	if err := mcpWebsitePrivateDirectory(site, false); err != nil {
		return nil, err
	}
	path := mcpWebsitePrivatePath(site, mcpWebsiteProxyName)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 16<<10 {
		return nil, errors.New("managed proxy is missing or invalid")
	}
	data, err := os.ReadFile(path)
	if err != nil || mcpWebsiteHash(string(data)) != owner.ProxyHash {
		return nil, errors.New("managed proxy was modified")
	}
	return data, nil
}

func mcpWebsiteReadProxy(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16<<10 {
		return nil, errors.New("managed proxy is not a regular file")
	}
	return os.ReadFile(path)
}

func mcpWebsitePayloadHash(req request.MCPWebsiteBinding) string {
	return mcpWebsiteHash(fmt.Sprintf("%s\n%s\n%t\n%d", req.Operation, req.Domain, req.EnableSSL, req.SSLID))
}

func mcpWebsiteBindHash(req request.MCPWebsiteBinding) string {
	return mcpWebsiteHash(fmt.Sprintf("%s\n%t\n%d", req.Domain, req.EnableSSL, req.SSLID))
}

func (s *MCPWebsiteBindingService) matchesInitialSite(req request.MCPWebsiteBinding, site model.Website) bool {
	host, port, err := mcpWebsiteDomain(req.Domain)
	if err != nil || (site.Protocol == constant.ProtocolHTTPS) != req.EnableSSL || site.WebsiteSSLID != req.SSLID {
		return false
	}
	if port == 0 {
		installed, installErr := s.openresty()
		if installErr != nil {
			return false
		}
		port = installed.HttpPort
		if req.EnableSSL {
			port = installed.HttpsPort
		}
	}
	for _, domain := range site.Domains {
		if domain.Domain == host && domain.Port == port {
			return true
		}
	}
	return false
}

func (s *MCPWebsiteBindingService) load(req request.MCPWebsiteBinding) (model.Website, error) {
	if req.WebsiteID > 0 {
		return websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	}
	return websiteRepo.GetFirst(websiteRepo.WithAlias(mcpWebsiteAlias(req.BindingID)))
}

func mcpWebsiteProxyIncluded(site model.Website) bool {
	server, err := getServer(site)
	if err != nil {
		return false
	}
	include := mcpWebsiteInclude(site)
	for _, directive := range server.FindDirectives("include") {
		parameters := directive.GetParameters()
		if len(parameters) == 1 && strings.Trim(parameters[0], "\"'") == include {
			return true
		}
	}
	return false
}

func (s *MCPWebsiteBindingService) status(req request.MCPWebsiteBinding, site model.Website, owner mcpWebsiteOwner) response.MCPWebsiteBinding {
	result := response.MCPWebsiteBinding{State: "bound", WebsiteID: site.ID, BindingID: req.BindingID, RequestID: owner.RequestID,
		Domain: site.PrimaryDomain, EnableSSL: site.Protocol == constant.ProtocolHTTPS, SSLID: site.WebsiteSSLID, WebsiteStatus: site.Status}
	if !result.EnableSSL {
		result.SSLID = 0
	}
	if owner.Operation == "unbind" {
		result.State = "unbound"
		return result
	}
	if _, err := mcpWebsiteCurrentProxy(site, owner); err != nil {
		result.State, result.Reason = "unavailable", "proxy_modified"
		return result
	}
	if mcpWebsiteHash(string(mcpWebsiteProxy(req))) != owner.ProxyHash {
		result.State, result.Reason = "unavailable", "upstream_changed"
		return result
	}
	if site.Status != constant.WebRunning {
		result.State, result.Reason = "unavailable", "website_stopped"
		return result
	}
	if !mcpWebsiteProxyIncluded(site) {
		result.State, result.Reason = "unavailable", "proxy_not_loaded"
		return result
	}
	host, _, err := mcpWebsiteDomain(site.PrimaryDomain)
	if err != nil {
		result.State, result.Reason = "unavailable", "website_modified"
		return result
	}
	if result.EnableSSL && !s.validCertificate(site.WebsiteSSLID, host) {
		result.State, result.Reason = "unavailable", "certificate_invalid"
		return result
	}
	installed, err := s.openresty()
	if err != nil || installed.ID == 0 || installed.Status != constant.StatusRunning {
		result.State, result.Reason = "unavailable", "openresty_unavailable"
		return result
	}
	matching, matchingSSL := 0, 0
	for _, domain := range site.Domains {
		if domain.Domain != host || domain.Port < 1 {
			continue
		}
		if matching == 0 || !domain.SSL {
			matching = domain.Port
		}
		if domain.SSL {
			matchingSSL = domain.Port
		}
	}
	if matching == 0 {
		result.State, result.Reason = "unavailable", "website_modified"
		return result
	}
	port := matching
	if result.EnableSSL {
		port = matchingSSL
		if port == 0 {
			plan := buildWebsiteTLSPlan(site.Domains, installed.HttpPort, installed.HttpsPort)
			if slices.Contains(plan.httpsPorts, installed.HttpsPort) {
				port = installed.HttpsPort
			} else if len(plan.httpsPorts) > 0 {
				port = plan.httpsPorts[0]
			}
		}
	}
	if port == 0 {
		result.State, result.Reason = "unavailable", "website_modified"
		return result
	}
	scheme := "http"
	if result.EnableSSL {
		scheme = "https"
	}
	address := host
	if result.EnableSSL && port != 443 || !result.EnableSSL && port != 80 {
		address = net.JoinHostPort(host, strconv.Itoa(port))
	}
	result.Endpoint = (&url.URL{Scheme: scheme, Host: address, Path: "/mcp"}).String()
	return result
}

func (s *MCPWebsiteBindingService) Handle(ctx context.Context, req request.MCPWebsiteBinding) (result response.MCPWebsiteBinding) {
	mcpWebsiteMutation.Lock()
	defer mcpWebsiteMutation.Unlock()
	mutationStarted := false
	defer func() {
		if mutationStarted && result.Error != "" {
			result.Definitive = false
		}
	}()
	if req.Operation == "bind" && validMCPWebsiteID(req.BindingID) {
		_, existingErr := websiteRepo.GetFirst(websiteRepo.WithAlias(mcpWebsiteAlias(req.BindingID)))
		_, privateErr := os.Lstat(filepath.Join(GetWebSiteRootDir(), mcpWebsitePrivateDir, req.BindingID))
		mutationStarted = !errors.Is(existingErr, gorm.ErrRecordNotFound) || !os.IsNotExist(privateErr)
	}
	if ctx.Err() != nil {
		return mcpWebsiteFailure("request_cancelled", true, false)
	}
	if req.Operation != "status" && req.Operation != "bind" && req.Operation != "update" && req.Operation != "unbind" && req.Operation != "sync" {
		return mcpWebsiteFailure("invalid_operation", true, false)
	}
	if req.BindingID == "" && req.WebsiteID == 0 && req.Operation == "status" {
		return response.MCPWebsiteBinding{State: "unbound"}
	}
	if !validMCPWebsiteID(req.BindingID) || !validateMCPWebsiteUpstream(req) {
		return mcpWebsiteFailure("invalid_configuration", true, false)
	}
	if req.Operation != "status" && req.Operation != "sync" && !validMCPWebsiteID(req.RequestID) {
		return mcpWebsiteFailure("invalid_request", true, false)
	}
	if req.Operation == "bind" {
		host, port, err := mcpWebsiteDomain(req.Domain)
		if err != nil || (!req.EnableSSL && req.SSLID != 0) || (req.EnableSSL && (req.SSLID == 0 || !s.validCertificate(req.SSLID, host))) {
			return mcpWebsiteFailure("invalid_domain_or_certificate", true, false)
		}
		req.Domain = host
		if port != 0 {
			req.Domain = net.JoinHostPort(host, strconv.Itoa(port))
		}
	}
	site, err := s.load(req)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if req.Operation == "status" || req.Operation == "unbind" {
			if req.WebsiteID == 0 || req.Operation == "unbind" {
				return response.MCPWebsiteBinding{State: "unbound", BindingID: req.BindingID, RequestID: req.RequestID}
			}
			return response.MCPWebsiteBinding{State: "unavailable", WebsiteID: req.WebsiteID, BindingID: req.BindingID, Reason: "website_missing"}
		}
		if req.Operation != "bind" {
			return mcpWebsiteFailure("website_missing", true, true)
		}
		installed, installErr := s.openresty()
		if installErr != nil || installed.ID == 0 || installed.Status != constant.StatusRunning {
			return mcpWebsiteFailure("openresty_unavailable", true, false)
		}
		group, groupErr := groupRepo.Get(groupRepo.WithByWebsiteDefault())
		if groupErr != nil || group.ID == 0 {
			return mcpWebsiteFailure("website_group_unavailable", true, false)
		}
		domain, port, _ := mcpWebsiteDomain(req.Domain)
		if port == 0 {
			port = installed.HttpPort
			if req.EnableSSL {
				port = installed.HttpsPort
			}
		}
		if existing, conflictErr := websiteDomainRepo.GetFirst(websiteDomainRepo.WithDomain(domain), websiteDomainRepo.WithPort(port)); conflictErr == nil && existing.ID > 0 {
			return mcpWebsiteFailure("domain_conflict", true, true)
		} else if conflictErr != nil && !errors.Is(conflictErr, gorm.ErrRecordNotFound) {
			return mcpWebsiteFailure("website_unavailable", false, false)
		}
		create := request.WebsiteCreate{Alias: mcpWebsiteAlias(req.BindingID), Type: constant.Static, Remark: mcpWebsiteRemark(req), WebsiteGroupID: group.ID,
			Domains: []request.WebsiteDomain{{Domain: domain, Port: port, SSL: req.EnableSSL}}, SSLConfig: request.SSLConfig{EnableSSL: req.EnableSSL, WebsiteSSLID: req.SSLID}}
		mutationStarted = true
		if createErr := s.createWebsite(create); createErr != nil {
			return mcpWebsiteFailure("website_create_failed", false, false)
		}
		site, err = s.load(req)
	}
	if err != nil {
		return mcpWebsiteFailure("website_unavailable", false, false)
	}
	if site.Alias != mcpWebsiteAlias(req.BindingID) {
		return mcpWebsiteFailure("website_ownership_changed", true, true)
	}
	owner, ownerErr := mcpWebsiteReadOwner(site)
	if ownerErr == nil && owner.RequestID == req.RequestID && req.RequestID != "" {
		mutationStarted = true
	}
	if req.Operation == "update" {
		if req.Domain != "" && req.Domain != site.PrimaryDomain {
			return mcpWebsiteFailure("domain_changed", true, true)
		}
		req.Domain = site.PrimaryDomain
		host, _, domainErr := mcpWebsiteDomain(req.Domain)
		if domainErr != nil || (!req.EnableSSL && req.SSLID != 0) || (req.EnableSSL && (req.SSLID == 0 || !s.validCertificate(req.SSLID, host))) {
			return mcpWebsiteFailure("invalid_domain_or_certificate", true, false)
		}
	}
	if info, dirErr := os.Lstat(GetSitePath(site, SiteDir)); dirErr == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return mcpWebsiteFailure("website_directory_changed", true, true)
	} else if dirErr != nil && !os.IsNotExist(dirErr) {
		return mcpWebsiteFailure("website_unavailable", false, false)
	}
	if ownerErr != nil {
		if (req.Operation != "bind" && req.Operation != "unbind") || site.Remark != mcpWebsiteRemark(req) {
			return mcpWebsiteFailure("website_ownership_changed", true, true)
		}
		if req.Operation == "bind" && !s.matchesInitialSite(req, site) {
			return mcpWebsiteFailure("binding_request_changed", true, true)
		}
		owner = mcpWebsiteOwner{BindingID: req.BindingID, WebsiteID: site.ID, KeyHash: mcpWebsiteHash(req.ProxyKey), BindHash: mcpWebsiteBindHash(req)}
	} else if owner.BindingID != req.BindingID || owner.WebsiteID != site.ID || owner.KeyHash != mcpWebsiteHash(req.ProxyKey) {
		return mcpWebsiteFailure("website_ownership_changed", true, true)
	}
	if req.Operation == "status" {
		return s.status(req, site, owner)
	}
	operationHash := mcpWebsitePayloadHash(req)
	if owner.RequestID == req.RequestID && req.RequestID != "" {
		if owner.RequestHash != operationHash {
			return mcpWebsiteFailure("operation_request_changed", true, true)
		}
		if owner.Operation == "unbind" || owner.ProxyHash == mcpWebsiteHash(string(mcpWebsiteProxy(req))) {
			return s.status(req, site, owner)
		}
	}
	if req.Operation == "bind" && owner.BindHash != mcpWebsiteBindHash(req) {
		return mcpWebsiteFailure("binding_request_changed", true, true)
	}
	if owner.ProxyHash != "" && owner.Operation != "unbind" && req.Operation != "unbind" && !mcpWebsiteProxyIncluded(site) {
		return mcpWebsiteFailure("proxy_not_loaded", true, true)
	}
	proxyPath := mcpWebsitePrivatePath(site, mcpWebsiteProxyName)
	if err = mcpWebsitePrivateDirectory(site, true); err != nil {
		return mcpWebsiteFailure("proxy_directory_changed", true, true)
	}
	if ownerErr != nil {
		mutationStarted = true
		if err = mcpWebsiteWriteOwner(site, owner); err != nil {
			return mcpWebsiteFailure("ownership_write_failed", false, false)
		}
	}
	var oldProxy []byte
	if owner.ProxyHash != "" && owner.Operation != "unbind" {
		oldProxy, err = mcpWebsiteCurrentProxy(site, owner)
		if err != nil {
			actual, readErr := mcpWebsiteReadProxy(proxyPath)
			if req.Operation == "unbind" && os.IsNotExist(readErr) {
				oldProxy = nil
			} else if readErr == nil && string(actual) == string(mcpWebsiteProxy(req)) {
				oldProxy = actual
			} else {
				return mcpWebsiteFailure("proxy_modified", true, true)
			}
		}
	} else if _, statErr := os.Lstat(proxyPath); statErr == nil {
		actual, readErr := mcpWebsiteReadProxy(proxyPath)
		if readErr != nil || string(actual) != string(mcpWebsiteProxy(req)) {
			return mcpWebsiteFailure("proxy_already_exists", true, true)
		}
		oldProxy = actual
	} else if !os.IsNotExist(statErr) {
		return mcpWebsiteFailure("proxy_unavailable", false, false)
	}
	if req.Operation == "unbind" {
		if mcpWebsiteProxyIncluded(site) {
			mutationStarted = true
			if err = s.removeInclude(site); err != nil {
				return mcpWebsiteFailure("proxy_reload_failed", false, false)
			}
		}
		if len(oldProxy) > 0 {
			mutationStarted = true
			if err = os.Remove(proxyPath); err != nil {
				return mcpWebsiteFailure("proxy_remove_failed", false, false)
			}
		}
		owner.Operation, owner.RequestID, owner.RequestHash = "unbind", req.RequestID, operationHash
		mutationStarted = true
		if err = mcpWebsiteWriteOwner(site, owner); err != nil {
			return mcpWebsiteFailure("ownership_write_failed", false, false)
		}
		return s.status(req, site, owner)
	}
	if req.Operation == "update" {
		if req.Domain != site.PrimaryDomain {
			return mcpWebsiteFailure("domain_changed", true, true)
		}
		if site.Protocol == constant.ProtocolHTTPS && !req.EnableSSL || site.Protocol != constant.ProtocolHTTPS && req.EnableSSL || req.EnableSSL && site.WebsiteSSLID != req.SSLID {
			mutationStarted = true
			if err = s.updateHTTPS(context.WithoutCancel(ctx), request.WebsiteHTTPSOp{WebsiteID: site.ID, Enable: req.EnableSSL, Type: constant.SSLExisted, WebsiteSSLID: req.SSLID, HttpConfig: constant.HTTPSOnly}); err != nil {
				return mcpWebsiteFailure("certificate_update_failed", false, false)
			}
			site, err = s.load(req)
			if err != nil {
				return mcpWebsiteFailure("website_unavailable", false, false)
			}
		}
	}
	newProxy := mcpWebsiteProxy(req)
	mutationStarted = true
	if err = mcpWebsiteAtomicWrite(proxyPath, newProxy); err != nil {
		return mcpWebsiteFailure("proxy_write_failed", false, false)
	}
	if err = s.applyInclude(site); err != nil {
		if len(oldProxy) == 0 {
			_ = os.Remove(proxyPath)
		} else {
			_ = mcpWebsiteAtomicWrite(proxyPath, oldProxy)
		}
		_ = s.checkReload()
		return mcpWebsiteFailure("proxy_reload_failed", false, false)
	}
	owner.ProxyHash = mcpWebsiteHash(string(newProxy))
	if req.Operation != "sync" || req.RequestID != "" {
		owner.RequestID, owner.RequestHash, owner.Operation = req.RequestID, operationHash, req.Operation
	}
	if err = mcpWebsiteWriteOwner(site, owner); err != nil {
		return mcpWebsiteFailure("ownership_write_failed", false, false)
	}
	return s.status(req, site, owner)
}
