package v1

import (
	"os"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags Container Docker
// @Summary Load docker status
// @Description 获取 docker 服务状态
// @Produce json
// @Success 200 {string} status
// @Security ApiKeyAuth
// @Router /containers/docker/status [get]
func (b *BaseApi) LoadDockerStatus(c *gin.Context) {
	status := dockerService.LoadDockerStatus()
	helper.SuccessWithData(c, status)
}

// @Tags Container Docker
// @Summary Load docker daemon.json
// @Description 获取 docker 配置信息(表单)
// @Produce json
// @Success 200 {object} string
// @Security ApiKeyAuth
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
// @Description 获取 docker 配置信息
// @Produce json
// @Success 200 {object} dto.DaemonJsonConf
// @Security ApiKeyAuth
// @Router /containers/daemonjson [get]
func (b *BaseApi) LoadDaemonJson(c *gin.Context) {
	conf := dockerService.LoadDockerConf()
	helper.SuccessWithData(c, conf)
}

// @Tags Container Docker
// @Summary Update docker daemon.json
// @Description 修改 docker 配置信息
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/daemonjson/update [post]
// @x-panel-log {"bodyKeys":["key", "value"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新 docker daemon.json 配置 [key]=>[value]","formatEN":"Updated the docker daemon.json configuration [key]=>[value]"}
func (b *BaseApi) UpdateDaemonJson(c *gin.Context) {
	var req dto.SettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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
// @Description 修改 docker 日志配置
// @Accept json
// @Param request body dto.LogOption true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/daemonjson/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新 docker daemon.json 日志配置","formatEN":"Updated the docker daemon.json log option"}
func (b *BaseApi) UpdateLogOption(c *gin.Context) {
	var req dto.LogOption
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := dockerService.UpdateLogOption(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Docker
// @Summary Update docker daemon.json by upload file
// @Description 上传替换 docker 配置文件
// @Accept json
// @Param request body dto.DaemonJsonUpdateByFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/daemonjson/update/byfile [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新 docker daemon.json 配置","formatEN":"Updated the docker daemon.json configuration"}
func (b *BaseApi) UpdateDaemonJsonByFile(c *gin.Context) {
	var req dto.DaemonJsonUpdateByFile
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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
// @Description Docker 操作
// @Accept json
// @Param request body dto.DockerOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/docker/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"docker 服务 [operation]","formatEN":"[operation] docker service"}
func (b *BaseApi) OperateDocker(c *gin.Context) {
	var req dto.DockerOperation
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := dockerService.OperateDocker(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
