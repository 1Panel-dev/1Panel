package migrations

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/init/migration/helper"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var MoveOpsReportMenuToAgent = &gormigrate.Migration{
	ID: "20261103-move-ops-report-menu-to-agent",
	Migrate: func(tx *gorm.DB) error {
		return helper.UpdateHideMenu(tx, func(menus []dto.ShowMenu) []dto.ShowMenu {
			for i := range menus {
				if menus[i].Label != "Xpack-Menu" {
					continue
				}
				found := false
				for j := range menus[i].Children {
					if menus[i].Children[j].Label == "OpsReport" {
						menus[i].Children[j].Path = "/xpack/ops-report"
						found = true
					}
				}
				if !found {
					menus[i].Children = helper.UpsertMenuByLabel(menus[i].Children, dto.ShowMenu{
						ID: "122", Title: "xpack.opsReport.name", IsShow: true, Label: "OpsReport", Path: "/xpack/ops-report", Sort: 500,
					}, "NodeDashboard")
				}
			}
			return menus
		})
	},
}
