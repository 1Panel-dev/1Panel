package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/cmd/server/nginx_conf"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func getDomain(domainStr string, websiteID uint) (model.WebsiteDomain, error) {
	domain := model.WebsiteDomain{
		WebsiteID: websiteID,
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
			return model.WebsiteDomain{}, err
		}
		domain.Port = portN
		return domain, nil
	}
	return model.WebsiteDomain{}, nil
}

func createStaticHtml(website *model.Website) error {
	nginxInstall, err := getAppInstallByKey(constant.AppNginx)
	if err != nil {
		return err
	}

	indexFolder := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name, "www", "sites", website.Alias)
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

func createWebsiteFolder(nginxInstall model.AppInstall, website *model.Website) error {
	nginxFolder := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name)
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
		if err := fileOp.CreateDir(path.Join(siteFolder, "data"), 0755); err != nil {
			return err
		}
		if err := fileOp.CreateDir(path.Join(siteFolder, "ssl"), 0755); err != nil {
			return err
		}
		if website.Type == constant.Static {
			if err := createStaticHtml(website); err != nil {
				return err
			}
		}
	}
	return fileOp.CopyDir(path.Join(nginxFolder, "www", "common", "waf", "rules"), path.Join(siteFolder, "waf"))
}

func configDefaultNginx(website *model.Website, domains []model.WebsiteDomain, appInstall *model.AppInstall) error {
	nginxInstall, err := getAppInstallByKey(constant.AppNginx)
	if err != nil {
		return err
	}
	if err := createWebsiteFolder(nginxInstall, website); err != nil {
		return err
	}

	nginxFileName := website.Alias + ".conf"
	configPath := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name, "conf", "conf.d", nginxFileName)
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

	switch website.Type {
	case constant.Deployment:
		proxy := fmt.Sprintf("http://127.0.0.1:%d", appInstall.HttpPort)
		server.UpdateRootProxy([]string{proxy})
	case constant.Static:
		server.UpdateRoot(path.Join("/www/sites", website.Alias))
		server.UpdateRootLocation()
	case constant.Proxy:
		server.UpdateRootProxy([]string{website.Proxy})
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
	nginxApp, err := appRepo.GetFirst(appRepo.WithKey(constant.AppNginx))
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
	configPath := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name, "conf", "conf.d", nginxFileName)
	fileOp := files.NewFileOp()

	if !fileOp.Stat(configPath) {
		return nil
	}
	if err := fileOp.DeleteFile(configPath); err != nil {
		return err
	}
	sitePath := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name, "www", "sites", website.PrimaryDomain)
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
	}
	for _, domain := range domains {
		server.AddServerName(domain)
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxFull.Install.ContainerName)
}

func deleteListenAndServerName(website model.Website, ports []int, domains []string) error {
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

func createPemFile(website model.Website, websiteSSL model.WebsiteSSL) error {
	nginxApp, err := appRepo.GetFirst(appRepo.WithKey(constant.AppNginx))
	if err != nil {
		return err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		return err
	}

	configDir := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name, "www", "sites", website.Alias, "ssl")
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

func applySSL(website model.Website, websiteSSL model.WebsiteSSL) error {
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

type WebsiteInfo struct {
	WebsiteName string `json:"websiteName"`
	WebsiteType string `json:"websiteType"`
}

func handleWebsiteBackup(backupType, baseDir, backupDir, domain, backupName string) error {
	website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(domain))
	if err != nil {
		return err
	}

	tmpDir := fmt.Sprintf("%s/%s/%s", baseDir, backupDir, backupName)
	if _, err := os.Stat(tmpDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			if err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", tmpDir, err)
			}
		}
	}
	if err := saveWebsiteJson(&website, tmpDir); err != nil {
		return err
	}

	nginxInfo, err := appInstallRepo.LoadBaseInfo(constant.AppNginx, "")
	if err != nil {
		return err
	}
	nginxConfFile := fmt.Sprintf("%s/nginx/%s/conf/conf.d/%s.conf", constant.AppInstallDir, nginxInfo.Name, website.Alias)
	fileOp := files.NewFileOp()
	if err := fileOp.CopyFile(nginxConfFile, tmpDir); err != nil {
		return err
	}

	if website.Type == constant.Deployment {
		if err := mysqlOperation(&website, "backup", tmpDir); err != nil {
			return err
		}
		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		websiteDir := fmt.Sprintf("%s/%s/%s", constant.AppInstallDir, app.App.Key, app.Name)
		if err := handleTar(websiteDir, tmpDir, fmt.Sprintf("%s.app.tar.gz", website.Alias), ""); err != nil {
			return err
		}
	}
	websiteDir := path.Join(constant.AppInstallDir, "nginx", nginxInfo.Name, "www", "sites", website.Alias)
	if err := handleTar(websiteDir, tmpDir, fmt.Sprintf("%s.web.tar.gz", website.Alias), ""); err != nil {
		return err
	}
	if err := handleTar(tmpDir, fmt.Sprintf("%s/%s", baseDir, backupDir), backupName+".tar.gz", ""); err != nil {
		return err
	}
	_ = os.RemoveAll(tmpDir)

	record := &model.BackupRecord{
		Type:       "website-" + website.Type,
		Name:       website.PrimaryDomain,
		DetailName: "",
		Source:     backupType,
		BackupType: backupType,
		FileDir:    backupDir,
		FileName:   fmt.Sprintf("%s.tar.gz", backupName),
	}
	if baseDir != constant.TmpDir || backupType == "LOCAL" {
		record.Source = "LOCAL"
		record.FileDir = fmt.Sprintf("%s/%s", baseDir, backupDir)
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
	}
	return nil
}

func handleWebsiteRecover(website *model.Website, fileDir string) error {
	nginxInfo, err := appInstallRepo.LoadBaseInfo(constant.AppNginx, "")
	if err != nil {
		return err
	}
	nginxConfPath := fmt.Sprintf("%s/nginx/%s/conf/conf.d", constant.AppInstallDir, nginxInfo.Name)
	if err := files.NewFileOp().CopyFile(path.Join(fileDir, website.Alias+".conf"), nginxConfPath); err != nil {
		return err
	}

	if website.Type == constant.Deployment {
		if err := mysqlOperation(website, "recover", fileDir); err != nil {
			return err
		}

		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		appDir := fmt.Sprintf("%s/%s", constant.AppInstallDir, app.App.Key)
		if err := handleUnTar(fmt.Sprintf("%s/%s.app.tar.gz", fileDir, website.Alias), appDir); err != nil {
			return err
		}
		if _, err := compose.Restart(fmt.Sprintf("%s/%s/docker-compose.yml", appDir, app.Name)); err != nil {
			return err
		}
	}
	siteDir := fmt.Sprintf("%s/nginx/%s/www/sites", constant.AppInstallDir, nginxInfo.Name)
	if err := handleUnTar(fmt.Sprintf("%s/%s.web.tar.gz", fileDir, website.Alias), siteDir); err != nil {
		return err
	}
	cmd := exec.Command("docker", "exec", "-i", nginxInfo.ContainerName, "nginx", "-s", "reload")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	_ = os.RemoveAll(fileDir)

	return nil
}

func mysqlOperation(website *model.Website, operation, filePath string) error {
	mysqlInfo, err := appInstallRepo.LoadBaseInfo(constant.AppMysql, "")
	if err != nil {
		return err
	}
	resource, err := appInstallResourceRepo.GetFirst(appInstallResourceRepo.WithAppInstallId(website.AppInstallID))
	if err != nil {
		return err
	}
	db, err := mysqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
	if err != nil {
		return err
	}
	if operation == "backup" {
		dbFile := fmt.Sprintf("%s/%s.sql", filePath, website.PrimaryDomain)
		outfile, _ := os.OpenFile(dbFile, os.O_RDWR|os.O_CREATE, 0755)
		defer outfile.Close()
		cmd := exec.Command("docker", "exec", mysqlInfo.ContainerName, "mysqldump", "-uroot", "-p"+mysqlInfo.Password, db.Name)
		cmd.Stdout = outfile
		_ = cmd.Run()
		_ = cmd.Wait()
		return nil
	}
	cmd := exec.Command("docker", "exec", "-i", mysqlInfo.ContainerName, "mysql", "-uroot", "-p"+mysqlInfo.Password, db.Name)
	sqlfile, err := os.Open(fmt.Sprintf("%s/%s.sql", filePath, website.Alias))
	if err != nil {
		return err
	}
	defer sqlfile.Close()
	cmd.Stdin = sqlfile
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func saveWebsiteJson(website *model.Website, tmpDir string) error {
	var websiteInfo WebsiteInfo
	websiteInfo.WebsiteType = website.Type
	websiteInfo.WebsiteName = website.PrimaryDomain
	remarkInfo, _ := json.Marshal(websiteInfo)
	path := fmt.Sprintf("%s/website.json", tmpDir)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(string(remarkInfo))
	write.Flush()
	return nil
}

func deleteWebsiteFolder(nginxInstall model.AppInstall, website *model.Website) error {
	nginxFolder := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name)
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
