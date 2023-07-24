package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags Database
// @Summary Create remote database
// @Description 创建远程数据库
// @Accept json
// @Param request body dto.RemoteDBCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/remote [post]
// @x-panel-log {"bodyKeys":["name", "type"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建远程数据库 [name][type]","formatEN":"create remote database [name][type]"}
func (b *BaseApi) CreateRemoteDB(c *gin.Context) {
	var req dto.RemoteDBCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := remoteDBService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database
// @Summary Page remote databases
// @Description 获取远程数据库列表分页
// @Accept json
// @Param request body dto.RemoteDBSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /databases/remote/search [post]
func (b *BaseApi) SearchRemoteDB(c *gin.Context) {
	var req dto.RemoteDBSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := remoteDBService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Database
// @Summary List remote databases
// @Description 获取快速命令列表
// @Success 200 {array} dto.RemoteDBOption
// @Security ApiKeyAuth
// @Router /databases/remote/list/:type [get]
func (b *BaseApi) ListRemoteDB(c *gin.Context) {
	dbType, err := helper.GetStrParamByKey(c, "type")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	list, err := remoteDBService.List(dbType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Database
// @Summary Delete remote database
// @Description 删除远程数据库
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/remote/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"ids","isList":true,"db":"databases","output_column":"name","output_value":"names"}],"formatZH":"删除远程数据库 [names]","formatEN":"delete remote database [names]"}
func (b *BaseApi) DeleteRemoteDB(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := remoteDBService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database
// @Summary Update remote database
// @Description 更新远程数据库
// @Accept json
// @Param request body dto.RemoteDBUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/remote/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新远程数据库 [name]","formatEN":"update remote database [name]"}
func (b *BaseApi) UpdateRemoteDB(c *gin.Context) {
	var req dto.RemoteDBUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := remoteDBService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
