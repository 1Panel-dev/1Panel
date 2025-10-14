package helper

import (
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"gorm.io/gorm"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
)

func LoadMenus() string {
	item := []dto.ShowMenu{
		{ID: "1", Disabled: true, Title: "menu.home", IsShow: true, Label: "Home-Menu", Path: "/"},
		{ID: "2", Disabled: true, Title: "menu.apps", IsShow: true, Label: "App-Menu", Path: "/apps/all"},
		{ID: "3", Disabled: false, Title: "menu.website", IsShow: true, Label: "Website-Menu", Path: "/websites",
			Children: []dto.ShowMenu{
				{ID: "31", Disabled: false, Title: "menu.website", IsShow: true, Label: "Website", Path: "/websites"},
				{ID: "32", Disabled: false, Title: "menu.ssl", IsShow: true, Label: "SSL", Path: "/websites/ssl"},
				{ID: "33", Disabled: false, Title: "menu.runtime", IsShow: true, Label: "PHP", Path: "/websites/runtimes/php"},
			}},
		{ID: "4", Disabled: false, Title: "menu.aiTools", IsShow: true, Label: "AI-Menu", Path: "/ai/model",
			Children: []dto.ShowMenu{
				{ID: "41", Disabled: false, Title: "aiTools.model.model", IsShow: true, Label: "OllamaModel", Path: "/ai/model"},
				{ID: "42", Disabled: false, Title: "menu.mcp", IsShow: true, Label: "MCPServer", Path: "/ai/mcp"},
				{ID: "43", Disabled: false, Title: "aiTools.gpu.gpu", IsShow: true, Label: "GPU", Path: "/ai/gpu"},
			}},
		{ID: "5", Disabled: false, Title: "menu.database", IsShow: true, Label: "Database-Menu", Path: "/databases"},
		{ID: "6", Disabled: false, Title: "menu.container", IsShow: true, Label: "Container-Menu", Path: "/containers"},
		{ID: "7", Disabled: false, Title: "menu.system", IsShow: true, Label: "System-Menu", Path: "/hosts/files",
			Children: []dto.ShowMenu{
				{ID: "71", Disabled: false, Title: "menu.files", IsShow: true, Label: "File", Path: "/hosts/files"},
				{ID: "72", Disabled: false, Title: "menu.monitor", IsShow: true, Label: "Monitorx", Path: "/hosts/monitor/monitor"},
				{ID: "74", Disabled: false, Title: "menu.firewall", IsShow: true, Label: "FirewallPort", Path: "/hosts/firewall/port"},
				{ID: "75", Disabled: false, Title: "menu.processManage", IsShow: true, Label: "Process", Path: "/hosts/process/process"},
				{ID: "76", Disabled: false, Title: "menu.ssh", IsShow: true, Label: "SSH", Path: "/hosts/ssh/ssh"},
			}},
		{ID: "8", Disabled: false, Title: "menu.terminal", IsShow: true, Label: "Terminal-Menu", Path: "/hosts/terminal"},
		{ID: "10", Disabled: false, Title: "menu.cronjob", IsShow: true, Label: "Cronjob-Menu", Path: "/cronjobs"},
		{ID: "9", Disabled: false, Title: "menu.toolbox", IsShow: true, Label: "Toolbox-Menu", Path: "/toolbox"},
		{ID: "11", Disabled: false, Title: "xpack.menu", IsShow: true, Label: "Xpack-Menu",
			Children: []dto.ShowMenu{
				{ID: "118", Disabled: false, Title: "xpack.app.app", IsShow: true, Label: "XApp", Path: "/xpack/app"},
				{ID: "112", Disabled: false, Title: "xpack.waf.name", IsShow: true, Label: "Dashboard", Path: "/xpack/waf/dashboard"},
				{ID: "111", Disabled: false, Title: "xpack.node.nodeManagement", IsShow: true, Label: "Node", Path: "/xpack/node"},
				{ID: "119", Disabled: false, Title: "xpack.upage", IsShow: true, Label: "Upage", Path: "/xpack/upage"},
				{ID: "113", Disabled: false, Title: "xpack.monitor.name", IsShow: true, Label: "MonitorDashboard", Path: "/xpack/monitor/dashboard"},
				{ID: "114", Disabled: false, Title: "xpack.tamper.tamper", IsShow: true, Label: "Tamper", Path: "/xpack/tamper"},
				{ID: "115", Disabled: false, Title: "xpack.exchange.exchange", IsShow: true, Label: "FileExange", Path: "/xpack/exchange/file"},
				{ID: "117", Disabled: false, Title: "xpack.setting.setting", IsShow: true, Label: "XSetting", Path: "/xpack/setting"},
			}},
		{ID: "12", Disabled: false, Title: "menu.logs", IsShow: true, Label: "Log-Menu", Path: "/logs"},
		{ID: "13", Disabled: true, Title: "menu.settings", IsShow: true, Label: "Setting-Menu", Path: "/settings"},
	}
	menu, _ := json.Marshal(item)
	return string(menu)
}

func MenuSort() []dto.MenuLabelSort {
	var MenuLabelsWithSort = []dto.MenuLabelSort{
		{Label: "Home-Menu", Sort: 100},
		{Label: "App-Menu", Sort: 200},
		{Label: "Website-Menu", Sort: 300},
		{Label: "Website", Sort: 100},
		{Label: "SSL", Sort: 200},
		{Label: "PHP", Sort: 300},
		{Label: "AI-Menu", Sort: 400},
		{Label: "OllamaModel", Sort: 100},
		{Label: "MCPServer", Sort: 200},
		{Label: "GPU", Sort: 300},
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
		{Label: "XApp", Sort: 100},
		{Label: "Dashboard", Sort: 200},
		{Label: "Node", Sort: 300},
		{Label: "Upage", Sort: 400},
		{Label: "MonitorDashboard", Sort: 500},
		{Label: "Tamper", Sort: 600},
		{Label: "Cluster", Sort: 700},
		{Label: "FileExange", Sort: 800},
		{Label: "XSetting", Sort: 900},
		{Label: "Log-Menu", Sort: 1200},
		{Label: "Setting-Menu", Sort: 1300},
	}
	return MenuLabelsWithSort
}

func AddMenu(newMenu dto.ShowMenu, parentMenuID string, tx *gorm.DB) error {
	var menuJSON string
	if err := tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Pluck("value", &menuJSON).Error; err != nil {
		return err
	}
	if strings.Contains(menuJSON, fmt.Sprintf(`"%s"`, newMenu.Label)) && strings.Contains(menuJSON, fmt.Sprintf(`"%s"`, newMenu.Path)) {
		return nil
	}
	var menus []dto.ShowMenu
	if err := json.Unmarshal([]byte(menuJSON), &menus); err != nil {
		return tx.Model(&model.Setting{}).
			Where("key = ?", "HideMenu").
			Update("value", LoadMenus()).Error
	}
	for i, menu := range menus {
		if menu.ID == parentMenuID {
			exists := false
			for _, child := range menu.Children {
				if child.ID == newMenu.ID {
					exists = true
					break
				}
			}
			if !exists {
				menus[i].Children = append([]dto.ShowMenu{newMenu}, menus[i].Children...)
			}
			break
		}
	}
	updatedJSON, err := json.Marshal(menus)
	if err != nil {
		return tx.Model(&model.Setting{}).
			Where("key = ?", "HideMenu").
			Update("value", LoadMenus()).Error
	}
	return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", string(updatedJSON)).Error
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
