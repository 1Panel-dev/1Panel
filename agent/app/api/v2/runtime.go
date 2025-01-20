package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Runtime
// @Summary List runtimes
// @Accept json
// @Param request body request.RuntimeSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/search [post]
func (b *BaseApi) SearchRuntimes(c *gin.Context) {
	var req request.RuntimeSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, items, err := runtimeService.Page(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: items,
	})
}

// @Tags Runtime
// @Summary Create runtime
// @Accept json
// @Param request body request.RuntimeCreate true "request"
// @Success 200 {object} model.Runtime
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建运行环境 [name]","formatEN":"Create runtime [name]"}
func (b *BaseApi) CreateRuntime(c *gin.Context) {
	var req request.RuntimeCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	ssl, err := runtimeService.Create(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, ssl)
}

// @Tags Website
// @Summary Delete runtime
// @Accept json
// @Param request body request.RuntimeDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除运行环境 [name]","formatEN":"Delete runtime [name]"}
func (b *BaseApi) DeleteRuntime(c *gin.Context) {
	var req request.RuntimeDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := runtimeService.Delete(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Delete runtime
// @Accept json
// @Success 200 {array} dto.AppResource
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /installed/delete/check/:id [get]
func (b *BaseApi) DeleteRuntimeCheck(c *gin.Context) {
	runTimeId, err := helper.GetIntParamByKey(c, "id")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	checkData, err := runtimeService.DeleteCheck(runTimeId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, checkData)
}

// @Tags Runtime
// @Summary Update runtime
// @Accept json
// @Param request body request.RuntimeUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新运行环境 [name]","formatEN":"Update runtime [name]"}
func (b *BaseApi) UpdateRuntime(c *gin.Context) {
	var req request.RuntimeUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := runtimeService.Update(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Get runtime
// @Accept json
// @Param id path string true "request"
// @Success 200 {object} response.RuntimeDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/:id [get]
func (b *BaseApi) GetRuntime(c *gin.Context) {
	id, err := helper.GetIntParamByKey(c, "id")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	res, err := runtimeService.Get(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Runtime
// @Summary Get Node package scripts
// @Accept json
// @Param request body request.NodePackageReq true "request"
// @Success 200 {array} response.PackageScripts
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/node/package [post]
func (b *BaseApi) GetNodePackageRunScript(c *gin.Context) {
	var req request.NodePackageReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := runtimeService.GetNodePackageRunScript(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Runtime
// @Summary Operate runtime
// @Accept json
// @Param request body request.RuntimeOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/operate [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"操作运行环境 [name]","formatEN":"Operate runtime [name]"}
func (b *BaseApi) OperateRuntime(c *gin.Context) {
	var req request.RuntimeOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := runtimeService.OperateRuntime(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Get Node modules
// @Accept json
// @Param request body request.NodeModuleReq true "request"
// @Success 200 {array} response.NodeModule
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/node/modules [post]
func (b *BaseApi) GetNodeModules(c *gin.Context) {
	var req request.NodeModuleReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := runtimeService.GetNodeModules(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Runtime
// @Summary Operate Node modules
// @Accept json
// @Param request body request.NodeModuleReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/node/modules/operate [post]
func (b *BaseApi) OperateNodeModules(c *gin.Context) {
	var req request.NodeModuleOperateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := runtimeService.OperateNodeModules(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Sync runtime status
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Router /runtimes/sync [post]
func (b *BaseApi) SyncStatus(c *gin.Context) {
	err := runtimeService.SyncRuntimeStatus()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Get php runtime extension
// @Accept json
// @Param id path string true "request"
// @Success 200 {object} response.PHPExtensionRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/:id/extensions [get]
func (b *BaseApi) GetRuntimeExtension(c *gin.Context) {
	id, err := helper.GetIntParamByKey(c, "id")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	res, err := runtimeService.GetPHPExtensions(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Runtime
// @Summary Install php extension
// @Accept json
// @Param request body request.PHPExtensionInstallReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/extensions/install [post]
func (b *BaseApi) InstallPHPExtension(c *gin.Context) {
	var req request.PHPExtensionInstallReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := runtimeService.InstallPHPExtension(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary UnInstall php extension
// @Accept json
// @Param request body request.PHPExtensionInstallReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/extensions/uninstall [post]
func (b *BaseApi) UnInstallPHPExtension(c *gin.Context) {
	var req request.PHPExtensionInstallReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := runtimeService.UnInstallPHPExtension(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Load php runtime conf
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.PHPConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/config/:id [get]
func (b *BaseApi) GetPHPConfig(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	data, err := runtimeService.GetPHPConfig(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Runtime
// @Summary Update runtime php conf
// @Accept json
// @Param request body request.PHPConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/config [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[domain] PHP 配置修改","formatEN":"[domain] PHP conf update"}
func (b *BaseApi) UpdatePHPConfig(c *gin.Context) {
	var req request.PHPConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := runtimeService.UpdatePHPConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Update php conf file
// @Accept json
// @Param request body request.PHPFileUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"php 配置修改 [domain]","formatEN":"Nginx conf update [domain]"}
func (b *BaseApi) UpdatePHPFile(c *gin.Context) {
	var req request.PHPFileUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := runtimeService.UpdatePHPConfigFile(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Get php conf file
// @Accept json
// @Param request body request.PHPFileReq true "request"
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/file [post]
func (b *BaseApi) GetPHPConfigFile(c *gin.Context) {
	var req request.PHPFileReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := runtimeService.GetPHPConfigFile(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Runtime
// @Summary Update fpm config
// @Accept json
// @Param request body request.FPMConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/fpm/config [post]
func (b *BaseApi) UpdateFPMConfig(c *gin.Context) {
	var req request.FPMConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := runtimeService.UpdateFPMConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Get fpm config
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} request.FPMConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/fpm/config/:id [get]
func (b *BaseApi) GetFPMConfig(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	data, err := runtimeService.GetFPMConfig(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Runtime
// @Summary Get supervisor process
// @Accept json
// @Param id path integer true "request"
// @Success 200 {array} response.SupervisorProcessConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/supervisor/process/:id [get]
func (b *BaseApi) GetSupervisorProcess(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	data, err := runtimeService.GetSupervisorProcess(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Runtime
// @Summary Operate supervisor process
// @Accept json
// @Param request body request.PHPSupervisorProcessConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/supervisor/process/operate [post]
func (b *BaseApi) OperateSupervisorProcess(c *gin.Context) {
	var req request.PHPSupervisorProcessConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := runtimeService.OperateSupervisorProcess(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Operate supervisor process file
// @Accept json
// @Param request body request.PHPSupervisorProcessFileReq true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/supervisor/process/file/operate [post]
func (b *BaseApi) OperateSupervisorProcessFile(c *gin.Context) {
	var req request.PHPSupervisorProcessFileReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := runtimeService.OperateSupervisorProcessFile(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}
