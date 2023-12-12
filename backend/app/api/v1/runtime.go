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
// @Description 获取运行环境列表
// @Accept json
// @Param request body request.RuntimeSearch true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Description 创建运行环境
// @Accept json
// @Param request body request.RuntimeCreate true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Description 删除运行环境
// @Accept json
// @Param request body request.RuntimeDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /runtimes/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除网站 [name]","formatEN":"Delete website [name]"}
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

// @Tags Runtime
// @Summary Update runtime
// @Description 更新运行环境
// @Accept json
// @Param request body request.RuntimeUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Description 获取运行环境
// @Accept json
// @Param id path string true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /runtimes/:id [get]
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
// @Description 获取 Node 项目的 scripts
// @Accept json
// @Param request body request.NodePackageReq true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Description 操作运行环境
// @Accept json
// @Param request body request.RuntimeOperate true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Description 获取 Node 项目的 modules
// @Accept json
// @Param request body request.NodeModuleReq true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Description 操作 Node 项目 modules
// @Accept json
// @Param request body request.NodeModuleReq true "request"
// @Success 200
// @Security ApiKeyAuth
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
