package helper

import (
	"encoding/json"

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
				{ID: "75", Disabled: false, Title: "menu.supervisor", IsShow: true, Label: "Process", Path: "/hosts/process/process"},
				{ID: "76", Disabled: false, Title: "menu.ssh", IsShow: true, Label: "SSH", Path: "/hosts/ssh/ssh"},
			}},
		{ID: "8", Disabled: false, Title: "menu.terminal", IsShow: true, Label: "Terminal-Menu", Path: "/hosts/terminal"},
		{ID: "9", Disabled: false, Title: "menu.toolbox", IsShow: true, Label: "Toolbox-Menu", Path: "/toolbox"},
		{ID: "10", Disabled: false, Title: "menu.cronjob", IsShow: true, Label: "Cronjob-Menu", Path: "/cronjobs"},
		{ID: "11", Disabled: false, Title: "xpack.menu", IsShow: true, Label: "Xpack-Menu",
			Children: []dto.ShowMenu{
				{ID: "118", Disabled: false, Title: "xpack.app.app", IsShow: true, Label: "XApp", Path: "/xpack/app"},
				{ID: "112", Disabled: false, Title: "xpack.waf.name", IsShow: true, Label: "Dashboard", Path: "/xpack/waf/dashboard"},
				{ID: "111", Disabled: false, Title: "xpack.node.nodeManagement", IsShow: true, Label: "Node", Path: "/xpack/node"},
				{ID: "113", Disabled: false, Title: "xpack.monitor.name", IsShow: true, Label: "MonitorDashboard", Path: "/xpack/monitor/dashboard"},
				{ID: "114", Disabled: false, Title: "xpack.tamper.tamper", IsShow: true, Label: "Tamper", Path: "/xpack/tamper"},
				{ID: "115", Disabled: false, Title: "xpack.exchange.exchange", IsShow: true, Label: "FileExange", Path: "/xpack/exchange/file"},
				{ID: "116", Disabled: false, Title: "xpack.alert.alert", IsShow: true, Label: "XAlertDashboard", Path: "/xpack/alert/dashboard"},
				{ID: "117", Disabled: false, Title: "xpack.setting.setting", IsShow: true, Label: "XSetting", Path: "/xpack/setting"},
			}},
		{ID: "12", Disabled: false, Title: "menu.logs", IsShow: true, Label: "Log-Menu", Path: "/logs"},
		{ID: "13", Disabled: true, Title: "menu.settings", IsShow: true, Label: "Setting-Menu", Path: "/settings"},
	}
	menu, _ := json.Marshal(item)
	return string(menu)
}
