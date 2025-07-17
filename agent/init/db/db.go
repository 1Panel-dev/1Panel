package db

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

func Init() {
	global.DB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "agent.db"), "agent")
	global.TaskDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "task.db"), "task")
	global.MonitorDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "monitor.db"), "monitor")
	global.AlertDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "alert.db"), "alert")

	if global.IsMaster {
		global.CoreDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "core.db"), "core")
	}
}
