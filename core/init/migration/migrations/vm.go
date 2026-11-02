package migrations

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/init/migration/helper"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var MoveVirtualMachineMenuToXpack = &gormigrate.Migration{
	ID: "20260908-move-vm-menu-to-xpack",
	Migrate: func(tx *gorm.DB) error {
		return helper.UpdateHideMenu(tx, func(menus []dto.ShowMenu) []dto.ShowMenu {
			for i := range menus {
				if menus[i].Label != "Xpack-Menu" {
					continue
				}
				for j := range menus[i].Children {
					if menus[i].Children[j].Label == "VirtualMachine" {
						// Keep the user's visibility preference and ordering.
						menus[i].Children[j].Path = "/xpack/vm"
						return menus
					}
				}
				menus[i].Children = helper.UpsertMenuByLabel(menus[i].Children, dto.ShowMenu{
					ID: "123", Title: "xpack.vm.title", IsShow: true,
					Label: "VirtualMachine", Path: "/xpack/vm", Sort: 900,
				}, "Sync")
				break
			}
			return menus
		})
	},
}

var MoveVirtualMachineMenuToCommunity = &gormigrate.Migration{
	ID: "20261102-move-vm-menu-to-community",
	Migrate: func(tx *gorm.DB) error {
		return helper.UpdateHideMenu(tx, func(menus []dto.ShowMenu) []dto.ShowMenu {
			systemIndex := -1
			for i := range menus {
				if menus[i].Label == "System-Menu" {
					systemIndex = i
					break
				}
			}
			if systemIndex < 0 {
				return menus
			}
			item := dto.ShowMenu{ID: "123", Title: "xpack.vm.title", IsShow: true, Label: "VirtualMachine", Path: "/vm", Sort: 900}
			for i := range menus {
				if menus[i].Label != "Xpack-Menu" {
					continue
				}
				children := menus[i].Children[:0]
				for _, child := range menus[i].Children {
					if child.Label == "VirtualMachine" {
						item = child
						item.Path = "/vm"
						continue
					}
					children = append(children, child)
				}
				menus[i].Children = children
			}
			for j := range menus[systemIndex].Children {
				if menus[systemIndex].Children[j].Label == "VirtualMachine" {
					menus[systemIndex].Children[j].Path = "/vm"
					return menus
				}
			}
			menus[systemIndex].Children = append(menus[systemIndex].Children, item)
			return menus
		})
	},
}
