package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Command
// @Summary Create command
// @Description 创建快速命令
// @Accept json
// @Param request body dto.CommandOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts/command [post]
// @x-panel-log {"bodyKeys":["name","command"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建快捷命令 [name][command]","formatEN":"create quick command [name][command]"}
func (b *BaseApi) CreateCommand(c *gin.Context) {
	var req dto.CommandOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := commandService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Command
// @Summary Page commands
// @Description 获取快速命令列表分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /hosts/command/search [post]
func (b *BaseApi) SearchCommand(c *gin.Context) {
	var req dto.SearchCommandWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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

// @Tags Command
// @Summary Tree commands
// @Description 获取快速命令树
// @Accept json
// @Success 200 {Array} dto.CommandTree
// @Security ApiKeyAuth
// @Router /hosts/command/tree [get]
func (b *BaseApi) SearchCommandTree(c *gin.Context) {
	list, err := commandService.SearchForTree()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Command
// @Summary List commands
// @Description 获取快速命令列表
// @Success 200 {object} dto.CommandInfo
// @Security ApiKeyAuth
// @Router /hosts/command [get]
func (b *BaseApi) ListCommand(c *gin.Context) {
	list, err := commandService.List()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Command
// @Summary Delete command
// @Description 删除快速命令
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts/command/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"commands","output_column":"name","output_value":"names"}],"formatZH":"删除快捷命令 [names]","formatEN":"delete quick command [names]"}
func (b *BaseApi) DeleteCommand(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := commandService.Delete(req.Ids); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Command
// @Summary Update command
// @Description 更新快速命令
// @Accept json
// @Param request body dto.CommandOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts/command/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新快捷命令 [name]","formatEN":"update quick command [name]"}
func (b *BaseApi) UpdateCommand(c *gin.Context) {
	var req dto.CommandOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["group_id"] = req.GroupID
	upMap["command"] = req.Command
	if err := commandService.Update(req.ID, upMap); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
