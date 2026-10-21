package service

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
)

func (u *DashboardService) ListQuickDirectories() ([]model.QuickDirectory, error) {
	return launcherRepo.ListQuickDirectories()
}

func quickDirectoryValues(req dto.QuickDirectoryCreate) (model.QuickDirectory, error) {
	item := model.QuickDirectory{Path: req.Path, Alias: strings.TrimSpace(req.Alias)}
	if strings.TrimSpace(item.Path) == "" {
		return item, buserr.WithDetail("ErrInvalidParams", "path is required", nil)
	}
	return item, nil
}

func (u *DashboardService) CreateQuickDirectory(req dto.QuickDirectoryCreate) (*model.QuickDirectory, error) {
	item, err := quickDirectoryValues(req)
	if err != nil {
		return nil, err
	}
	if err := launcherRepo.CreateQuickDirectory(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (u *DashboardService) UpdateQuickDirectory(req dto.QuickDirectoryUpdate) (*model.QuickDirectory, error) {
	item, err := quickDirectoryValues(req.QuickDirectoryCreate)
	if err != nil {
		return nil, err
	}
	item.ID = req.ID
	item.Legacy = req.Legacy
	if err := launcherRepo.UpdateQuickDirectory(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (u *DashboardService) DeleteQuickDirectory(req dto.QuickDirectoryDelete) error {
	return launcherRepo.DeleteQuickDirectory(req.ID, req.Legacy)
}
