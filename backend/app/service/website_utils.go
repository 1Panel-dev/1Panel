package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/cmd/server/nginx_conf"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"os"
	"path"
	"reflect"
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
	nginxApp, err := appRepo.GetFirst(appRepo.WithKey("nginx"))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		return err
	}
	indexFolder := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "www", website.Alias)
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

func configDefaultNginx(website *model.WebSite, domains []model.WebSiteDomain) error {

	nginxApp, err := appRepo.GetFirst(appRepo.WithKey("nginx"))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
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
	if website.Type == "deployment" {
		appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		proxy := fmt.Sprintf("http://127.0.0.1:%d", appInstall.HttpPort)
		server.UpdateRootProxy([]string{proxy})
	} else {
		server.UpdateRoot(path.Join("/www/root", website.Alias))
		server.UpdateRootLocation()
	}

	config.FilePath = configPath
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	if err := opNginx(nginxInstall.ContainerName, "check"); err != nil {
		return err
	}
	return opNginx(nginxInstall.ContainerName, "reload")
}

func opNginx(containerName, operate string) error {
	nginxCmd := fmt.Sprintf("docker exec -i %s %s", containerName, "nginx -s reload")
	if operate == "check" {
		nginxCmd = fmt.Sprintf("docker exec -i %s %s", containerName, "nginx -t")
	}
	if out, err := cmd.Exec(nginxCmd); err != nil {
		return errors.New(out)
	}
	return nil
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

func nginxCheckAndReload(oldContent string, filePath string, containerName string) error {

	if err := opNginx(containerName, "check"); err != nil {
		_ = files.NewFileOp().WriteFile(filePath, strings.NewReader(oldContent), 0644)
		return err
	}

	if err := opNginx(containerName, "reload"); err != nil {
		_ = files.NewFileOp().WriteFile(filePath, strings.NewReader(oldContent), 0644)
		return err
	}

	return nil
}

func getNginxConfig(alias string) (dto.NginxConfig, error) {
	var nginxConfig dto.NginxConfig

	nginxInstall, err := getAppInstallByKey("nginx")
	if err != nil {
		return nginxConfig, err
	}

	configPath := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "conf", "conf.d", alias+".conf")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nginxConfig, err
	}
	config := parser.NewStringParser(string(content)).Parse()
	config.FilePath = configPath
	nginxConfig.Config = config
	nginxConfig.OldContent = string(content)
	nginxConfig.ContainerName = nginxInstall.ContainerName
	nginxConfig.FilePath = configPath

	return nginxConfig, nil
}

func addListenAndServerName(website model.WebSite, ports []int, domains []string) error {

	nginxConfig, err := getNginxConfig(website.Alias)
	if err != nil {
		return nil
	}
	config := nginxConfig.Config
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
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxConfig.ContainerName)
}

func deleteListenAndServerName(website model.WebSite, ports []int, domains []string) error {

	nginxConfig, err := getNginxConfig(website.Alias)
	if err != nil {
		return nil
	}
	config := nginxConfig.Config
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
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxConfig.ContainerName)
}

func getNginxConfigByKeys(website model.WebSite, keys []string) ([]dto.NginxParam, error) {
	nginxConfig, err := getNginxConfig(website.Alias)
	if err != nil {
		return nil, err
	}
	config := nginxConfig.Config
	server := config.FindServers()[0]

	var res []dto.NginxParam
	for _, key := range keys {
		dirs := server.FindDirectives(key)
		for _, dir := range dirs {
			nginxParam := dto.NginxParam{
				Name:   dir.GetName(),
				Params: dir.GetParameters(),
			}
			if isRepeatKey(key) {
				nginxParam.SecondKey = dir.GetParameters()[0]
			}
			res = append(res, nginxParam)
		}
	}
	return res, nil
}

func updateNginxConfig(website model.WebSite, params []dto.NginxParam, scope dto.NginxKey) error {
	nginxConfig, err := getNginxConfig(website.Alias)
	if err != nil {
		return err
	}
	config := nginxConfig.Config
	updateConfig(config, scope)
	server := config.FindServers()[0]
	for _, p := range params {
		newDir := components.Directive{
			Name:       p.Name,
			Parameters: p.Params,
		}
		if isRepeatKey(p.Name) {
			server.UpdateDirectiveBySecondKey(p.Name, p.SecondKey, newDir)
		} else {
			server.UpdateDirectives(p.Name, newDir)
		}
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxConfig.ContainerName)
}

func updateConfig(config *components.Config, scope dto.NginxKey) {
	newConfig := &components.Config{}
	switch scope {
	case dto.LimitConn:
		newConfig = parser.NewStringParser(string(nginx_conf.Limit)).Parse()
	}
	if reflect.DeepEqual(newConfig, &components.Config{}) {
		return
	}

	for _, dir := range newConfig.GetDirectives() {
		newDir := components.Directive{
			Name:       dir.GetName(),
			Parameters: dir.GetParameters(),
		}
		if isRepeatKey(dir.GetName()) {
			config.UpdateDirectiveBySecondKey(dir.GetName(), dir.GetParameters()[0], newDir)
		} else {
			config.UpdateDirectives(dir.GetName(), newDir)
		}
	}
}

func getNginxParamsFromStaticFile(scope dto.NginxKey) []dto.NginxParam {
	var nginxParams []dto.NginxParam
	newConfig := &components.Config{}
	switch scope {
	case dto.SSL:
		newConfig = parser.NewStringParser(string(nginx_conf.SSL)).Parse()
	}
	for _, dir := range newConfig.GetDirectives() {
		nginxParams = append(nginxParams, dto.NginxParam{
			Name:   dir.GetName(),
			Params: dir.GetParameters(),
		})
	}
	return nginxParams
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

func deleteNginxConfig(website model.WebSite, keys []string) error {
	nginxConfig, err := getNginxConfig(website.Alias)
	if err != nil {
		return err
	}
	config := nginxConfig.Config
	config.RemoveDirectives(keys)
	server := config.FindServers()[0]
	server.RemoveDirectives(keys)
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxConfig.ContainerName)
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

	configDir := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "ssl", website.Alias)
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

	nginxConfig, err := getNginxConfig(website.Alias)
	if err != nil {
		return nil
	}
	config := nginxConfig.Config
	server := config.FindServers()[0]
	server.UpdateListen("443", false)
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}

	if err := createPemFile(website, websiteSSL); err != nil {
		return err
	}
	nginxParams := getNginxParamsFromStaticFile(dto.SSL)
	for i, param := range nginxParams {
		if param.Name == "ssl_certificate" {
			nginxParams[i].Params = []string{path.Join("/etc/nginx/ssl", website.Alias, "fullchain.pem")}
		}
		if param.Name == "ssl_certificate_key" {
			nginxParams[i].Params = []string{path.Join("/etc/nginx/ssl", website.Alias, "privkey.pem")}
		}
	}
	if err := updateNginxConfig(website, nginxParams, dto.SSL); err != nil {
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
				if isRepeatKey(k) {
					param.SecondKey = param.Params[0]
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

func isRepeatKey(key string) bool {

	if _, ok := dto.RepeatKeys[key]; ok {
		return true
	}
	return false
}

func toMapStr(m map[string]interface{}) map[string]string {
	ret := make(map[string]string, len(m))
	for k, v := range m {
		ret[k] = fmt.Sprint(v)
	}
	return ret
}
