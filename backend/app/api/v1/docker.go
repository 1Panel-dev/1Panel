package v1

import (
	"os"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Container Docker
// @Summary Load docker status
// @Produce json
// @Success 200 {string} status
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/docker/status [get]
func (b *BaseApi) LoadDockerStatus(c *gin.Context) {
	status := dockerService.LoadDockerStatus()
	helper.SuccessWithData(c, status)
}

// @Tags Container Docker
// @Summary Load docker daemon.json
// @Produce json
// @Success 200 {object} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/daemonjson/file [get]
func (b *BaseApi) LoadDaemonJsonFile(c *gin.Context) {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil {
		helper.SuccessWithData(c, "daemon.json is not find in path")
		return
	}
	content, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, string(content))
}

// @Tags Container Docker
// @Summary Load docker daemon.json
// @Produce json
// @Success 200 {object} dto.DaemonJsonConf
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/daemonjson [get]
func (b *BaseApi) LoadDaemonJson(c *gin.Context) {
	conf := dockerService.LoadDockerConf()
	helper.SuccessWithData(c, conf)
}

// @Tags Container Docker
// @Summary Update docker daemon.json
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/daemonjson/update [post]
// @x-panel-log {"bodyKeys":["key", "value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新配置 [key]","formatEN":"Updated configuration [key]"}
func (b *BaseApi) UpdateDaemonJson(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := dockerService.UpdateConf(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Docker
// @Summary Update docker daemon.json log option
// @Accept json
// @Param request body dto.LogOption true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/logoption/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新日志配置","formatEN":"Updated the log option"}
func (b *BaseApi) UpdateLogOption(c *gin.Context) {
	var req dto.LogOption
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}

	if err := dockerService.UpdateLogOption(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Docker
// @Summary Update docker daemon.json ipv6 option
// @Accept json
// @Param request body dto.LogOption true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/ipv6option/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 ipv6 配置","formatEN":"Updated the ipv6 option"}
func (b *BaseApi) UpdateIpv6Option(c *gin.Context) {
	var req dto.Ipv6Option
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}

	if err := dockerService.UpdateIpv6Option(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Docker
// @Summary Update docker daemon.json by upload file
// @Accept json
// @Param request body dto.DaemonJsonUpdateByFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/daemonjson/update/byfile [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新配置文件","formatEN":"Updated configuration file"}
func (b *BaseApi) UpdateDaemonJsonByFile(c *gin.Context) {
	var req dto.DaemonJsonUpdateByFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := dockerService.UpdateConfByFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Docker
// @Summary Operate docker
// @Accept json
// @Param request body dto.DockerOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/docker/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"docker 服务 [operation]","formatEN":"[operation] docker service"}
func (b *BaseApi) OperateDocker(c *gin.Context) {
	var req dto.DockerOperation
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := dockerService.OperateDocker(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
