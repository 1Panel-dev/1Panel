package service

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
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
	if req.Value == constant.StatusEnable {
		if launcher.ID != 0 {
			go syncLauncherToAgent(launcher, "create")
			return nil
		}
		launcher.Key = req.Key
		if err := launcherRepo.Create(&launcher); err != nil {
			return err
		}
		go syncLauncherToAgent(launcher, "create")
		return nil
	}
	if launcher.ID == 0 {
		go syncLauncherToAgent(launcher, "delete")
		return nil
	}
	if err := launcherRepo.Delete(repo.WithByKey(req.Key)); err != nil {
		return err
	}
	go syncLauncherToAgent(launcher, "delete")
	return nil
}

func syncLauncherToAgent(launcher model.AppLauncher, operation string) {
	itemData, _ := json.Marshal(launcher)
	itemJson := dto.SyncToAgent{Name: launcher.Key, Operation: operation, Data: string(itemData)}
	bodyItem, _ := json.Marshal(itemJson)
	_ = xpack.RequestToAllAgent("/api/v2/backups/sync", http.MethodPost, bytes.NewReader((bodyItem)))
}
