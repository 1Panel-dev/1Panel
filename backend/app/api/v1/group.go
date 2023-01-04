package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// Create group
// @Tags System Group
// @Summary Create group
// @Description 创建系统组
// @Accept json
// @Param request body dto.GroupOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /groups [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建组 [name][type]","formatEN":"create group [name][type]"}
func (b *BaseApi) CreateGroup(c *gin.Context) {
	var req dto.GroupOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := groupService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Delete group
// @Tags System Group
// @Summary Delete group
// @Description 删除系统组
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /groups/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"groups","output_colume":"name","output_value":"name"}],"formatZH":"删除组 [name]","formatEN":"delete group [name]"}
func (b *BaseApi) DeleteGroup(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := groupService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Update group
// @Tags System Group
// @Summary Update group
// @Description 更新系统组
// @Accept json
// @Param request body dto.GroupOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /groups/update [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新组 [name][type]","formatEN":"update group [name][type]"}
func (b *BaseApi) UpdateGroup(c *gin.Context) {
	var req dto.GroupOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := groupService.Update(req.ID, req.Name); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Search group info
// @Tags System Group
// @Summary Search group info
// @Description 查询系统组
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} dto.GroupInfo
// @Security ApiKeyAuth
// @Router /groups/:id [get]
func (b *BaseApi) GetGroupInfo(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	group, err := groupService.GetGroupInfo(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, group)
}

// List group
// @Tags System Group
// @Summary Search group list
// @Description 查询系统组
// @Accept json
// @Param request body dto.GroupSearch true "request"
// @Success 200 {anrry} dto.GroupInfo
// @Security ApiKeyAuth
// @Router /groups/search [post]
func (b *BaseApi) ListGroup(c *gin.Context) {
	var req dto.GroupSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	list, err := groupService.List(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}
