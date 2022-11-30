package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/cmd/server/nginx_conf"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"path"
	"strconv"
	"strings"
)

func getDomain(domainStr string, websiteID uint) (model.WebSiteDomain, error) {
	domain := model.WebSiteDomain{
		WebSiteID: websiteID,
	}
	domainArray := strings.Split(domainStr, ":")
	if len(domainArray) == 1 {
		domain.Domain = domainArray[0]
		domain.Port = 80
		return domain, nil
	}
	if len(domainArray) > 1 {
		domain.Domain = domainArray[0]
		portStr := domainArray[1]
		portN, err := strconv.Atoi(portStr)
		if err != nil {
			return model.WebSiteDomain{}, err
		}
		domain.Port = portN
		return domain, nil
	}
	return model.WebSiteDomain{}, nil
}

func createStaticHtml(website *model.WebSite) error {
	nginxInstall, err := getAppInstallByKey("nginx")
	if err != nil {
		return err
	}

	indexFolder := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "www", "sites", website.Alias)
	indexPath := path.Join(indexFolder, "index.html")
	indexContent := string(nginx_conf.Index)
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
	if err := fileOp.WriteFile(indexPath, strings.NewReader(indexContent), 0755); err != nil {
		return err
	}
	return nil
}

func createWebsiteFolder(nginxInstall model.AppInstall, website *model.WebSite) error {

	nginxFolder := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name)
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
		if err := fileOp.CreateDir(path.Join(siteFolder, "waf", "rules"), 0755); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "data"), 0755); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "ssl"), 0755); err != nil {
			return err
		}
	}
	return fileOp.CopyDir(path.Join(nginxFolder, "www", "common", "waf", "rules"), path.Join(siteFolder, "waf", "rules"))
}

func configDefaultNginx(website *model.WebSite, domains []model.WebSiteDomain) error {

	nginxInstall, err := getAppInstallByKey("nginx")
	if err != nil {
		return err
	}
	if err := createWebsiteFolder(nginxInstall, website); err != nil {
		return err
	}

	nginxFileName := website.Alias + ".conf"
	configPath := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "conf", "conf.d", nginxFileName)
	nginxContent := string(nginx_conf.WebsiteDefault)
	config := parser.NewStringParser(nginxContent).Parse()
	servers := config.FindServers()
	if len(servers) == 0 {
		return errors.New("nginx config is not valid")
	}
	server := servers[0]
	var serverNames []string
	for _, domain := range domains {
		serverNames = append(serverNames, domain.Domain)
		server.UpdateListen(strconv.Itoa(domain.Port), false)
	}
	server.UpdateServerName(serverNames)

	siteFolder := path.Join("/www", "sites", website.Alias)
	commonFolder := path.Join("/www", "common")
	server.UpdateDirective("access_log", []string{path.Join(siteFolder, "log", "access.log")})
	server.UpdateDirective("access_by_lua_file", []string{path.Join(commonFolder, "waf", "access.lua")})
	server.UpdateDirective("set", []string{"$RulePath", path.Join(siteFolder, "waf", "rules")})
	server.UpdateDirective("set", []string{"$logdir", path.Join(siteFolder, "log")})

	if website.Type == "deployment" {
		appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		proxy := fmt.Sprintf("http://127.0.0.1:%d", appInstall.HttpPort)
		server.UpdateRootProxy([]string{proxy})
	} else {
		server.UpdateRoot(path.Join("/www/sites", website.Alias))
		server.UpdateRootLocation()
	}

	config.FilePath = configPath
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}

	if err := opNginx(nginxInstall.ContainerName, constant.NginxCheck); err != nil {
		return err
	}
	return opNginx(nginxInstall.ContainerName, constant.NginxReload)
}

func delNginxConfig(website model.WebSite) error {

	nginxApp, err := appRepo.GetFirst(appRepo.WithKey("nginx"))
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
	configPath := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "conf", "conf.d", nginxFileName)
	fileOp := files.NewFileOp()

	if !fileOp.Stat(configPath) {
		return nil
	}
	if err := fileOp.DeleteFile(configPath); err != nil {
		return err
	}
	return opNginx(nginxInstall.ContainerName, "reload")
}

func addListenAndServerName(website model.WebSite, ports []int, domains []string) error {

	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	for _, port := range ports {
		server.AddListen(strconv.Itoa(port), false)
	}
	for _, domain := range domains {
		server.AddServerName(domain)
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxFull.Install.ContainerName)
}

func deleteListenAndServerName(website model.WebSite, ports []int, domains []string) error {

	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	for _, port := range ports {
		server.DeleteListen(strconv.Itoa(port))
	}
	for _, domain := range domains {
		server.DeleteServerName(domain)
	}

	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxFull.Install.ContainerName)
}

func getKeysFromStaticFile(scope dto.NginxKey) []string {
	var res []string
	newConfig := &components.Config{}
	switch scope {
	case dto.SSL:
		newConfig = parser.NewStringParser(string(nginx_conf.SSL)).Parse()
	}
	for _, dir := range newConfig.GetDirectives() {
		res = append(res, dir.GetName())
	}
	return res
}

func createPemFile(website model.WebSite, websiteSSL model.WebSiteSSL) error {
	nginxApp, err := appRepo.GetFirst(appRepo.WithKey("nginx"))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		return err
	}

	configDir := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "www", "sites", website.Alias, "ssl")
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

func applySSL(website model.WebSite, websiteSSL model.WebSiteSSL) error {

	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	server.UpdateListen("443", false, "ssl")
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
