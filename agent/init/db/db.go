package db

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

func Init() {
	global.DB = common.LoadDBConnByPath(path.Join(global.CONF.System.DbPath, "agent.db"), "agent")
	global.TaskDB = common.LoadDBConnByPath(path.Join(global.CONF.System.DbPath, "task.db"), "task")
	global.MonitorDB = common.LoadDBConnByPath(path.Join(global.CONF.System.DbPath, "monitor.db"), "monitor")
}
