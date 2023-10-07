package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags Database
// @Summary Create database
// @Description 创建远程数据库
// @Accept json
// @Param request body dto.DatabaseCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/db [post]
// @x-panel-log {"bodyKeys":["name", "type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建远程数据库 [name][type]","formatEN":"create database [name][type]"}
func (b *BaseApi) CreateDatabase(c *gin.Context) {
	var req dto.DatabaseCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := databaseService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database
// @Summary Check database
// @Description 检测远程数据库连接性
// @Accept json
// @Param request body dto.DatabaseCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/db/check [post]
// @x-panel-log {"bodyKeys":["name", "type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"检测远程数据库 [name][type] 连接性","formatEN":"check if database [name][type] is connectable"}
func (b *BaseApi) CheckDatabase(c *gin.Context) {
	var req dto.DatabaseCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	helper.SuccessWithData(c, databaseService.CheckDatabase(req))
}

// @Tags Database
// @Summary Page databases
// @Description 获取远程数据库列表分页
// @Accept json
// @Param request body dto.DatabaseSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /databases/db/search [post]
func (b *BaseApi) SearchDatabase(c *gin.Context) {
	var req dto.DatabaseSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := databaseService.SearchWithPage(req)
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
// @Summary List databases
// @Description 获取远程数据库列表
// @Success 200 {array} dto.DatabaseOption
// @Security ApiKeyAuth
// @Router /databases/db/list/:type [get]
func (b *BaseApi) ListDatabase(c *gin.Context) {
	dbType, err := helper.GetStrParamByKey(c, "type")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	list, err := databaseService.List(dbType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Database
// @Summary Get databases
// @Description 获取远程数据库
// @Success 200 {object} dto.DatabaseInfo
// @Security ApiKeyAuth
// @Router /databases/db/:name [get]
func (b *BaseApi) GetDatabase(c *gin.Context) {
	name, err := helper.GetStrParamByKey(c, "name")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	data, err := databaseService.Get(name)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database
// @Summary Delete database
// @Description 删除远程数据库
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/db/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"databases","output_column":"name","output_value":"names"}],"formatZH":"删除远程数据库 [names]","formatEN":"delete database [names]"}
func (b *BaseApi) DeleteDatabase(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := databaseService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database
// @Summary Update database
// @Description 更新远程数据库
// @Accept json
// @Param request body dto.DatabaseUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/db/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新远程数据库 [name]","formatEN":"update database [name]"}
func (b *BaseApi) UpdateDatabase(c *gin.Context) {
	var req dto.DatabaseUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := databaseService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
