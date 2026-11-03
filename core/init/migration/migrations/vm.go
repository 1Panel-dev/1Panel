package migrations

import (
	"sort"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/init/migration/helper"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var MoveVirtualMachineMenuToTopLevel = &gormigrate.Migration{
	ID: "20261103-move-vm-menu-to-top-level",
	Migrate: func(tx *gorm.DB) error {
		return helper.UpdateHideMenu(tx, func(menus []dto.ShowMenu) []dto.ShowMenu {
			item := dto.ShowMenu{ID: "123", Title: "xpack.vm.vm", IsShow: true, Label: "VirtualMachine", Path: "/vm", Sort: 650}
			found := false
			updated := make([]dto.ShowMenu, 0, len(menus)+1)
			for _, menu := range menus {
				if menu.Label == item.Label {
					if !found {
						item.IsShow = menu.IsShow
						found = true
					}
					continue
				}
				children := menu.Children[:0]
				for _, child := range menu.Children {
					if child.Label == item.Label {
						if !found {
							item.IsShow = child.IsShow
							found = true
						}
						continue
					}
					children = append(children, child)
				}
				menu.Children = children
				updated = append(updated, menu)
			}
			sort.SliceStable(updated, func(i, j int) bool {
				if updated[i].Sort == updated[j].Sort {
					return updated[i].ID < updated[j].ID
				}
				return updated[i].Sort < updated[j].Sort
			})
			insertIndex := len(updated)
			for i := range updated {
				if updated[i].Label == "Container-Menu" {
					insertIndex = i + 1
					break
				}
			}
			updated = append(updated, dto.ShowMenu{})
			copy(updated[insertIndex+1:], updated[insertIndex:])
			updated[insertIndex] = item
			for i := range updated {
				updated[i].Sort = (i + 1) * 100
			}
			return updated
		})
	},
}

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
