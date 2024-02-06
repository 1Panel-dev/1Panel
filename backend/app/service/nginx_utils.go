package service

import (
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/cmd/server/nginx_conf"
	"os"
	"path"
	"strings"
	"time"
)

func getNginxFull(website *model.Website) (dto.NginxFull, error) {
	var nginxFull dto.NginxFull
	nginxInstall, err := getAppInstallByKey("openresty")
	if err != nil {
		return nginxFull, err
	}
	nginxFull.Install = nginxInstall
	nginxFull.Dir = path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name)
	nginxFull.ConfigDir = path.Join(nginxFull.Dir, "conf")
	nginxFull.ConfigFile = "nginx.conf"
	nginxFull.SiteDir = path.Join(nginxFull.Dir, "www")

	var nginxConfig dto.NginxConfig
	nginxConfig.FilePath = path.Join(nginxFull.Dir, "conf", "nginx.conf")
	content, err := os.ReadFile(path.Join(nginxFull.ConfigDir, nginxFull.ConfigFile))
	if err != nil {
		return nginxFull, err
	}
	config, err := parser.NewStringParser(string(content)).Parse()
	if err != nil {
		return dto.NginxFull{}, err
	}
	config.FilePath = nginxConfig.FilePath
	nginxConfig.OldContent = string(content)
	nginxConfig.Config = config

	nginxFull.RootConfig = nginxConfig

	if website != nil {
		nginxFull.Website = *website
		var siteNginxConfig dto.NginxConfig
		nginxFileName := website.Alias + ".conf"
		siteConfigPath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "conf", "conf.d", nginxFileName)
		siteNginxConfig.FilePath = siteConfigPath
		siteNginxContent, err := os.ReadFile(siteConfigPath)
		if err != nil {
			return nginxFull, err
		}
		siteConfig, err := parser.NewStringParser(string(siteNginxContent)).Parse()
		if err != nil {
			return dto.NginxFull{}, err
		}
		siteConfig.FilePath = siteConfigPath
		siteNginxConfig.Config = siteConfig
		siteNginxConfig.OldContent = string(siteNginxContent)
		nginxFull.SiteConfig = siteNginxConfig
	}

	return nginxFull, nil
}

func getNginxParamsByKeys(scope string, keys []string, website *model.Website) ([]response.NginxParam, error) {
	nginxFull, err := getNginxFull(website)
	if err != nil {
		return nil, err
	}
	var res []response.NginxParam
	var block components.IBlock
	if scope == constant.NginxScopeHttp {
		block = nginxFull.RootConfig.Config.FindHttp()
	} else {
		block = nginxFull.SiteConfig.Config.FindServers()[0]
	}
	for _, key := range keys {
		dirs := block.FindDirectives(key)
		for _, dir := range dirs {
			nginxParam := response.NginxParam{
				Name:   dir.GetName(),
				Params: dir.GetParameters(),
			}
			res = append(res, nginxParam)
		}
		if len(dirs) == 0 {
			nginxParam := response.NginxParam{
				Name:   key,
				Params: []string{},
			}
			res = append(res, nginxParam)
		}
	}
	return res, nil
}

func updateNginxConfig(scope string, params []dto.NginxParam, website *model.Website) error {
	nginxFull, err := getNginxFull(website)
	if err != nil {
		return err
	}
	var block components.IBlock
	var config dto.NginxConfig
	if scope == constant.NginxScopeHttp {
		config = nginxFull.RootConfig
		block = nginxFull.RootConfig.Config.FindHttp()
	} else if scope == constant.NginxScopeServer {
		config = nginxFull.SiteConfig
		block = nginxFull.SiteConfig.Config.FindServers()[0]
	} else {
		config = nginxFull.SiteConfig
		block = config.Config.Block
	}

	for _, p := range params {
		if p.UpdateScope == constant.NginxScopeOut {
			config.Config.UpdateDirective(p.Name, p.Params)
		} else {
			block.UpdateDirective(p.Name, p.Params)
		}
	}
	if err := nginx.WriteConfig(config.Config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(config.OldContent, config.FilePath, nginxFull.Install.ContainerName)
}

func deleteNginxConfig(scope string, params []dto.NginxParam, website *model.Website) error {
	nginxFull, err := getNginxFull(website)
	if err != nil {
		return err
	}
	var block components.IBlock
	var config dto.NginxConfig
	if scope == constant.NginxScopeHttp {
		config = nginxFull.RootConfig
		block = nginxFull.RootConfig.Config.FindHttp()
	} else if scope == constant.NginxScopeServer {
		config = nginxFull.SiteConfig
		block = nginxFull.SiteConfig.Config.FindServers()[0]
	} else {
		config = nginxFull.SiteConfig
		block = config.Config.Block
	}

	for _, param := range params {
		block.RemoveDirective(param.Name, param.Params)
	}

	if err := nginx.WriteConfig(config.Config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(config.OldContent, config.FilePath, nginxFull.Install.ContainerName)
}

func getNginxParamsFromStaticFile(scope dto.NginxKey, newParams []dto.NginxParam) []dto.NginxParam {
	var (
		newConfig = &components.Config{}
		err       error
	)

	updateScope := "in"
	switch scope {
	case dto.SSL:
		newConfig, err = parser.NewStringParser(string(nginx_conf.SSL)).Parse()
	case dto.CACHE:
		newConfig, err = parser.NewStringParser(string(nginx_conf.Cache)).Parse()
	case dto.ProxyCache:
		newConfig, err = parser.NewStringParser(string(nginx_conf.ProxyCache)).Parse()
	}
	if err != nil {
		return nil
	}
	for _, dir := range newConfig.GetDirectives() {
		addParam := dto.NginxParam{
			Name:        dir.GetName(),
			Params:      dir.GetParameters(),
			UpdateScope: updateScope,
		}
		isExist := false
		for _, newParam := range newParams {
			if newParam.Name == dir.GetName() {
				if components.IsRepeatKey(newParam.Name) {
					if len(newParam.Params) > 0 && newParam.Params[0] == dir.GetParameters()[0] {
						isExist = true
					}
				} else {
					isExist = true
				}
			}
		}
		if !isExist {
			newParams = append(newParams, addParam)
		}
	}
	return newParams
}

func opNginx(containerName, operate string) error {
	nginxCmd := fmt.Sprintf("docker exec -i %s %s", containerName, "nginx -s reload")
	if operate == constant.NginxCheck {
		nginxCmd = fmt.Sprintf("docker exec -i %s %s", containerName, "nginx -t")
	}
	if out, err := cmd.ExecWithTimeOut(nginxCmd, 20*time.Second); err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	return nil
}

func nginxCheckAndReload(oldContent string, filePath string, containerName string) error {
	if err := opNginx(containerName, constant.NginxCheck); err != nil {
		_ = files.NewFileOp().WriteFile(filePath, strings.NewReader(oldContent), 0644)
		return err
	}
	if err := opNginx(containerName, constant.NginxReload); err != nil {
		_ = files.NewFileOp().WriteFile(filePath, strings.NewReader(oldContent), 0644)
		return err
	}
	return nil
}
