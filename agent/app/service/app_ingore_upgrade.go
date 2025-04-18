package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type AppIgnoreUpgradeService struct {
}

type IAppIgnoreUpgradeService interface {
	List() ([]response.AppIgnoreUpgradeDTO, error)
	CreateAppIgnore(req request.AppIgnoreUpgradeReq) error
	Delete(req request.ReqWithID) error
}

func NewIAppIgnoreUpgradeService() IAppIgnoreUpgradeService {
	return AppIgnoreUpgradeService{}
}

func (a AppIgnoreUpgradeService) List() ([]response.AppIgnoreUpgradeDTO, error) {
	var res []response.AppIgnoreUpgradeDTO
	ignores, err := appIgnoreUpgradeRepo.List()
	if err != nil {
		return nil, err
	}
	for _, ignore := range ignores {
		dto := response.AppIgnoreUpgradeDTO{
			ID:          ignore.ID,
			AppID:       ignore.AppID,
			AppDetailID: ignore.AppDetailID,
			Scope:       ignore.Scope,
		}
		app, err := appRepo.GetFirst(repo.WithByID(ignore.AppID))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = appIgnoreUpgradeRepo.Delete(repo.WithByID(ignore.ID))
			continue
		}
		dto.Icon = app.Icon
		if ignore.Scope == "version" {
			appDetail, err := appDetailRepo.GetFirst(repo.WithByID(ignore.AppDetailID))
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_ = appIgnoreUpgradeRepo.Delete(repo.WithByID(ignore.ID))
				continue
			}
			dto.Version = appDetail.Version
		}
		res = append(res, dto)
	}
	return res, nil
}

func (a AppIgnoreUpgradeService) CreateAppIgnore(req request.AppIgnoreUpgradeReq) error {
	appIgnoreUpgrade := model.AppIgnoreUpgrade{
		AppID: req.AppID,
		Scope: req.Scope,
	}
	if req.Scope == "version" {
		appIgnoreUpgrade.AppDetailID = req.AppDetailID
	}
	if req.Scope == "all" {
		_ = appIgnoreUpgradeRepo.Delete(appInstallRepo.WithAppId(req.AppID))
	}
	if err := appIgnoreUpgradeRepo.Create(&appIgnoreUpgrade); err != nil {
		return err
	}
	return nil
}

func (a AppIgnoreUpgradeService) Delete(req request.ReqWithID) error {
	return appIgnoreUpgradeRepo.Delete(repo.WithByID(req.ID))
}
