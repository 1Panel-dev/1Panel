package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"os"
	"path"
)

func getDefaultNginxConfig() (dto.NginxConfig, error) {
	var nginxConfig dto.NginxConfig

	nginxInstall, err := getAppInstallByKey("nginx")
	if err != nil {
		return nginxConfig, err
	}

	configPath := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "conf", "nginx.conf")
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

func getHttpConfigByKeys(keys []string) ([]dto.NginxParam, error) {
	nginxConfig, err := getDefaultNginxConfig()
	if err != nil {
		return nil, err
	}
	config := nginxConfig.Config
	http := config.FindHttp()

	var res []dto.NginxParam
	for _, key := range keys {
		dirs := http.FindDirectives(key)
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
		if len(dirs) == 0 {
			nginxParam := dto.NginxParam{
				Name:   key,
				Params: []string{},
			}
			res = append(res, nginxParam)
		}
	}
	return res, nil
}

func updateHttpNginxConfig(params []dto.NginxParam) error {
	nginxConfig, err := getDefaultNginxConfig()
	if err != nil {
		return err
	}
	config := nginxConfig.Config
	http := config.FindHttp()
	for _, p := range params {
		newDir := components.Directive{
			Name:       p.Name,
			Parameters: p.Params,
		}
		if isRepeatKey(p.Name) {
			http.UpdateDirectiveBySecondKey(p.Name, p.SecondKey, newDir)
		} else {
			http.UpdateDirectives(p.Name, newDir)
		}
	}
	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxConfig.ContainerName)
}
