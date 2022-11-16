package service

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/jinzhu/copier"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
)

type DashboardService struct{}

type IDashboardService interface {
	LoadBaseInfo() (*dto.DashboardBase, error)
}

func NewIDashboardService() IDashboardService {
	return &DashboardService{}
}
func (u *DashboardService) LoadBaseInfo() (*dto.DashboardBase, error) {
	var baseInfo dto.DashboardBase
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}
	if err := copier.Copy(baseInfo, hostInfo); err != nil {
		return nil, err
	}
	appInstall, err := appInstallRepo.GetBy()
	if err != nil {
		return nil, err
	}
	for _, app := range appInstall {
		switch app.App.Key {
		case "dateease":
			baseInfo.DateeaseEnabled = true
		case "halo":
			baseInfo.HaloEnabled = true
		case "metersphere":
			baseInfo.MeterSphereEnabled = true
		case "jumpserver":
			baseInfo.JumpServerEnabled = true
		}
	}
	baseInfo.AppInstalldNumber = len(appInstall)
	dbs, err := mysqlRepo.List()
	if err != nil {
		return nil, err
	}
	baseInfo.DatabaseNumber = len(dbs)
	cornjobs, err := cronjobRepo.List()
	if err != nil {
		return nil, err
	}
	baseInfo.DatabaseNumber = len(cornjobs)

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	baseInfo.CPUModelName = cpuInfo[0].ModelName
	baseInfo.CPUCores, _ = cpu.Counts(false)
	baseInfo.CPULogicalCores, _ = cpu.Counts(true)
	totalPercent, _ := cpu.Percent(1*time.Second, false)
	if len(totalPercent) == 1 {
		baseInfo.CPUPercent = totalPercent[0]
	}

	return &baseInfo, nil
}
