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

func configDefaultNginx(website *model.WebSite, domains []model.WebSiteDomain) error {

	nginxApp, err := appRepo.GetFirst(appRepo.WithKey("nginx"))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		return err
	}
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
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
	proxy := fmt.Sprintf("http://%s:%d", appInstall.ServiceName, appInstall.HttpPort)
	server.UpdateRootProxy([]string{proxy})

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
	nginxApp, err := appRepo.GetFirst(appRepo.WithKey("nginx"))
	if err != nil {
		return nginxConfig, err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
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
				nginxParam.IsRepeatKey = true
				nginxParam.SecondKey = dir.GetParameters()[0]
			}
			res = append(res, nginxParam)
		}
	}
	return res, nil
}

func updateNginxConfig(website model.WebSite, params []dto.NginxParam, scope dto.NginxScope) error {
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
		if p.IsRepeatKey {
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

func updateConfig(config *components.Config, scope dto.NginxScope) {
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
		config.UpdateDirectiveBySecondKey(dir.GetName(), dir.GetParameters()[0], newDir)
	}
}

func getNginxParamsFromStaticFile(scope dto.NginxScope) []dto.NginxParam {
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

func createPemFile(websiteSSL model.WebSiteSSL) error {
	nginxApp, err := appRepo.GetFirst(appRepo.WithKey("nginx"))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		return err
	}

	configDir := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "ssl", websiteSSL.PrimaryDomain)
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

func getParamArray(key string, param interface{}) []string {
	var res []string
	switch param.(type) {
	case string:
		if key == "index" {
			res = strings.Split(param.(string), "\n")
			return res
		}

		res = strings.Split(param.(string), " ")
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
					param.IsRepeatKey = true
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

	switch params.(type) {
	case map[string]interface{}:
		return handleParamMap(toMapStr(params.(map[string]interface{})), keys)
	case []interface{}:

		if mArray, ok := params.([]interface{}); ok {
			for _, mA := range mArray {
				if m, ok := mA.(map[string]interface{}); ok {
					nginxParams = append(nginxParams, handleParamMap(toMapStr(m), keys)...)
				}
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
