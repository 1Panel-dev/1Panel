package db

import (
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

func Init() {
	global.DB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "agent.db"), "agent")
	global.TaskDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "task.db"), "task")
	global.MonitorDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "monitor.db"), "monitor")
	global.GPUMonitorDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "gpu_monitor.db"), "gpu_monitor")
	global.AlertDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "alert.db"), "alert")

	if _, err := os.Stat(path.Join(global.Dir.DbDir, "core.db")); err == nil {
		global.CoreDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "core.db"), "core")
	}
}
