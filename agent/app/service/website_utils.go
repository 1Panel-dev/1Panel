package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/utils/xpack"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func handleChineseDomain(domain string) (string, error) {
	if common.ContainsChinese(domain) {
		return common.PunycodeEncode(domain)
	}
	return domain, nil
}

func createIndexFile(website *model.Website, runtime *model.Runtime) error {
	var (
		indexPath      string
		indexContent   string
		websiteService = NewIWebsiteService()
		indexFolder    = GetSitePath(*website, SiteIndexDir)
	)

	switch website.Type {
	case constant.Static:
		indexPath = path.Join(indexFolder, "index.html")
		indexHtml, _ := websiteService.GetDefaultHtml("index")
		indexContent = indexHtml.Content
	case constant.Runtime:
		if runtime.Type == constant.RuntimePHP {
			indexPath = path.Join(indexFolder, "index.php")
			indexPhp, _ := websiteService.GetDefaultHtml("php")
			indexContent = indexPhp.Content
		}
	}

	fileOp := files.NewFileOp()
	if !fileOp.Stat(indexFolder) {
		if err := fileOp.CreateDir(indexFolder, constant.DirPerm); err != nil {
			return err
		}
	}
	if !fileOp.Stat(indexPath) {
		if err := fileOp.CreateFile(indexPath); err != nil {
			return err
		}
	}
	if website.Type == constant.Runtime && runtime.Resource == constant.ResourceAppstore {
		if err := chownRootDir(indexFolder); err != nil {
			return err
		}
	}
	if err := fileOp.WriteFile(indexPath, strings.NewReader(indexContent), constant.DirPerm); err != nil {
		return err
	}

	html404, _ := websiteService.GetDefaultHtml("404")
	path404 := path.Join(indexFolder, "404.html")
	if err := fileOp.WriteFile(path404, strings.NewReader(html404.Content), constant.DirPerm); err != nil {
		return err
	}

	return nil
}

func createProxyFile(website *model.Website) error {
	proxyFolder := GetSitePath(*website, SiteProxyDir)
	filePath := path.Join(proxyFolder, "root.conf")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(proxyFolder) {
		if err := fileOp.CreateDir(proxyFolder, constant.DirPerm); err != nil {
			return err
		}
	}
	if !fileOp.Stat(filePath) {
		if err := fileOp.CreateFile(filePath); err != nil {
			return err
		}
	}
	config, err := parser.NewStringParser(string(nginx_conf.Proxy)).Parse()
	if err != nil {
		return err
	}
	config.FilePath = filePath
	directives := config.Directives
	location, ok := directives[0].(*components.Location)
	if !ok {
		return errors.New("error")
	}
	location.ChangePath("^~", "/")
	location.UpdateDirective("proxy_pass", []string{website.Proxy})
	location.UpdateDirective("proxy_set_header", []string{"Host", "$host"})
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}
	return nil
}

func createWebsiteFolder(website *model.Website, runtime *model.Runtime) error {
	siteFolder := GteSiteDir(website.Alias)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(siteFolder) {
		if err := fileOp.CreateDir(siteFolder, constant.DirPerm); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "log"), constant.DirPerm); err != nil {
			return err
		}
		if err := fileOp.CreateFile(path.Join(siteFolder, "log", "access.log")); err != nil {
			return err
		}
		if err := fileOp.CreateFile(path.Join(siteFolder, "log", "error.log")); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "index"), constant.DirPerm); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "ssl"), constant.DirPerm); err != nil {
			return err
		}
		if website.Type == constant.Runtime {
			if runtime.Type == constant.RuntimePHP && runtime.Resource == constant.ResourceLocal {
				phpPoolDir := path.Join(siteFolder, "php-pool")
				if err := fileOp.CreateDir(phpPoolDir, constant.DirPerm); err != nil {
					return err
				}
				if err := fileOp.CreateFile(path.Join(phpPoolDir, "php-fpm.sock")); err != nil {
					return err
				}
			}
		}
		if website.Type == constant.Static || (website.Type == constant.Runtime && runtime.Type == constant.RuntimePHP) {
			if err := createIndexFile(website, runtime); err != nil {
				return err
			}
		}
		if website.Type == constant.Proxy {
			if err := createProxyFile(website); err != nil {
				return err
			}
		}
	}
	return nil
}

func configDefaultNginx(website *model.Website, domains []model.WebsiteDomain, appInstall *model.AppInstall, runtime *model.Runtime) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	if err = createWebsiteFolder(website, runtime); err != nil {
		return err
	}
	configPath := GetSitePath(*website, SiteConf)
	nginxContent := string(nginx_conf.WebsiteDefault)
	config, err := parser.NewStringParser(nginxContent).Parse()
	if err != nil {
		return err
	}
	servers := config.FindServers()
	if len(servers) == 0 {
		return errors.New("nginx config is not valid")
	}
	server := servers[0]
	server.DeleteListen("80")
	var serverNames []string
	for _, domain := range domains {
		serverNames = append(serverNames, domain.Domain)
		setListen(server, strconv.Itoa(domain.Port), website.IPV6, false, website.DefaultServer, false)
	}
	server.UpdateServerName(serverNames)

	siteFolder := path.Join("/www", "sites", website.Alias)
	server.UpdateDirective("access_log", []string{path.Join(siteFolder, "log", "access.log"), "main"})
	server.UpdateDirective("error_log", []string{path.Join(siteFolder, "log", "error.log")})

	rootIndex := path.Join("/www/sites", website.Alias, "index")
	switch website.Type {
	case constant.Deployment:
		proxy := fmt.Sprintf("http://127.0.0.1:%d", appInstall.HttpPort)
		server.UpdateRootProxy([]string{proxy})
	case constant.Static:
		server.UpdateRoot(rootIndex)
		server.UpdateDirective("error_page", []string{"404", "/404.html"})
	case constant.Proxy:
		nginxInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
		server.UpdateDirective("include", []string{nginxInclude})
	case constant.Runtime:
		switch runtime.Type {
		case constant.RuntimePHP:
			server.UpdateDirective("error_page", []string{"404", "/404.html"})
			if runtime.Resource == constant.ResourceLocal {
				server.UpdateRoot(rootIndex)
				localPath := path.Join(GetSitePath(*website, SiteIndexDir), "index.php")
				server.UpdatePHPProxy([]string{website.Proxy}, localPath)
			} else {
				server.UpdateRoot(rootIndex)
				server.UpdatePHPProxy([]string{website.Proxy}, "")
			}
		case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
			server.UpdateRootProxy([]string{fmt.Sprintf("http://%s", website.Proxy)})
		}
	case constant.Subsite:
		parentWebsite, err := websiteRepo.GetFirst(repo.WithByID(website.ParentWebsiteID))
		if err != nil {
			return err
		}
		website.Proxy = parentWebsite.Proxy
		rootIndex = path.Join("/www/sites", parentWebsite.Alias, "index", website.SiteDir)
		server.UpdateDirective("error_page", []string{"404", "/404.html"})
		if parentWebsite.Type == constant.Runtime {
			parentRuntime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(parentWebsite.RuntimeID))
			if err != nil {
				return err
			}
			website.RuntimeID = parentRuntime.ID
			if parentRuntime.Type == constant.RuntimePHP {
				if parentRuntime.Resource == constant.ResourceLocal {
					server.UpdateRoot(rootIndex)
					localPath := path.Join(nginxInstall.GetPath(), rootIndex, "index.php")
					server.UpdatePHPProxy([]string{website.Proxy}, localPath)
				} else {
					server.UpdateRoot(rootIndex)
					server.UpdatePHPProxy([]string{website.Proxy}, "")
				}
			}
		}
		if parentWebsite.Type == constant.Static {
			server.UpdateRoot(rootIndex)
		}
	}

	config.FilePath = configPath
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	if err = opNginx(nginxInstall.ContainerName, constant.NginxCheck); err != nil {
		return err
	}
	if err = opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		return err
	}
	return nil
}

func moveDefaultWafConfig(websiteDir string, defaultConfigContent []byte, defaultRuleDir string, fileOp files.FileOp) error {
	if !fileOp.Stat(websiteDir) {
		if err := fileOp.CreateDir(websiteDir, constant.DirPerm); err != nil {
			return err
		}
	}
	if err := fileOp.SaveFileWithByte(path.Join(websiteDir, "config.json"), defaultConfigContent, constant.DirPerm); err != nil {
		return err
	}
	websiteRuleDir := path.Join(websiteDir, "rules")
	if !fileOp.Stat(websiteRuleDir) {
		if err := fileOp.CreateDir(websiteRuleDir, constant.DirPerm); err != nil {
			return err
		}
	}
	defaultRulesName := []string{"acl", "args", "cookie", "defaultUaBlack", "defaultUrlBlack", "fileExt", "header", "methodWhite", "cdn"}
	for _, ruleName := range defaultRulesName {
		srcPath := path.Join(defaultRuleDir, ruleName+".json")
		if fileOp.Stat(srcPath) {
			_ = fileOp.Copy(srcPath, websiteRuleDir)
		}
	}
	return nil
}

func createAllWebsitesWAFConfig(websites []model.Website) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	wafDataPath := path.Join(nginxInstall.GetPath(), "1pwaf", "data")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(wafDataPath) {
		return nil
	}
	websitesConfigPath := path.Join(wafDataPath, "conf", "sites.json")
	var websitesArray []request.WafWebsite
	for _, website := range websites {
		wafWebsite := request.WafWebsite{
			Key:     website.Alias,
			Domains: make([]string, 0),
			Host:    make([]string, 0),
		}
		websiteDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(website.ID))
		for _, domain := range websiteDomains {
			wafWebsite.Domains = append(wafWebsite.Domains, domain.Domain)
			wafWebsite.Host = append(wafWebsite.Host, domain.Domain+":"+strconv.Itoa(domain.Port))
		}
		websitesArray = append(websitesArray, wafWebsite)
	}
	websitesContent, err := json.Marshal(websitesArray)
	if err != nil {
		return err
	}
	if err := fileOp.SaveFileWithByte(websitesConfigPath, websitesContent, constant.DirPerm); err != nil {
		return err
	}
	var (
		defaultConfigPath = path.Join(wafDataPath, "conf", "siteConfig.json")
		defaultRuleDir    = path.Join(wafDataPath, "rules")
		sitesDir          = path.Join(wafDataPath, "sites")
	)
	defaultConfigContent, err := fileOp.GetContent(defaultConfigPath)
	if err != nil {
		return err
	}

	for _, website := range websites {
		websiteDir := path.Join(sitesDir, website.Alias)
		if err := moveDefaultWafConfig(websiteDir, defaultConfigContent, defaultRuleDir, fileOp); err != nil {
			return err
		}
	}
	return nil
}

func createOpenBasedirConfig(website *model.Website) {
	fileOp := files.NewFileOp()
	userIniPath := path.Join(GetSitePath(*website, SiteIndexDir), ".user.ini")
	_ = fileOp.CreateFile(userIniPath)
	_ = fileOp.SaveFile(userIniPath, fmt.Sprintf("open_basedir=/www/sites/%s/index:/tmp/", website.Alias), 0644)
}

func createWafConfig(website *model.Website, domains []model.WebsiteDomain) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	wafDataPath := path.Join(nginxInstall.GetPath(), "1pwaf", "data")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(wafDataPath) {
		return nil
	}
	websitesConfigPath := path.Join(wafDataPath, "conf", "sites.json")
	content, err := fileOp.GetContent(websitesConfigPath)
	if err != nil {
		return err
	}
	var websitesArray []request.WafWebsite
	if len(content) != 0 {
		if err := json.Unmarshal(content, &websitesArray); err != nil {
			return err
		}
	}
	wafWebsite := request.WafWebsite{
		Key:     website.Alias,
		Domains: make([]string, 0),
		Host:    make([]string, 0),
	}

	for _, domain := range domains {
		wafWebsite.Domains = append(wafWebsite.Domains, domain.Domain)
		wafWebsite.Host = append(wafWebsite.Host, domain.Domain+":"+strconv.Itoa(domain.Port))
	}
	websitesArray = append(websitesArray, wafWebsite)
	websitesContent, err := json.Marshal(websitesArray)
	if err != nil {
		return err
	}
	if err := fileOp.SaveFileWithByte(websitesConfigPath, websitesContent, constant.DirPerm); err != nil {
		return err
	}

	var (
		sitesDir          = path.Join(wafDataPath, "sites")
		defaultConfigPath = path.Join(wafDataPath, "conf", "siteConfig.json")
		defaultRuleDir    = path.Join(wafDataPath, "rules")
		websiteDir        = path.Join(sitesDir, website.Alias)
	)

	defaultConfigContent, err := fileOp.GetContent(defaultConfigPath)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = fileOp.DeleteDir(websiteDir)
		}
	}()

	if err := moveDefaultWafConfig(websiteDir, defaultConfigContent, defaultRuleDir, fileOp); err != nil {
		return err
	}

	if err = opNginx(nginxInstall.ContainerName, constant.NginxCheck); err != nil {
		return err
	}
	if err = opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		return err
	}

	return nil
}

func delNginxConfig(website model.Website, force bool) error {
	configPath := GetSitePath(website, SiteConf)
	fileOp := files.NewFileOp()

	if !fileOp.Stat(configPath) {
		return nil
	}
	if err := fileOp.DeleteFile(configPath); err != nil {
		return err
	}
	sitePath := GteSiteDir(website.Alias)
	if fileOp.Stat(sitePath) {
		xpack.RemoveTamper(website.Alias)
		_ = fileOp.DeleteDir(sitePath)
	}

	nginxApp, err := appRepo.GetFirst(appRepo.WithKey(constant.AppOpenresty))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		if force {
			return nil
		}
		return err
	}
	return nil
}

func delWafConfig(website model.Website, force bool) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	if !common.CompareVersion(nginxInstall.Version, "1.21.4.3-2-0") {
		return nil
	}
	wafDataPath := path.Join(nginxInstall.GetPath(), "1pwaf", "data")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(wafDataPath) {
		return nil
	}
	monitorDir := path.Join(wafDataPath, "db", "sites", website.Alias)
	if fileOp.Stat(monitorDir) {
		_ = fileOp.DeleteDir(monitorDir)
	}
	websitesConfigPath := path.Join(wafDataPath, "conf", "sites.json")
	content, err := fileOp.GetContent(websitesConfigPath)
	if err != nil {
		return err
	}
	var websitesArray []request.WafWebsite
	var newWebsiteArray []request.WafWebsite
	if len(content) > 0 {
		if err = json.Unmarshal(content, &websitesArray); err != nil {
			return err
		}
	}
	for _, wafWebsite := range websitesArray {
		if wafWebsite.Key != website.Alias {
			newWebsiteArray = append(newWebsiteArray, wafWebsite)
		}
	}
	websitesContent, err := json.Marshal(newWebsiteArray)
	if err != nil {
		return err
	}
	if err := fileOp.SaveFileWithByte(websitesConfigPath, websitesContent, constant.DirPerm); err != nil {
		return err
	}

	_ = fileOp.DeleteDir(path.Join(wafDataPath, "sites", website.Alias))

	if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		if force {
			return nil
		}
		return err
	}
	return nil
}

func isHttp3(server *components.Server) bool {
	for _, listen := range server.Listens {
		for _, param := range listen.Parameters {
			if param == "quic" {
				return true
			}
		}
	}
	return false
}

func addListenAndServerName(website model.Website, domains []model.WebsiteDomain) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	http3 := isHttp3(server)

	var allDomains []string
	existDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(website.ID))
	for _, domain := range existDomains {
		allDomains = append(allDomains, domain.Domain)
	}

	for _, domain := range domains {
		setListen(server, strconv.Itoa(domain.Port), website.IPV6, http3, website.DefaultServer, website.Protocol == constant.ProtocolHTTPS && domain.SSL)
		allDomains = append(allDomains, domain.Domain)
	}
	server.UpdateServerName(allDomains)

	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxFull.Install.ContainerName)
}

func deleteListenAndServerName(website model.Website, binds []string, domains []string) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	for _, bind := range binds {
		server.DeleteListen(bind)
		server.DeleteListen("[::]:" + bind)
	}
	for _, domain := range domains {
		server.DeleteServerName(domain)
	}

	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxFull.Install.ContainerName)
}

func setListen(server *components.Server, port string, ipv6, http3, defaultServer, ssl bool) {
	var params []string
	if ssl {
		params = []string{"ssl"}
	}
	server.UpdateListen(port, defaultServer, params...)
	if ssl && http3 {
		server.UpdateListen(port, defaultServer, "quic")
	}
	if !ipv6 {
		return
	}
	server.UpdateListen("[::]:"+port, defaultServer, params...)
	if ssl && http3 {
		server.UpdateListen("[::]:"+port, defaultServer, "quic")
	}
}

func removeSSLListen(website model.Website, binds []string) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	http3 := isHttp3(server)
	for _, bind := range binds {
		server.DeleteListen(bind)
		if website.IPV6 {
			server.DeleteListen("[::]:" + bind)
		}
		setListen(server, bind, website.IPV6, http3, website.DefaultServer, website.Protocol == constant.ProtocolHTTPS)
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxFull.Install.ContainerName)
}

func createPemFile(website model.Website, websiteSSL model.WebsiteSSL) error {
	configDir := GetSitePath(website, SiteSSLDir)
	fileOp := files.NewFileOp()

	if !fileOp.Stat(configDir) {
		if err := fileOp.CreateDir(configDir, constant.DirPerm); err != nil {
			return err
		}
	}

	fullChainFile := path.Join(configDir, "fullchain.pem")
	privatePemFile := path.Join(configDir, "privkey.pem")

	if !fileOp.Stat(fullChainFile) {
		if err := fileOp.CreateFile(fullChainFile); err != nil {
			return err
		}
	}
	if !fileOp.Stat(privatePemFile) {
		if err := fileOp.CreateFile(privatePemFile); err != nil {
			return err
		}
	}

	if err := fileOp.WriteFile(fullChainFile, strings.NewReader(websiteSSL.Pem), constant.DirPerm); err != nil {
		return err
	}
	if err := fileOp.WriteFile(privatePemFile, strings.NewReader(websiteSSL.PrivateKey), constant.DirPerm); err != nil {
		return err
	}
	return nil
}

func getHttpsPort(website *model.Website) ([]int, error) {
	websiteDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(website.ID))
	var httpsPorts []int
	for _, domain := range websiteDomains {
		if domain.SSL {
			httpsPorts = append(httpsPorts, domain.Port)
		}
	}
	if len(httpsPorts) == 0 {
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return nil, err
		}
		httpsPorts = append(httpsPorts, nginxInstall.HttpsPort)
	}
	return httpsPorts, nil
}

func applySSL(website *model.Website, websiteSSL model.WebsiteSSL, req request.WebsiteHTTPSOp) error {
	nginxFull, err := getNginxFull(website)
	if err != nil {
		return nil
	}
	domains, err := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(website.ID))
	if err != nil {
		return nil
	}
	noDefaultPort := true
	httpPorts := make(map[int]struct{})
	for _, domain := range domains {
		if domain.Port == 80 {
			noDefaultPort = false
		}
		if domain.Port != 80 && !domain.SSL {
			httpPorts[domain.Port] = struct{}{}
		}
	}
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]

	httpPort := strconv.Itoa(nginxFull.Install.HttpPort)
	httpsPort, err := getHttpsPort(website)
	if err != nil {
		return err
	}
	httpPortIPV6 := "[::]:" + httpPort

	for _, port := range httpsPort {
		if _, ok := httpPorts[port]; !ok {
			server.DeleteListen(strconv.Itoa(port))
		}
		setListen(server, strconv.Itoa(port), website.IPV6, req.Http3, website.DefaultServer, true)
	}

	server.UpdateDirective("http2", []string{"on"})

	switch req.HttpConfig {
	case constant.HTTPSOnly:
		server.RemoveListenByBind(httpPort)
		server.RemoveListenByBind(httpPortIPV6)
		server.RemoveDirective("if", []string{"($scheme"})
	case constant.HTTPToHTTPS:
		if !noDefaultPort {
			server.UpdateListen(httpPort, website.DefaultServer)
		}
		if website.IPV6 {
			server.UpdateListen(httpPortIPV6, website.DefaultServer)
		}
		server.AddHTTP2HTTPS()
	case constant.HTTPAlso:
		if !noDefaultPort {
			server.UpdateListen(httpPort, website.DefaultServer)
		}
		server.RemoveDirective("if", []string{"($scheme"})
		if website.IPV6 {
			server.UpdateListen(httpPortIPV6, website.DefaultServer)
		}
	}

	if !req.Hsts {
		server.RemoveDirective("add_header", []string{"Strict-Transport-Security", "\"max-age=31536000\""})
	}
	if !req.Http3 {
		for _, port := range httpsPort {
			server.RemoveListen(strconv.Itoa(port), "quic")
			if website.IPV6 {
				httpsPortIPV6 := "[::]:" + strconv.Itoa(port)
				server.RemoveListen(httpsPortIPV6, "quic")
			}
		}
		server.RemoveDirective("add_header", []string{"Alt-Svc"})
	}

	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	if err = createPemFile(*website, websiteSSL); err != nil {
		return err
	}
	nginxParams := getNginxParamsFromStaticFile(dto.SSL, []dto.NginxParam{})
	for i, param := range nginxParams {
		if param.Name == "ssl_certificate" {
			nginxParams[i].Params = []string{path.Join("/www", "sites", website.Alias, "ssl", "fullchain.pem")}
		}
		if param.Name == "ssl_certificate_key" {
			nginxParams[i].Params = []string{path.Join("/www", "sites", website.Alias, "ssl", "privkey.pem")}
		}
		if param.Name == "ssl_protocols" {
			nginxParams[i].Params = req.SSLProtocol
			if len(req.SSLProtocol) == 0 {
				nginxParams[i].Params = []string{"TLSv1.3", "TLSv1.2"}
			}
		}
		if param.Name == "ssl_ciphers" {
			nginxParams[i].Params = []string{req.Algorithm}
			if len(req.Algorithm) == 0 {
				nginxParams[i].Params = []string{"ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:!aNULL:!eNULL:!EXPORT:!DSS:!DES:!RC4:!3DES:!MD5:!PSK:!KRB5:!SRP:!CAMELLIA:!SEED"}
			}
		}
	}
	if req.Hsts {
		nginxParams = append(nginxParams, dto.NginxParam{
			Name:   "add_header",
			Params: []string{"Strict-Transport-Security", "\"max-age=31536000\""},
		})
	}
	if req.Http3 {
		nginxParams = append(nginxParams, dto.NginxParam{
			Name:   "add_header",
			Params: []string{"Alt-Svc", "'h3=\":443\"; ma=2592000'"},
		})
	}

	if err := updateNginxConfig(constant.NginxScopeServer, nginxParams, website); err != nil {
		return err
	}
	return nil
}

func getParamArray(key string, param interface{}) []string {
	var res []string
	switch p := param.(type) {
	case string:
		if key == "index" {
			res = strings.Split(p, "\n")
			return res
		}

		res = strings.Split(p, " ")
		return res
	}
	return res
}

func handleParamMap(paramMap map[string]string, keys []string) []dto.NginxParam {
	var nginxParams []dto.NginxParam
	for k, v := range paramMap {
		for _, name := range keys {
			if name == k {
				param := dto.NginxParam{
					Name:   k,
					Params: getParamArray(k, v),
				}
				nginxParams = append(nginxParams, param)
			}
		}
	}
	return nginxParams
}

func getNginxParams(params interface{}, keys []string) []dto.NginxParam {
	var nginxParams []dto.NginxParam

	switch p := params.(type) {
	case map[string]interface{}:
		return handleParamMap(toMapStr(p), keys)
	case []interface{}:
		for _, mA := range p {
			if m, ok := mA.(map[string]interface{}); ok {
				nginxParams = append(nginxParams, handleParamMap(toMapStr(m), keys)...)
			}
		}
	}
	return nginxParams
}

func toMapStr(m map[string]interface{}) map[string]string {
	ret := make(map[string]string, len(m))
	for k, v := range m {
		ret[k] = fmt.Sprint(v)
	}
	return ret
}

func deleteWebsiteFolder(website *model.Website) error {
	siteFolder := GetSitePath(*website, SiteDir)
	fileOp := files.NewFileOp()
	if fileOp.Stat(siteFolder) {
		_ = fileOp.DeleteDir(siteFolder)
	}
	nginxFilePath := GetSitePath(*website, SiteConf)
	if fileOp.Stat(nginxFilePath) {
		_ = fileOp.DeleteFile(nginxFilePath)
	}
	return nil
}

func opWebsite(website *model.Website, operate string) error {
	nginxInstall, err := getNginxFull(website)
	if err != nil {
		return err
	}
	config := nginxInstall.SiteConfig.Config
	servers := config.FindServers()
	if len(servers) == 0 {
		return errors.New("nginx config is not valid")
	}
	server := servers[0]
	if operate == constant.StopWeb {
		proxyInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
		server.RemoveDirective("include", []string{proxyInclude})
		rewriteInclude := fmt.Sprintf("/www/sites/%s/rewrite/%s.conf", website.Alias, website.Alias)
		server.RemoveDirective("include", []string{rewriteInclude})

		switch website.Type {
		case constant.Deployment:
			server.RemoveDirective("location", []string{"/"})
		case constant.Runtime:
			runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
			if err != nil {
				return err
			}
			if runtime.Type == constant.RuntimePHP {
				server.RemoveDirective("location", []string{"~", "[^/]\\.php(/|$)"})
			} else {
				server.RemoveDirective("location", []string{"/"})
			}
		}
		server.UpdateRoot("/usr/share/nginx/html/stop")
		website.Status = constant.WebStopped
	}
	if operate == constant.StartWeb {
		absoluteIncludeDir := GetSitePath(*website, SiteProxyDir)
		fileOp := files.NewFileOp()
		if fileOp.Stat(absoluteIncludeDir) && !files.IsEmptyDir(absoluteIncludeDir) {
			proxyInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
			server.UpdateDirective("include", []string{proxyInclude})
		}
		rewriteInclude := fmt.Sprintf("/www/sites/%s/rewrite/%s.conf", website.Alias, website.Alias)
		absoluteRewritePath := GetSitePath(*website, SiteReWritePath)
		if fileOp.Stat(absoluteRewritePath) {
			server.UpdateDirective("include", []string{rewriteInclude})
		}
		rootIndex := path.Join("/www/sites", website.Alias, "index")
		if website.SiteDir != "/" {
			rootIndex = path.Join(rootIndex, website.SiteDir)
		}
		switch website.Type {
		case constant.Deployment:
			server.RemoveDirective("root", nil)
			appInstall, err := appInstallRepo.GetFirst(repo.WithByID(website.AppInstallID))
			if err != nil {
				return err
			}
			proxy := fmt.Sprintf("http://127.0.0.1:%d", appInstall.HttpPort)
			server.UpdateRootProxy([]string{proxy})
		case constant.Static:
			server.UpdateRoot(rootIndex)
			server.UpdateRootLocation()
		case constant.Proxy:
			server.RemoveDirective("root", nil)
		case constant.Runtime:
			server.UpdateRoot(rootIndex)
			localPath := ""
			runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
			if err != nil {
				return err
			}
			if runtime.Type == constant.RuntimePHP {
				if website.ProxyType == constant.RuntimeProxyUnix || website.ProxyType == constant.RuntimeProxyTcp {
					localPath = path.Join(GetSitePath(*website, SiteIndexDir), "index.php")
				}
				server.UpdatePHPProxy([]string{website.Proxy}, localPath)
			} else {
				proxy := fmt.Sprintf("http://127.0.0.1:%s", runtime.Port)
				server.UpdateRootProxy([]string{proxy})
			}
		}
		website.Status = constant.WebRunning
		now := time.Now()
		if website.ExpireDate.Before(now) {
			defaultDate, _ := time.Parse(constant.DateLayout, constant.WebsiteDefaultExpireDate)
			website.ExpireDate = defaultDate
		}
	}

	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxInstall.SiteConfig.OldContent, config.FilePath, nginxInstall.Install.ContainerName)
}

func changeIPV6(website model.Website, enable bool) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	listens := server.Listens
	if enable {
		for _, listen := range listens {
			if strings.HasPrefix(listen.Bind, "[::]:") {
				continue
			}
			exist := false
			ipv6Bind := fmt.Sprintf("[::]:%s", listen.Bind)
			for _, li := range listens {
				if li.Bind == ipv6Bind {
					exist = true
					break
				}
			}
			if !exist {
				server.UpdateListen(ipv6Bind, false, listen.GetParameters()[1:]...)
			}
		}
	} else {
		for _, listen := range listens {
			if strings.HasPrefix(listen.Bind, "[::]:") {
				server.RemoveListenByBind(listen.Bind)
			}
		}
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxFull.SiteConfig.OldContent, config.FilePath, nginxFull.Install.ContainerName)
}

func checkIsLinkApp(website model.Website) bool {
	if website.Type == constant.Deployment {
		return true
	}
	if website.Type == constant.Runtime {
		runtime, _ := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
		return runtime.Resource == constant.ResourceAppstore
	}
	return false
}

func chownRootDir(path string) error {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(1 * time.Second))
	_, err := cmdMgr.RunWithStdoutBashCf(`chown -R 1000:1000 "%s"`, path)
	if err != nil {
		return err
	}
	return nil
}

func getWebsiteDomains(domains []request.WebsiteDomain, defaultPort int, websiteID uint) (domainModels []model.WebsiteDomain, addPorts []int, addDomains []string, err error) {
	var (
		ports     = make(map[int]struct{})
		existPort = make(map[int]struct{})
	)
	existDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(websiteID))
	for _, domain := range existDomains {
		existPort[domain.Port] = struct{}{}
	}
	for _, domain := range domains {
		if domain.Domain == "" {
			continue
		}
		if !common.IsValidDomain(domain.Domain) {
			err = buserr.WithName("ErrDomainFormat", domain.Domain)
			return
		}
		var domainModel model.WebsiteDomain
		domainModel.Domain, err = handleChineseDomain(domain.Domain)
		if err != nil {
			return
		}
		domainModel.Domain = strings.ToLower(domainModel.Domain)
		domainModel.Port = domain.Port
		if domain.Port == 0 {
			domain.Port = defaultPort
		}
		domainModel.SSL = domain.SSL
		domainModel.WebsiteID = websiteID
		domainModels = append(domainModels, domainModel)
		if _, ok := existPort[domainModel.Port]; !ok {
			ports[domainModel.Port] = struct{}{}
		}
		if exist, _ := websiteDomainRepo.GetFirst(websiteDomainRepo.WithDomain(domainModel.Domain), websiteDomainRepo.WithWebsiteId(websiteID)); exist.ID == 0 {
			addDomains = append(addDomains, domainModel.Domain)
		}
	}
	for _, domain := range domainModels {
		if exist, _ := websiteDomainRepo.GetFirst(websiteDomainRepo.WithDomain(domain.Domain), websiteDomainRepo.WithPort(domain.Port)); exist.ID > 0 {
			website, _ := websiteRepo.GetFirst(repo.WithByID(exist.WebsiteID))
			err = buserr.WithName("ErrDomainIsUsed", website.PrimaryDomain)
			return
		}
	}

	for port := range ports {
		if port == defaultPort {
			addPorts = append(addPorts, port)
			continue
		}
		if existPorts, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithPort(port)); len(existPorts) == 0 {
			errMap := make(map[string]interface{})
			errMap["port"] = port
			appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithPort(port))
			if appInstall.ID > 0 {
				errMap["type"] = i18n.GetMsgByKey("TYPE_APP")
				errMap["name"] = appInstall.Name
				err = buserr.WithMap("ErrPortExist", errMap, nil)
				return
			}
			runtime, _ := runtimeRepo.GetFirst(context.Background(), runtimeRepo.WithPort(port))
			if runtime != nil {
				errMap["type"] = i18n.GetMsgByKey("TYPE_RUNTIME")
				errMap["name"] = runtime.Name
				err = buserr.WithMap("ErrPortExist", errMap, nil)
				return
			}
			if common.ScanPort(port) {
				err = buserr.WithDetail("ErrPortInUsed", port, nil)
				return
			}
		}
		if existPorts, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(websiteID), websiteDomainRepo.WithPort(port)); len(existPorts) == 0 {
			addPorts = append(addPorts, port)
		}
	}

	return
}

func saveCertificateFile(websiteSSL *model.WebsiteSSL, logger *log.Logger) {
	if websiteSSL.PushDir {
		fileOp := files.NewFileOp()
		var (
			pushErr error
			MsgMap  = map[string]interface{}{"path": websiteSSL.Dir, "status": i18n.GetMsgByKey("Success")}
		)
		if pushErr = fileOp.SaveFile(path.Join(websiteSSL.Dir, "privkey.pem"), websiteSSL.PrivateKey, constant.FilePerm); pushErr != nil {
			MsgMap["status"] = i18n.GetMsgByKey("Failed")
			logger.Println(i18n.GetMsgWithMap("PushDirLog", MsgMap))
			logger.Println("Push dir failed:" + pushErr.Error())
		}
		if pushErr = fileOp.SaveFile(path.Join(websiteSSL.Dir, "fullchain.pem"), websiteSSL.Pem, constant.FilePerm); pushErr != nil {
			MsgMap["status"] = i18n.GetMsgByKey("Failed")
			logger.Println(i18n.GetMsgWithMap("PushDirLog", MsgMap))
			logger.Println("Push dir failed:" + pushErr.Error())
		}
		if pushErr == nil {
			logger.Println(i18n.GetMsgWithMap("PushDirLog", MsgMap))
		}
	}
}

func GetSystemSSL() (bool, uint) {
	sslSetting, err := settingRepo.Get(settingRepo.WithByKey("SSL"))
	if err != nil {
		global.LOG.Errorf("load service ssl from setting failed, err: %v", err)
		return false, 0
	}
	if sslSetting.Value == "enable" {
		sslID, _ := settingRepo.Get(settingRepo.WithByKey("SSLID"))
		idValue, _ := strconv.Atoi(sslID.Value)
		if idValue > 0 {
			return true, uint(idValue)
		}
	}
	return false, 0
}

func UpdateSSLConfig(websiteSSL model.WebsiteSSL) error {
	websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(websiteSSL.ID))
	if len(websites) > 0 {
		for _, website := range websites {
			if err := createPemFile(website, websiteSSL); err != nil {
				return buserr.WithMap("ErrUpdateWebsiteSSL", map[string]interface{}{"name": website.PrimaryDomain, "err": err.Error()}, err)
			}
		}
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return err
		}
		if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
			return buserr.WithErr("ErrSSLApply", err)
		}
	}
	enable, sslID := GetSystemSSL()
	if enable && sslID == websiteSSL.ID {
		fileOp := files.NewFileOp()
		secretDir := path.Join(global.Dir.DataDir, "secret")
		if err := fileOp.WriteFile(path.Join(secretDir, "server.crt"), strings.NewReader(websiteSSL.Pem), 0600); err != nil {
			global.LOG.Errorf("Failed to update the SSL certificate File for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return err
		}
		if err := fileOp.WriteFile(path.Join(secretDir, "server.key"), strings.NewReader(websiteSSL.PrivateKey), 0600); err != nil {
			global.LOG.Errorf("Failed to update the SSL certificate for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return err
		}
	}
	return nil
}

func ChangeHSTSConfig(enable bool, website model.Website) error {
	includeDir := GetSitePath(website, SiteProxyDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		return nil
	}
	err := filepath.Walk(includeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if filepath.Ext(path) == ".conf" {
				par, err := parser.NewParser(path)
				if err != nil {
					return err
				}
				config, err := par.Parse()
				if err != nil {
					return err
				}
				config.FilePath = path
				directives := config.Directives
				location, ok := directives[0].(*components.Location)
				if !ok {
					return nil
				}
				if enable {
					location.UpdateDirective("add_header", []string{"Strict-Transport-Security", "\"max-age=31536000\""})
				} else {
					location.RemoveDirective("add_header", []string{"Strict-Transport-Security", "\"max-age=31536000\""})
				}
				if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
					return buserr.WithErr("ErrUpdateBuWebsite", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func checkSSLStatus(expireDate time.Time) string {
	now := time.Now()
	daysUntilExpiry := int(expireDate.Sub(now).Hours() / 24)

	if daysUntilExpiry < 0 {
		return "danger"
	} else if daysUntilExpiry <= 10 {
		return "warning"
	}
	return "success"
}

func getResourceContent(fileOp files.FileOp, resourcePath string) (string, error) {
	if fileOp.Stat(resourcePath) {
		content, err := fileOp.GetContent(resourcePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	return "", nil
}

func GetWebSiteRootDir() string {
	siteSetting, _ := settingRepo.Get(settingRepo.WithByKey("WEBSITE_DIR"))
	dir := siteSetting.Value
	if dir == "" {
		dir = path.Join(global.Dir.DataDir, "www")
	}
	return dir
}

func GteSiteDir(alias string) string {
	return path.Join(GetWebSiteRootDir(), "sites", alias)
}

const (
	SiteConf              = "SiteConf"
	SiteAccessLog         = "access.log"
	SiteErrorLog          = "error.log"
	WebsiteRootDir        = "WebsiteRootDir"
	SiteDir               = "SiteDir"
	SiteIndexDir          = "SiteIndexDir"
	SiteProxyDir          = "SiteProxyDir"
	SiteSSLDir            = "SiteSSLDir"
	SiteReWritePath       = "SiteReWritePath"
	SiteRedirectDir       = "SiteRedirectDir"
	SiteCacheDir          = "SiteCacheDir"
	SiteConfDir           = "SiteConfDir"
	SitesRootDir          = "SitesRootDir"
	DefaultDir            = "DefaultDir"
	DefaultRewriteDir     = "DefaultRewriteDir"
	SiteRootAuthBasicPath = "SiteRootAuthBasicPath"
	SitePathAuthBasicDir  = "SitePathAuthBasicDir"
	SiteUpstreamDir       = "SiteUpstreamDir"
)

func GetSitePath(website model.Website, confType string) string {
	switch confType {
	case SiteConf:
		return path.Join(GetWebSiteRootDir(), "conf.d", website.Alias+".conf")
	case SiteAccessLog:
		return path.Join(GteSiteDir(website.Alias), "log", "access.log")
	case SiteErrorLog:
		return path.Join(GteSiteDir(website.Alias), "log", "error.log")
	case SiteDir:
		return GteSiteDir(website.Alias)
	case SiteIndexDir:
		return path.Join(GteSiteDir(website.Alias), "index")
	case SiteCacheDir:
		return path.Join(GteSiteDir(website.Alias), "cache")
	case SiteProxyDir:
		return path.Join(GteSiteDir(website.Alias), "proxy")
	case SiteSSLDir:
		return path.Join(GteSiteDir(website.Alias), "ssl")
	case SiteReWritePath:
		return path.Join(GteSiteDir(website.Alias), "rewrite", website.Alias+".conf")
	case SiteRedirectDir:
		return path.Join(GteSiteDir(website.Alias), "redirect")
	case SiteRootAuthBasicPath:
		return path.Join(GteSiteDir(website.Alias), "auth_basic", "auth.pass")
	case SitePathAuthBasicDir:
		return path.Join(GteSiteDir(website.Alias), "path_auth")
	case SiteUpstreamDir:
		return path.Join(GteSiteDir(website.Alias), "upstream")
	}
	return ""
}

func GetOpenrestyDir(confType string) string {
	switch confType {
	case WebsiteRootDir:
		return GetWebSiteRootDir()
	case SiteConfDir:
		return path.Join(GetWebSiteRootDir(), "conf.d")
	case SitesRootDir:
		return path.Join(GetWebSiteRootDir(), "sites")
	case DefaultDir:
		return path.Join(GetWebSiteRootDir(), "default")
	case DefaultRewriteDir:
		return path.Join(GetWebSiteRootDir(), "default", "rewrite")
	}
	return ""
}

func openProxyCache(website model.Website) error {
	cacheDir := GetSitePath(website, SiteCacheDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(cacheDir) {
		_ = fileOp.CreateDir(cacheDir, constant.DirPerm)
	}
	content, err := fileOp.GetContent(GetSitePath(website, SiteConf))
	if err != nil {
		return err
	}
	if strings.Contains(string(content), "proxy_cache_path") {
		return nil
	}
	proxyCachePath := fmt.Sprintf("/www/sites/%s/cache levels=1:2 keys_zone=proxy_cache_zone_of_%s:5m max_size=1g inactive=24h", website.Alias, website.Alias)
	return updateNginxConfig("", []dto.NginxParam{{Name: "proxy_cache_path", Params: []string{proxyCachePath}}}, &website)
}

func ConfigAllowIPs(ips []string, website model.Website) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return err
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	server.RemoveDirective("allow", nil)
	server.RemoveDirective("deny", nil)
	if len(ips) > 0 {
		server.UpdateAllowIPs(ips)
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, config.FilePath, nginxFull.Install.ContainerName)
}

func GetAllowIps(website model.Website) []string {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	dirs := server.GetDirectives()
	var ips []string
	for _, dir := range dirs {
		if dir.GetName() == "allow" {
			ips = append(ips, dir.GetParameters()...)
		}
	}
	return ips
}

func ConfigAIProxy(website model.Website) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	dirs := server.GetDirectives()
	for _, dir := range dirs {
		if dir.GetName() == "location" && dir.GetParameters()[0] == "/" {
			server.UpdateRootProxyForAi([]string{fmt.Sprintf("http://%s", website.Proxy)})
		}
	}
	return nil
}

func handleDefaultOwn(dir string) {
	parentDir := path.Dir(dir)
	info, err := os.Stat(parentDir)
	if err != nil {
		return
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	uid, gid := -1, -1
	if ok {
		uid, gid = int(stat.Uid), int(stat.Gid)
	}
	_ = os.Chown(dir, uid, gid)
}

func getSystemProxy(useProxy bool) *dto.SystemProxy {
	if !useProxy {
		return nil
	}
	settingService := NewISettingService()
	systemProxy, _ := settingService.GetSystemProxy()
	return systemProxy
}

func hasHttp3(params []string) bool {
	for _, param := range params {
		if param == "quic" {
			return true
		}
	}
	return false
}

func hasDefaultServer(params []string) bool {
	for _, param := range params {
		if param == "default_server" {
			return true
		}
	}
	return false
}
