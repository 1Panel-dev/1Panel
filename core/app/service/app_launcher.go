package service

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/constant"
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
	launchers, err := launcherRepo.List(commonRepo.WithOrderBy("created_at"))
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
	launcher, _ := launcherRepo.Get(commonRepo.WithByKey(req.Key))
	if req.Value == constant.StatusEnable {
		if launcher.ID != 0 {
			return nil
		}
		launcher.Key = req.Key
		return launcherRepo.Create(&launcher)
	}
	if launcher.ID == 0 {
		return nil
	}

	return launcherRepo.Delete(commonRepo.WithByKey(req.Key))
}
