package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
)

func buildRuntime(runtime *model.Runtime, service *docker.ComposeService) {
	err := service.ComposeBuild()
	if err != nil {
		runtime.Status = constant.RuntimeError
		runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + err.Error()
	} else {
		runtime.Status = constant.RuntimeNormal
	}
	_ = runtimeRepo.Save(runtime)
}
