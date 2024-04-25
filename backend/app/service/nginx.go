package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/utils/compose"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type NginxService struct {
}

type INginxService interface {
	GetNginxConfig() (*response.NginxFile, error)
	GetConfigByScope(req request.NginxScopeReq) ([]response.NginxParam, error)
	UpdateConfigByScope(req request.NginxConfigUpdate) error
	GetStatus() (response.NginxStatus, error)
	UpdateConfigFile(req request.NginxConfigFileUpdate) error
	ClearProxyCache() error
}

func NewINginxService() INginxService {
	return &NginxService{}
}

func (n NginxService) GetNginxConfig() (*response.NginxFile, error) {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	configPath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "conf", "nginx.conf")
	byteContent, err := files.NewFileOp().GetContent(configPath)
	if err != nil {
		return nil, err
	}
	return &response.NginxFile{Content: string(byteContent)}, nil
}

func (n NginxService) GetConfigByScope(req request.NginxScopeReq) ([]response.NginxParam, error) {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil, nil
	}
	return getNginxParamsByKeys(constant.NginxScopeHttp, keys, nil)
}

func (n NginxService) UpdateConfigByScope(req request.NginxConfigUpdate) error {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil
	}
	return updateNginxConfig(constant.NginxScopeHttp, getNginxParams(req.Params, keys), nil)
}

func (n NginxService) GetStatus() (response.NginxStatus, error) {
	httpPort, _, err := getAppInstallPort(constant.AppOpenresty)
	if err != nil {
		return response.NginxStatus{}, err
	}
	url := "http://127.0.0.1/nginx_status"
	if httpPort != 80 {
		url = fmt.Sprintf("http://127.0.0.1:%v/nginx_status", httpPort)
	}
	res, err := http.Get(url)
	if err != nil {
		return response.NginxStatus{}, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return response.NginxStatus{}, err
	}
	var status response.NginxStatus
	resArray := strings.Split(string(content), " ")
	status.Active = resArray[2]
	status.Accepts = resArray[7]
	status.Handled = resArray[8]
	status.Requests = resArray[9]
	status.Reading = resArray[11]
	status.Writing = resArray[13]
	status.Waiting = resArray[15]
	return status, nil
}

func (n NginxService) UpdateConfigFile(req request.NginxConfigFileUpdate) error {
	fileOp := files.NewFileOp()
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	filePath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "conf", "nginx.conf")
	if err != nil {
		return err
	}
	if req.Backup {
		backupPath := path.Join(path.Dir(filePath), "bak")
		if !fileOp.Stat(backupPath) {
			if err := fileOp.CreateDir(backupPath, 0755); err != nil {
				return err
			}
		}
		newFile := path.Join(backupPath, "nginx.bak"+"-"+time.Now().Format("2006-01-02-15-04-05"))
		if err := fileOp.Copy(filePath, backupPath); err != nil {
			return err
		}
		if err := fileOp.Rename(path.Join(backupPath, "nginx.conf"), newFile); err != nil {
			return err
		}
	}
	oldContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if err = fileOp.WriteFile(filePath, strings.NewReader(req.Content), 0644); err != nil {
		return err
	}
	return nginxCheckAndReload(string(oldContent), filePath, nginxInstall.ContainerName)
}

func (n NginxService) ClearProxyCache() error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	cacheDir := path.Join(nginxInstall.GetPath(), "www/common/proxy/proxy_cache_dir")
	fileOp := files.NewFileOp()
	if fileOp.Stat(cacheDir) {
		if err = fileOp.CleanDir(cacheDir); err != nil {
			return err
		}
		_, err = compose.Restart(nginxInstall.GetComposePath())
		if err != nil {
			return err
		}
	}
	return nil
}
