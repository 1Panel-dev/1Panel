package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Host tool
// @Summary Get tool status
// @Accept json
// @Param request body request.HostToolTypeReq true "request"
// @Success 200 {object} response.HostToolRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/status [post]
func (b *BaseApi) GetToolStatus(c *gin.Context) {
	var req request.HostToolTypeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	config, err := hostToolService.GetToolStatus(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, config)
}

// @Tags Host tool
// @Summary Create Host tool Config
// @Accept json
// @Param request body request.HostToolCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/init [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 [type] 配置","formatEN":"create [type] config"}
func (b *BaseApi) InitToolConfig(c *gin.Context) {
	var req request.HostToolCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostToolService.CreateToolConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host tool
// @Summary Operate tool
// @Accept json
// @Param request body request.HostToolOperateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/operate [post]
// @x-panel-log {"bodyKeys":["operate","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] [type] ","formatEN":"[operate] [type]"}
func (b *BaseApi) OperateTool(c *gin.Context) {
	var req request.HostToolOperateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := hostToolService.OperateTool(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host tool
// @Summary Get tool config
// @Accept json
// @Param request body request.HostToolTypeReq true "request"
// @Success 200 {object} response.HostToolConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/config/get [post]
func (b *BaseApi) GetToolConfig(c *gin.Context) {
	var req request.HostToolTypeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	config, err := hostToolService.GetToolConfig(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, config)
}

// @Tags Host tool
// @Summary Update tool config
// @Accept json
// @Param request body request.HostToolConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/config/set [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 [type] 主机工具配置文件 ","formatEN":"update [type] tool config"}
func (b *BaseApi) UpdateToolConfig(c *gin.Context) {
	var req request.HostToolConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostToolService.UpdateToolConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host tool
// @Summary Create Supervisor process
// @Accept json
// @Param request body request.SupervisorProcessConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/supervisor/process [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] 守护进程 ","formatEN":"[operate] process"}
func (b *BaseApi) OperateProcess(c *gin.Context) {
	var req request.SupervisorProcessConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err := hostToolService.OperateSupervisorProcess(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host tool
// @Summary Get Supervisor process config
// @Accept json
// @Success 200 {object} response.SupervisorProcessConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/supervisor/process [get]
func (b *BaseApi) GetProcess(c *gin.Context) {
	configs, err := hostToolService.GetSupervisorProcessConfig()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, configs)
}

// @Tags Host tool
// @Summary Get Supervisor process config file
// @Accept json
// @Param request body request.HostSupervisorProcessFileGetReq true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/supervisor/process/file/get [post]
func (b *BaseApi) GetProcessFile(c *gin.Context) {
	var req request.HostSupervisorProcessFileGetReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content, err := hostToolService.GetSupervisorProcessFile(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, content)
}

// @Tags Host tool
// @Summary Operate Supervisor process config file
// @Accept json
// @Param request body request.HostSupervisorProcessFileOperateReq true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tool/supervisor/process/file [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] Supervisor 进程文件 ","formatEN":"[operate] Supervisor Process Config file"}
func (b *BaseApi) OperateProcessFile(c *gin.Context) {
	var req request.HostSupervisorProcessFileOperateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content, err := hostToolService.OperateSupervisorProcessFile(request.SupervisorProcessFileReq{
		Name:    req.Name,
		Operate: req.Operate,
		Content: req.Content,
		File:    req.File,
	})
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, content)
}
