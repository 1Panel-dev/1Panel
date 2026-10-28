package job

import (
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type containerImageCheckUpdate struct{}

func NewContainerImageCheckUpdateJob() *containerImageCheckUpdate {
	return &containerImageCheckUpdate{}
}
func (j *containerImageCheckUpdate) Run() {
	status, err := service.GetContainerImageCheckUpdateStatus()
	if err != nil {
		global.LOG.Errorf("load container image update schedule: %v", err)
		return
	}
	if status.Enabled {
		if _, err := service.StartContainerImageCheckUpdate(); err != nil {
			global.LOG.Errorf("start container image update check: %v", err)
		}
	}
}
