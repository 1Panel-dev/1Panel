package service

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/utils/req_helper/proxy_local"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
)

type LauncherService struct{}

type IAppLauncher interface {
	Search() ([]string, error)
	ChangeShow(req dto.SettingUpdate) error
}

func NewIAppLauncher() IAppLauncher {
	return &LauncherService{}
}

func (u *LauncherService) Search() ([]string, error) {
	launchers, err := launcherRepo.List(repo.WithOrderBy("created_at"))
	if err != nil {
		return nil, err
	}
	var data []string
	for _, launcher := range launchers {
		data = append(data, launcher.Key)
	}
	return data, nil
}

func (u *LauncherService) ChangeShow(req dto.SettingUpdate) error {
	launcher, _ := launcherRepo.Get(repo.WithByKey(req.Key))
	if req.Value == constant.StatusEnable && launcher.ID == 0 {
		if err := launcherRepo.Create(&model.AppLauncher{Key: req.Key}); err != nil {
			return err
		}
	}
	if req.Value == constant.StatusDisable && launcher.ID != 0 {
		if err := launcherRepo.Delete(repo.WithByKey(req.Key)); err != nil {
			return err
		}
	}
	go syncLauncherToAgent()
	return nil
}

func syncLauncherToAgent() {
	launchers, _ := launcherRepo.List()
	var list []string
	launcherMap := make(map[string]struct{})
	for _, item := range launchers {
		if _, ok := launcherMap[item.Key]; ok {
			continue
		}
		launcherMap[item.Key] = struct{}{}
		list = append(list, item.Key)
	}
	launcherData := struct {
		Keys []string
	}{Keys: list}
	itemData, _ := json.Marshal(launcherData)
	_, _ = proxy_local.NewLocalClient("/api/v2/dashboard/app/launcher/sync", http.MethodPost, bytes.NewReader((itemData)))
	_ = xpack.RequestToAllAgent("/api/v2/dashboard/app/launcher/sync", http.MethodPost, bytes.NewReader((itemData)))
}
