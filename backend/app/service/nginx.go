package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"path"
)

type NginxService struct {
}

func (n NginxService) GetNginxConfig() (dto.FileInfo, error) {

	nginxInstall, err := getAppInstallByKey("nginx")
	if err != nil {
		return dto.FileInfo{}, err
	}

	configPath := path.Join(constant.AppInstallDir, "nginx", nginxInstall.Name, "conf", "nginx.conf")

	info, err := files.NewFileInfo(files.FileOption{
		Path:   configPath,
		Expand: true,
	})
	if err != nil {
		return dto.FileInfo{}, err
	}
	return dto.FileInfo{FileInfo: *info}, nil
}

func (n NginxService) GetConfigByScope(req dto.NginxScopeReq) ([]dto.NginxParam, error) {

	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil, nil
	}

	return getHttpConfigByKeys(keys)
}

func (n NginxService) UpdateConfigByScope(req dto.NginxConfigReq) error {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil
	}
	return updateHttpNginxConfig(getNginxParams(req.Params, keys))
}
