package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Host tool
// @Summary Get tool status
// @Accept json
// @Param request body request.HostToolReq true "request"
// @Success 200 {object} response.HostToolRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /host/tool [post]
func (b *BaseApi) GetToolStatus(c *gin.Context) {
	var req request.HostToolReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	config, err := hostToolService.GetToolStatus(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
// @Router /host/tool/create [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 [type] 配置","formatEN":"create [type] config"}
func (b *BaseApi) InitToolConfig(c *gin.Context) {
	var req request.HostToolCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostToolService.CreateToolConfig(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Host tool
// @Summary Operate tool
// @Accept json
// @Param request body request.HostToolReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /host/tool/operate [post]
// @x-panel-log {"bodyKeys":["operate","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] [type] ","formatEN":"[operate] [type]"}
func (b *BaseApi) OperateTool(c *gin.Context) {
	var req request.HostToolReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := hostToolService.OperateTool(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Host tool
// @Summary Get tool config
// @Accept json
// @Param request body request.HostToolConfig true "request"
// @Success 200 {object} response.HostToolConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /host/tool/config [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] 主机工具配置文件 ","formatEN":"[operate] tool config"}
func (b *BaseApi) OperateToolConfig(c *gin.Context) {
	var req request.HostToolConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	config, err := hostToolService.OperateToolConfig(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, config)
}

// @Tags Host tool
// @Summary Get tool logs
// @Accept json
// @Param request body request.HostToolLogReq true "request"
// @Success 200 {string} logContent
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /host/tool/log [post]
func (b *BaseApi) GetToolLog(c *gin.Context) {
	var req request.HostToolLogReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	logContent, err := hostToolService.GetToolLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, logContent)
}

// @Tags Host tool
// @Summary Create Supervisor process
// @Accept json
// @Param request body request.SupervisorProcessConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /host/tool/supervisor/process [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] 守护进程 ","formatEN":"[operate] process"}
func (b *BaseApi) OperateProcess(c *gin.Context) {
	var req request.SupervisorProcessConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err := hostToolService.OperateSupervisorProcess(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Host tool
// @Summary Get Supervisor process config
// @Accept json
// @Success 200 {object} response.SupervisorProcessConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /host/tool/supervisor/process [get]
func (b *BaseApi) GetProcess(c *gin.Context) {
	configs, err := hostToolService.GetSupervisorProcessConfig()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, configs)
}

// @Tags Host tool
// @Summary Get Supervisor process config
// @Accept json
// @Param request body request.SupervisorProcessFileReq true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /host/tool/supervisor/process/file [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] Supervisor 进程文件 ","formatEN":"[operate] Supervisor Process Config file"}
func (b *BaseApi) GetProcessFile(c *gin.Context) {
	var req request.SupervisorProcessFileReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content, err := hostToolService.OperateSupervisorProcessFile(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, content)
}
