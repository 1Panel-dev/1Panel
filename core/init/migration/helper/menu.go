package helper

import (
	"encoding/json"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

func UpdateHideMenu(tx *gorm.DB, update func([]dto.ShowMenu) []dto.ShowMenu) error {
	var menuJSON string
	if err := tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Pluck("value", &menuJSON).Error; err != nil {
		return err
	}
	if menuJSON == "" {
		menuJSON = LoadMenus()
	}

	var menus []dto.ShowMenu
	if err := json.Unmarshal([]byte(menuJSON), &menus); err != nil {
		return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", LoadMenus()).Error
	}

	updatedJSON, err := json.Marshal(update(menus))
	if err != nil {
		return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", LoadMenus()).Error
	}
	return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", string(updatedJSON)).Error
}

func UpsertChildMenuByLabel(tx *gorm.DB, parentLabel string, newMenu dto.ShowMenu, afterLabel string) error {
	return UpdateHideMenu(tx, func(menus []dto.ShowMenu) []dto.ShowMenu {
		for i := range menus {
			if menus[i].Label != parentLabel {
				continue
			}
			menus[i].Children = UpsertMenuByLabel(menus[i].Children, newMenu, afterLabel)
			break
		}
		return menus
	})
}

func UpdateChildMenuSortByLabel(tx *gorm.DB, parentLabel string, menuSort []dto.MenuLabelSort) error {
	sortMap := make(map[string]int)
	for _, item := range menuSort {
		sortMap[item.Label] = item.Sort
	}

	return UpdateHideMenu(tx, func(menus []dto.ShowMenu) []dto.ShowMenu {
		for i := range menus {
			if menus[i].Label != parentLabel {
				continue
			}
			for j := range menus[i].Children {
				if sortVal, ok := sortMap[menus[i].Children[j].Label]; ok {
					menus[i].Children[j].Sort = sortVal
				}
			}
			break
		}
		return menus
	})
}

func UpsertMenuByLabel(children []dto.ShowMenu, newMenu dto.ShowMenu, afterLabel string) []dto.ShowMenu {
	for i := range children {
		if children[i].Label != newMenu.Label {
			continue
		}
		children[i].Disabled = newMenu.Disabled
		children[i].Title = newMenu.Title
		children[i].Path = newMenu.Path
		children[i].Sort = newMenu.Sort
		children[i].IsShow = newMenu.IsShow
		return children
	}

	insertIndex := len(children)
	for i := range children {
		if children[i].Label == afterLabel {
			insertIndex = i + 1
			break
		}
	}

	children = append(children, dto.ShowMenu{})
	copy(children[insertIndex+1:], children[insertIndex:])
	children[insertIndex] = newMenu
	return children
}

func LoadMenus() string {
	item := []dto.ShowMenu{
		{ID: "1", Disabled: true, Title: "menu.home", IsShow: true, Label: "Home-Menu", Path: "/", Sort: 100},
		{ID: "2", Disabled: true, Title: "menu.apps", IsShow: true, Label: "App-Menu", Path: "/apps/all", Sort: 200},
		{ID: "3", Disabled: false, Title: "menu.aiTools", IsShow: true, Label: "AI-Menu", Path: "/ai/model/account", Sort: 300,
			Children: []dto.ShowMenu{
				{ID: "44", Disabled: false, Title: "aiTools.agents.agent", IsShow: true, Label: "Agents", Path: "/ai/agents/agent", Sort: 50},
				{ID: "41", Disabled: false, Title: "aiTools.model.model", IsShow: true, Label: "AIModel", Path: "/ai/model/account", Sort: 100},
				{ID: "42", Disabled: false, Title: "menu.mcp", IsShow: true, Label: "MCPServer", Path: "/ai/mcp", Sort: 200},
				{ID: "43", Disabled: false, Title: "aiTools.gpu.gpu", IsShow: true, Label: "GPU", Path: "/ai/gpu", Sort: 300},
			}},
		{ID: "4", Disabled: false, Title: "menu.website", IsShow: true, Label: "Website-Menu", Path: "/websites", Sort: 400,
			Children: []dto.ShowMenu{
				{ID: "31", Disabled: false, Title: "menu.website", IsShow: true, Label: "Website", Path: "/websites", Sort: 100},
				{ID: "32", Disabled: false, Title: "menu.ssl", IsShow: true, Label: "SSL", Path: "/websites/ssl", Sort: 200},
				{ID: "34", Disabled: false, Title: "menu.template", IsShow: true, Label: "WebsiteTemplate", Path: "/websites/templates", Sort: 250},
				{ID: "33", Disabled: false, Title: "menu.runtime", IsShow: true, Label: "PHP", Path: "/websites/runtimes/php", Sort: 300},
			}},
		{ID: "5", Disabled: false, Title: "menu.database", IsShow: true, Label: "Database-Menu", Path: "/databases", Sort: 500},
		{ID: "6", Disabled: false, Title: "menu.container", IsShow: true, Label: "Container-Menu", Path: "/containers", Sort: 600},
		{ID: "7", Disabled: false, Title: "menu.system", IsShow: true, Label: "System-Menu", Path: "/hosts/files", Sort: 700,
			Children: []dto.ShowMenu{
				{ID: "71", Disabled: false, Title: "menu.files", IsShow: true, Label: "File", Path: "/hosts/files", Sort: 100},
				{ID: "72", Disabled: false, Title: "menu.monitor", IsShow: true, Label: "Monitorx", Path: "/hosts/monitor/monitor", Sort: 200},
				{ID: "74", Disabled: false, Title: "menu.firewall", IsShow: true, Label: "FirewallPort", Path: "/hosts/firewall/rules", Sort: 300},
				{ID: "75", Disabled: false, Title: "menu.processManage", IsShow: true, Label: "Process", Path: "/hosts/process/process", Sort: 400},
				{ID: "76", Disabled: false, Title: "menu.ssh", IsShow: true, Label: "SSH", Path: "/hosts/ssh/ssh", Sort: 500},
				{ID: "77", Disabled: false, Title: "menu.disk", IsShow: true, Label: "Disk", Path: "/hosts/disk", Sort: 600},
			}},
		{ID: "8", Disabled: false, Title: "menu.terminal", IsShow: true, Label: "Terminal-Menu", Path: "/hosts/terminal", Sort: 800},
		{ID: "10", Disabled: false, Title: "menu.cronjob", IsShow: true, Label: "Cronjob-Menu", Path: "/cronjobs", Sort: 900},
		{ID: "9", Disabled: false, Title: "menu.toolbox", IsShow: true, Label: "Toolbox-Menu", Path: "/toolbox", Sort: 1000},
		{ID: "11", Disabled: false, Title: "xpack.menu", IsShow: true, Label: "Xpack-Menu", Sort: 1100,
			Children: []dto.ShowMenu{
				{ID: "118", Disabled: false, Title: "xpack.app.app", IsShow: true, Label: "XApp", Path: "/xpack/app", Sort: 100},
				{ID: "112", Disabled: false, Title: "xpack.waf.name", IsShow: true, Label: "Dashboard", Path: "/xpack/waf/dashboard", Sort: 200},
				{ID: "111", Disabled: false, Title: "xpack.node.nodeManagement", IsShow: true, Label: "NodeDashboard", Path: "/xpack/node/dashboard", Sort: 300},
				{ID: "113", Disabled: false, Title: "xpack.monitor.name", IsShow: true, Label: "MonitorDashboard", Path: "/xpack/monitor/dashboard", Sort: 600},
				{ID: "115", Disabled: false, Title: "xpack.sync.menu", IsShow: true, Label: "Sync", Path: "/xpack/sync", Sort: 700},
				{ID: "119", Disabled: false, Title: "xpack.upage", IsShow: true, Label: "Upage", Path: "/xpack/upage", Sort: 800},
				{ID: "114", Disabled: false, Title: "xpack.tamper.tamper", IsShow: true, Label: "Tamper", Path: "/xpack/tamper", Sort: 1000},
				{ID: "120", Disabled: false, Title: "xpack.cluster.cluster", IsShow: true, Label: "Cluster", Path: "/xpack/cluster", Sort: 1100},
				{ID: "117", Disabled: false, Title: "xpack.setting.setting", IsShow: true, Label: "XSetting", Path: "/xpack/setting", Sort: 1200},
			}},
		{ID: "12", Disabled: false, Title: "menu.logs", IsShow: true, Label: "Log-Menu", Path: "/logs", Sort: 1200},
		{ID: "13", Disabled: true, Title: "menu.settings", IsShow: true, Label: "Setting-Menu", Path: "/settings", Sort: 1300},
	}
	if global.CONF.Base.IsEnterprise {
		for i := range item {
			if item[i].Label == "AI-Menu" {
				item[i].Children = UpsertMenuByLabel(item[i].Children, dto.ShowMenu{
					ID:       "46",
					Disabled: false,
					Title:    "aiTools.aiProxy.title",
					IsShow:   true,
					Label:    "AIProxyManagement",
					Path:     "/ai/ai-proxy/model-pool",
					Sort:     150,
				}, "AIModel")
				item[i].Children = UpsertMenuByLabel(item[i].Children, dto.ShowMenu{
					ID:       "47",
					Disabled: false,
					Title:    "aiTools.skillsHub.title",
					IsShow:   true,
					Label:    "SkillsHub",
					Path:     "/ai/skills-hub",
					Sort:     155,
				}, "AIProxyManagement")
				item[i].Children = UpsertMenuByLabel(item[i].Children, dto.ShowMenu{
					ID:       "45",
					Disabled: false,
					Title:    "aiTools.benchmark.title",
					IsShow:   true,
					Label:    "AIBenchmark",
					Path:     "/ai/benchmark",
					Sort:     160,
				}, "SkillsHub")
				continue
			}
			if item[i].Label != "Xpack-Menu" {
				continue
			}
			item[i].Children = UpsertMenuByLabel(item[i].Children, dto.ShowMenu{
				ID:       "121",
				Disabled: false,
				Title:    "xpack.user.userManage",
				IsShow:   true,
				Label:    "UserManagement",
				Path:     "/enterprise/users",
				Sort:     400,
			}, "NodeDashboard")
			item[i].Children = UpsertMenuByLabel(item[i].Children, dto.ShowMenu{
				ID:       "122",
				Disabled: false,
				Title:    "xpack.opsReport.name",
				IsShow:   true,
				Label:    "OpsReport",
				Path:     "/enterprise/ops-report",
				Sort:     500,
			}, "UserManagement")
			item[i].Children = UpsertMenuByLabel(item[i].Children, dto.ShowMenu{
				ID:       "123",
				Disabled: false,
				Title:    "xpack.vm.title",
				IsShow:   true,
				Label:    "VirtualMachine",
				Path:     "/enterprise/vm",
				Sort:     900,
			}, "Upage")
			break
		}
	}
	menu, _ := json.Marshal(item)
	return string(menu)
}

func MenuSort() []dto.MenuLabelSort {
	var MenuLabelsWithSort = []dto.MenuLabelSort{
		{Label: "Home-Menu", Sort: 100},
		{Label: "App-Menu", Sort: 200},
		{Label: "AI-Menu", Sort: 300},
		{Label: "Agents", Sort: 50},
		{Label: "AIModel", Sort: 100},
		{Label: "AIProxyManagement", Sort: 150},
		{Label: "SkillsHub", Sort: 155},
		{Label: "AIBenchmark", Sort: 160},
		{Label: "MCPServer", Sort: 200},
		{Label: "GPU", Sort: 300},
		{Label: "Website-Menu", Sort: 400},
		{Label: "Website", Sort: 100},
		{Label: "SSL", Sort: 200},
		{Label: "WebsiteTemplate", Sort: 250},
		{Label: "PHP", Sort: 300},
		{Label: "Database-Menu", Sort: 500},
		{Label: "Container-Menu", Sort: 600},
		{Label: "System-Menu", Sort: 700},
		{Label: "File", Sort: 100},
		{Label: "Monitorx", Sort: 200},
		{Label: "FirewallPort", Sort: 300},
		{Label: "Process", Sort: 400},
		{Label: "SSH", Sort: 500},
		{Label: "Disk", Sort: 600},
		{Label: "Terminal-Menu", Sort: 800},
		{Label: "Cronjob-Menu", Sort: 900},
		{Label: "Toolbox-Menu", Sort: 1000},
		{Label: "Xpack-Menu", Sort: 1100},
		{Label: "Log-Menu", Sort: 1200},
		{Label: "Setting-Menu", Sort: 1300},
	}
	MenuLabelsWithSort = append(MenuLabelsWithSort, XpackMenuSort()...)
	return MenuLabelsWithSort
}

func XpackMenuSort() []dto.MenuLabelSort {
	return []dto.MenuLabelSort{
		{Label: "XApp", Sort: 100},
		{Label: "Dashboard", Sort: 200},
		{Label: "Node", Sort: 300},
		{Label: "NodeDashboard", Sort: 300},
		{Label: "UserManagement", Sort: 400},
		{Label: "OpsReport", Sort: 500},
		{Label: "MonitorDashboard", Sort: 600},
		{Label: "Sync", Sort: 700},
		{Label: "Upage", Sort: 800},
		{Label: "VirtualMachine", Sort: 900},
		{Label: "Tamper", Sort: 1000},
		{Label: "Cluster", Sort: 1100},
		{Label: "XSetting", Sort: 1200},
	}
}

func AddMenu(newMenu dto.ShowMenu, parentMenuID string, tx *gorm.DB) error {
	return UpdateHideMenu(tx, func(menus []dto.ShowMenu) []dto.ShowMenu {
		if menuExistsByLabelAndPath(menus, newMenu.Label, newMenu.Path) {
			return menus
		}
		for i := range menus {
			if menus[i].ID != parentMenuID || menuExistsByID(menus[i].Children, newMenu.ID) {
				continue
			}
			menus[i].Children = append([]dto.ShowMenu{newMenu}, menus[i].Children...)
			break
		}
		return menus
	})
}

func menuExistsByID(menus []dto.ShowMenu, id string) bool {
	for _, menu := range menus {
		if menu.ID == id || menuExistsByID(menu.Children, id) {
			return true
		}
	}
	return false
}

func menuExistsByLabelAndPath(menus []dto.ShowMenu, label, path string) bool {
	for _, menu := range menus {
		if menu.Label == label && menu.Path == path {
			return true
		}
		if menuExistsByLabelAndPath(menu.Children, label, path) {
			return true
		}
	}
	return false
}

func RemoveMenuByID(menus []dto.ShowMenu, id string) []dto.ShowMenu {
	var result []dto.ShowMenu
	for _, menu := range menus {
		if menu.ID == id {
			continue
		}

		if len(menu.Children) > 0 {
			menu.Children = RemoveMenuByID(menu.Children, id)
		}

		result = append(result, menu)
	}
	return result
}
