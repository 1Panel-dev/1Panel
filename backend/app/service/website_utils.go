package service

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"gopkg.in/yaml.v3"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/cmd/server/nginx_conf"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func getDomain(domainStr string, defaultPort int) (model.WebsiteDomain, error) {
	var (
		err    error
		domain = model.WebsiteDomain{}
		portN  int
	)
	domainArray := strings.Split(domainStr, ":")
	if len(domainArray) == 1 {
		domain.Domain, err = handleChineseDomain(domainArray[0])
		if err != nil {
			return domain, err
		}
		domain.Port = defaultPort
		return domain, nil
	}
	if len(domainArray) > 1 {
		domain.Domain, err = handleChineseDomain(domainArray[0])
		if err != nil {
			return domain, err
		}
		portStr := domainArray[1]
		portN, err = strconv.Atoi(portStr)
		if err != nil {
			return domain, buserr.WithName("ErrTypePort", portStr)
		}
		if portN <= 0 || portN > 65535 {
			return domain, buserr.New("ErrTypePortRange")
		}
		domain.Port = portN
		return domain, nil
	}
	return domain, nil
}

func handleChineseDomain(domain string) (string, error) {
	if common.ContainsChinese(domain) {
		return common.PunycodeEncode(domain)
	}
	return domain, nil
}

func createIndexFile(website *model.Website, runtime *model.Runtime) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}

	indexFolder := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "www", "sites", website.Alias, "index")
	indexPath := ""
	indexContent := ""
	switch website.Type {
	case constant.Static:
		indexPath = path.Join(indexFolder, "index.html")
		indexContent = string(nginx_conf.Index)
	case constant.Runtime:
		if runtime.Type == constant.RuntimePHP {
			indexPath = path.Join(indexFolder, "index.php")
			indexContent = string(nginx_conf.IndexPHP)
		}
	}

	fileOp := files.NewFileOp()
	if !fileOp.Stat(indexFolder) {
		if err := fileOp.CreateDir(indexFolder, 0755); err != nil {
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
	if err := fileOp.WriteFile(indexPath, strings.NewReader(indexContent), 0755); err != nil {
		return err
	}
	return nil
}

func createProxyFile(website *model.Website) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	proxyFolder := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "www", "sites", website.Alias, "proxy")
	filePath := path.Join(proxyFolder, "root.conf")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(proxyFolder) {
		if err := fileOp.CreateDir(proxyFolder, 0755); err != nil {
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
		return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
	}
	return nil
}

func createWebsiteFolder(nginxInstall model.AppInstall, website *model.Website, runtime *model.Runtime) error {
	nginxFolder := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name)
	siteFolder := path.Join(nginxFolder, "www", "sites", website.Alias)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(siteFolder) {
		if err := fileOp.CreateDir(siteFolder, 0755); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "log"), 0755); err != nil {
			return err
		}
		if err := fileOp.CreateFile(path.Join(siteFolder, "log", "access.log")); err != nil {
			return err
		}
		if err := fileOp.CreateFile(path.Join(siteFolder, "log", "error.log")); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "index"), 0775); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "ssl"), 0755); err != nil {
			return err
		}
		if website.Type == constant.Runtime {
			if runtime.Type == constant.RuntimePHP && runtime.Resource == constant.ResourceLocal {
				phpPoolDir := path.Join(siteFolder, "php-pool")
				if err := fileOp.CreateDir(phpPoolDir, 0755); err != nil {
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
	return fileOp.CopyDir(path.Join(nginxFolder, "www", "common", "waf", "rules"), path.Join(siteFolder, "waf"))
}

func configDefaultNginx(website *model.Website, domains []model.WebsiteDomain, appInstall *model.AppInstall, runtime *model.Runtime) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	if err := createWebsiteFolder(nginxInstall, website, runtime); err != nil {
		return err
	}

	nginxFileName := website.Alias + ".conf"
	configPath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "conf", "conf.d", nginxFileName)
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
		server.UpdateListen(strconv.Itoa(domain.Port), false)
		if website.IPV6 {
			server.UpdateListen("[::]:"+strconv.Itoa(domain.Port), false)
		}
	}
	server.UpdateServerName(serverNames)

	siteFolder := path.Join("/www", "sites", website.Alias)
	commonFolder := path.Join("/www", "common")
	server.UpdateDirective("access_log", []string{path.Join(siteFolder, "log", "access.log")})
	server.UpdateDirective("error_log", []string{path.Join(siteFolder, "log", "error.log")})
	server.UpdateDirective("access_by_lua_file", []string{path.Join(commonFolder, "waf", "access.lua")})
	server.UpdateDirective("set", []string{"$RulePath", path.Join(siteFolder, "waf", "rules")})
	server.UpdateDirective("set", []string{"$logdir", path.Join(siteFolder, "log")})

	rootIndex := path.Join("/www/sites", website.Alias, "index")
	switch website.Type {
	case constant.Deployment:
		proxy := fmt.Sprintf("http://127.0.0.1:%d", appInstall.HttpPort)
		server.UpdateRootProxy([]string{proxy})
	case constant.Static:
		server.UpdateRoot(rootIndex)
	case constant.Proxy:
		nginxInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
		server.UpdateDirective("include", []string{nginxInclude})
	case constant.Runtime:
		switch runtime.Type {
		case constant.RuntimePHP:
			if runtime.Resource == constant.ResourceLocal {
				switch runtime.Type {
				case constant.RuntimePHP:
					server.UpdateRoot(rootIndex)
					localPath := path.Join(nginxInstall.GetPath(), rootIndex, "index.php")
					server.UpdatePHPProxy([]string{website.Proxy}, localPath)
				}
			} else {
				server.UpdateRoot(rootIndex)
				server.UpdatePHPProxy([]string{website.Proxy}, "")
			}
		case constant.RuntimeNode:
			proxy := fmt.Sprintf("http://127.0.0.1:%d", runtime.Port)
			server.UpdateRootProxy([]string{proxy})
		}
	}

	config.FilePath = configPath
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	if err := opNginx(nginxInstall.ContainerName, constant.NginxCheck); err != nil {
		_ = deleteWebsiteFolder(nginxInstall, website)
		return err
	}
	if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		_ = deleteWebsiteFolder(nginxInstall, website)
		return err
	}
	return nil
}

func delNginxConfig(website model.Website, force bool) error {
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

	nginxFileName := website.Alias + ".conf"
	configPath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "conf", "conf.d", nginxFileName)
	fileOp := files.NewFileOp()

	if !fileOp.Stat(configPath) {
		return nil
	}
	if err := fileOp.DeleteFile(configPath); err != nil {
		return err
	}
	sitePath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "www", "sites", website.Alias)
	if fileOp.Stat(sitePath) {
		_ = fileOp.DeleteDir(sitePath)
	}

	if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		if force {
			return nil
		}
		return err
	}
	return nil
}

func addListenAndServerName(website model.Website, ports []int, domains []string) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	for _, port := range ports {
		server.AddListen(strconv.Itoa(port), false)
		if website.IPV6 {
			server.UpdateListen("[::]:"+strconv.Itoa(port), false)
		}
	}
	for _, domain := range domains {
		server.AddServerName(domain)
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
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

func createPemFile(website model.Website, websiteSSL model.WebsiteSSL) error {
	nginxApp, err := appRepo.GetFirst(appRepo.WithKey(constant.AppOpenresty))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		return err
	}

	configDir := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "www", "sites", website.Alias, "ssl")
	fileOp := files.NewFileOp()

	if !fileOp.Stat(configDir) {
		if err := fileOp.CreateDir(configDir, 0775); err != nil {
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

	if err := fileOp.WriteFile(fullChainFile, strings.NewReader(websiteSSL.Pem), 0644); err != nil {
		return err
	}
	if err := fileOp.WriteFile(privatePemFile, strings.NewReader(websiteSSL.PrivateKey), 0644); err != nil {
		return err
	}
	return nil
}

func applySSL(website model.Website, websiteSSL model.WebsiteSSL, req request.WebsiteHTTPSOp) error {
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]

	httpPort := strconv.Itoa(nginxFull.Install.HttpPort)
	httpsPort := strconv.Itoa(nginxFull.Install.HttpsPort)
	httpPortIPV6 := "[::]:" + httpPort
	httpsPortIPV6 := "[::]:" + httpsPort

	server.UpdateListen(httpsPort, website.DefaultServer, "ssl", "http2")
	if website.IPV6 {
		server.UpdateListen(httpsPortIPV6, website.DefaultServer, "ssl", "http2")
	}

	switch req.HttpConfig {
	case constant.HTTPSOnly:
		server.RemoveListenByBind(httpPort)
		server.RemoveListenByBind(httpPortIPV6)
		server.RemoveDirective("if", []string{"($scheme"})
	case constant.HTTPToHTTPS:
		server.UpdateListen(httpPort, website.DefaultServer)
		if website.IPV6 {
			server.UpdateListen(httpPortIPV6, website.DefaultServer)
		}
		server.AddHTTP2HTTPS()
	case constant.HTTPAlso:
		server.UpdateListen(httpPort, website.DefaultServer)
		server.RemoveDirective("if", []string{"($scheme"})
		if website.IPV6 {
			server.UpdateListen(httpPortIPV6, website.DefaultServer)
		}
	}

	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	if err := createPemFile(website, websiteSSL); err != nil {
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
		}
		if param.Name == "ssl_ciphers" {
			nginxParams[i].Params = []string{req.Algorithm}
		}
	}
	if err := updateNginxConfig(constant.NginxScopeServer, nginxParams, &website); err != nil {
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

func deleteWebsiteFolder(nginxInstall model.AppInstall, website *model.Website) error {
	nginxFolder := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name)
	siteFolder := path.Join(nginxFolder, "www", "sites", website.Alias)
	fileOp := files.NewFileOp()
	if fileOp.Stat(siteFolder) {
		_ = fileOp.DeleteDir(siteFolder)
	}
	nginxFilePath := path.Join(nginxFolder, "conf", "conf.d", website.PrimaryDomain+".conf")
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
			runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
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
		proxyInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
		absoluteIncludeDir := path.Join(nginxInstall.Install.GetPath(), fmt.Sprintf("/www/sites/%s/proxy", website.Alias))
		if files.NewFileOp().Stat(absoluteIncludeDir) {
			server.UpdateDirective("include", []string{proxyInclude})
		}
		server.UpdateDirective("include", []string{proxyInclude})
		rewriteInclude := fmt.Sprintf("/www/sites/%s/rewrite/%s.conf", website.Alias, website.Alias)
		absoluteRewritePath := path.Join(nginxInstall.Install.GetPath(), rewriteInclude)
		if files.NewFileOp().Stat(absoluteRewritePath) {
			server.UpdateDirective("include", []string{rewriteInclude})
		}
		rootIndex := path.Join("/www/sites", website.Alias, "index")
		if website.SiteDir != "/" {
			rootIndex = path.Join(rootIndex, website.SiteDir)
		}
		switch website.Type {
		case constant.Deployment:
			server.RemoveDirective("root", nil)
			appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
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
			runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
			if err != nil {
				return err
			}
			if runtime.Type == constant.RuntimePHP {
				if website.ProxyType == constant.RuntimeProxyUnix {
					localPath = path.Join(nginxInstall.Install.GetPath(), rootIndex, "index.php")
				}
				server.UpdatePHPProxy([]string{website.Proxy}, localPath)
			} else {
				proxy := fmt.Sprintf("http://127.0.0.1:%d", runtime.Port)
				server.UpdateRootProxy([]string{proxy})
			}
		}
		website.Status = constant.WebRunning
		now := time.Now()
		if website.ExpireDate.Before(now) {
			defaultDate, _ := time.Parse(constant.DateLayout, constant.DefaultDate)
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
		runtime, _ := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
		return runtime.Resource == constant.ResourceAppstore
	}
	return false
}

func chownRootDir(path string) error {
	_, err := cmd.ExecWithTimeOut(fmt.Sprintf(`chown -R 1000:1000 "%s"`, path), 1*time.Second)
	if err != nil {
		return err
	}
	return nil
}

func changeServiceName(newComposeContent, newServiceName string) (composeByte []byte, err error) {
	composeMap := make(map[string]interface{})
	if err = yaml.Unmarshal([]byte(newComposeContent), &composeMap); err != nil {
		return
	}
	value, ok := composeMap["services"]
	if !ok {
		err = buserr.New(constant.ErrFileParse)
		return
	}
	servicesMap := value.(map[string]interface{})

	index := 0
	serviceName := ""
	for k := range servicesMap {
		serviceName = k
		if index > 0 {
			continue
		}
		index++
	}
	if newServiceName != serviceName {
		servicesMap[newServiceName] = servicesMap[serviceName]
		delete(servicesMap, serviceName)
	}

	return yaml.Marshal(composeMap)
}

func getWebsiteDomains(domains string, defaultPort int, websiteID uint) (domainModels []model.WebsiteDomain, addPorts []int, addDomains []string, err error) {
	var (
		ports = make(map[int]struct{})
	)
	domainArray := strings.Split(domains, "\n")
	for _, domain := range domainArray {
		if domain == "" {
			continue
		}
		if !common.IsValidDomain(domain) {
			err = buserr.WithName("ErrDomainFormat", domain)
			return
		}
		var domainModel model.WebsiteDomain
		domainModel, err = getDomain(domain, defaultPort)
		if err != nil {
			return
		}
		if reflect.DeepEqual(domainModel, model.WebsiteDomain{}) {
			continue
		}
		domainModel.WebsiteID = websiteID
		domainModels = append(domainModels, domainModel)
		if domainModel.Port != defaultPort {
			ports[domainModel.Port] = struct{}{}
		}
		if exist, _ := websiteDomainRepo.GetFirst(websiteDomainRepo.WithDomain(domainModel.Domain), websiteDomainRepo.WithWebsiteId(websiteID)); exist.ID == 0 {
			addDomains = append(addDomains, domainModel.Domain)
		}
	}
	for _, domain := range domainModels {
		if exist, _ := websiteDomainRepo.GetFirst(websiteDomainRepo.WithDomain(domain.Domain), websiteDomainRepo.WithPort(domain.Port)); exist.ID > 0 {
			website, _ := websiteRepo.GetFirst(commonRepo.WithByID(exist.WebsiteID))
			err = buserr.WithName(constant.ErrDomainIsUsed, website.PrimaryDomain)
			return
		}
	}

	for port := range ports {
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
			runtime, _ := runtimeRepo.GetFirst(runtimeRepo.WithPort(port))
			if runtime != nil {
				errMap["type"] = i18n.GetMsgByKey("TYPE_RUNTIME")
				errMap["name"] = runtime.Name
				err = buserr.WithMap("ErrPortExist", errMap, nil)
				return
			}
			if common.ScanPort(port) {
				err = buserr.WithDetail(constant.ErrPortInUsed, port, nil)
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
		if pushErr = fileOp.SaveFile(path.Join(websiteSSL.Dir, "privkey.pem"), websiteSSL.PrivateKey, 0666); pushErr != nil {
			MsgMap["status"] = i18n.GetMsgByKey("Failed")
			logger.Println(i18n.GetMsgWithMap("PushDirLog", MsgMap))
			logger.Println("Push dir failed:" + pushErr.Error())
		}
		if pushErr = fileOp.SaveFile(path.Join(websiteSSL.Dir, "fullchain.pem"), websiteSSL.Pem, 0666); pushErr != nil {
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
			return buserr.WithErr(constant.ErrSSLApply, err)
		}
	}
	enable, sslID := GetSystemSSL()
	if enable && sslID == websiteSSL.ID {
		fileOp := files.NewFileOp()
		secretDir := path.Join(global.CONF.System.BaseDir, "1panel/secret")
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

func ChangeHSTSConfig(enable bool, nginxInstall model.AppInstall, website model.Website) error {
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "proxy")
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
					return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
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
