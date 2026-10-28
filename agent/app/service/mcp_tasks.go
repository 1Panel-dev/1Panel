package service

import (
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/taskscope"
)

func ConfigurePanelMCPTasks() error {
	return taskscope.Configure(taskscope.Rule{Scope: constant.TaskScopePanelMCP, Restricted: true, ExternalExecutor: true})
}
