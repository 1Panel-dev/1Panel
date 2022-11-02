package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
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

	nginxFileName := website.PrimaryDomain + ".conf"
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
		server.UpdateListen(string(rune(domain.Port)), false)
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
	if _, err := cmd.Exec(nginxCmd); err != nil {
		return err
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

	nginxFileName := website.PrimaryDomain + ".conf"
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

func delApp() error {
	return nil
}
