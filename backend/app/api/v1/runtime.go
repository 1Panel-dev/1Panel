package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

func (b *BaseApi) DeleteRuntimeCheck(c *gin.Context) {
	runTimeId, err := helper.GetIntParamByKey(c, "runTimeId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	checkData, err := runtimeService.DeleteCheck(runTimeId)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
// @Router /runtimes/{id} [get]
func (b *BaseApi) GetRuntime(c *gin.Context) {
	id, err := helper.GetIntParamByKey(c, "id")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	res, err := runtimeService.Get(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Runtime
// @Summary Sync runtime status
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/sync [post]
func (b *BaseApi) SyncStatus(c *gin.Context) {
	err := runtimeService.SyncRuntimeStatus()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
