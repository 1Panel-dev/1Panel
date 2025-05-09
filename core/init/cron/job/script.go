package job

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/global"
)

type script struct{}

func NewScriptJob() *script {
	return &script{}
}

func (s *script) Run() {
	if err := service.NewIScriptService().Sync(dto.OperateByTaskID{}); err != nil {
		global.LOG.Errorf("sync scripts from remote failed, err: %v", err)
	}
}
