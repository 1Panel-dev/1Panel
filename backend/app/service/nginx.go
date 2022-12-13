package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
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
	res, err := http.Get("http://127.0.0.1/nginx_status")
	if err != nil {
		return response.NginxStatus{}, err
	}
	content, err := ioutil.ReadAll(res.Body)
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
	if req.Backup {
		backupPath := path.Join(path.Dir(req.FilePath), "bak")
		if !fileOp.Stat(backupPath) {
			if err := fileOp.CreateDir(backupPath, 0755); err != nil {
				return err
			}
		}
		newFile := path.Join(backupPath, "nginx.bak"+"-"+time.Now().Format("2006-01-02-15-04-05"))
		if err := fileOp.Copy(req.FilePath, backupPath); err != nil {
			return err
		}
		if err := fileOp.Rename(path.Join(backupPath, "nginx.conf"), newFile); err != nil {
			return err
		}
	}
	oldContent, err := os.ReadFile(req.FilePath)
	if err != nil {
		return err
	}
	if err := fileOp.WriteFile(req.FilePath, strings.NewReader(req.Content), 0644); err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey("nginx")
	if err != nil {
		return err
	}
	return nginxCheckAndReload(string(oldContent), req.FilePath, nginxInstall.ContainerName)
}
