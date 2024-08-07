package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/gin-gonic/gin"
)

// @Tags System Group
// @Summary Create group
// @Description 创建系统组
// @Accept json
// @Param request body dto.GroupCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /groups [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建组 [name][type]","formatEN":"create group [name][type]"}
func (b *BaseApi) CreateGroup(c *gin.Context) {
	var req dto.GroupCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := groupService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Group
// @Summary Delete group
// @Description 删除系统组
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /groups/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"groups","output_column":"name","output_value":"name"},{"input_column":"id","input_value":"id","isList":false,"db":"groups","output_column":"type","output_value":"type"}],"formatZH":"删除组 [type][name]","formatEN":"delete group [type][name]"}
func (b *BaseApi) DeleteGroup(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := groupService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Group
// @Summary Update group
// @Description 更新系统组
// @Accept json
// @Param request body dto.GroupUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /groups/update [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新组 [name][type]","formatEN":"update group [name][type]"}
func (b *BaseApi) UpdateGroup(c *gin.Context) {
	var req dto.GroupUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := groupService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Group
// @Summary List groups
// @Description 查询系统组
// @Accept json
// @Param request body dto.GroupSearch true "request"
// @Success 200 {array} dto.GroupInfo
// @Security ApiKeyAuth
// @Router /groups/search [post]
func (b *BaseApi) ListGroup(c *gin.Context) {
	var req dto.GroupSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	list, err := groupService.List(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}
