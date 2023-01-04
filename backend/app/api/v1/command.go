package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// Create command
// @Tags Command
// @Summary Create command
// @Description 创建快速命令
// @Accept json
// @Param request body dto.CommandOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /commands [post]
// @x-panel-log {"bodyKeys":["name","command"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建快捷命令 [name][command]","formatEN":"create quick command [name][command]"}
func (b *BaseApi) CreateCommand(c *gin.Context) {
	var req dto.CommandOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := commandService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Page command
// @Tags Command
// @Summary Search command with page
// @Description 获取快速命令列表分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /commands [post]
func (b *BaseApi) SearchCommand(c *gin.Context) {
	var req dto.SearchWithPage
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := commandService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// List command
// @Tags Command
// @Summary Search command
// @Description 获取快速命令列表
// @Success 200 {object} dto.CommandInfo
// @Security ApiKeyAuth
// @Router /commands [get]
func (b *BaseApi) ListCommand(c *gin.Context) {
	list, err := commandService.List()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// Delete command
// @Tags Command
// @Summary Delete command
// @Description 删除快速命令
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /commands/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"ids","isList":true,"db":"commands","output_colume":"name","output_value":"names"}],"formatZH":"删除快捷命令 [names]","formatEN":"delete quick command [names]"}
func (b *BaseApi) DeleteCommand(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := commandService.Delete(req.Ids); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Update command
// @Tags Command
// @Summary Update command
// @Description 更新快速命令
// @Accept json
// @Param request body dto.CommandOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /commands/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新快捷命令 [name]","formatEN":"update quick command [name]"}
func (b *BaseApi) UpdateCommand(c *gin.Context) {
	var req dto.CommandOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["command"] = req.Command
	if err := commandService.Update(req.ID, upMap); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
