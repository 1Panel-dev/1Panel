package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Website Group
// @Summary List website groups
// @Description 获取网站组
// @Success 200 {anrry} model.WebsiteGroup
// @Security ApiKeyAuth
// @Router /websites/groups [get]
func (b *BaseApi) GetWebGroups(c *gin.Context) {
	list, err := websiteGroupService.GetGroups()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Website Group
// @Summary Create website group
// @Description 创建网站组
// @Accept json
// @Param request body request.WebsiteGroupCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/groups [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建网站组 [name]","formatEN":"Create website groups [name]"}
func (b *BaseApi) CreateWebGroup(c *gin.Context) {
	var req request.WebsiteGroupCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := websiteGroupService.CreateGroup(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website Group
// @Summary Update website group
// @Description 更新网站组
// @Accept json
// @Param request body request.WebsiteGroupUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/groups/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新网站组 [name]","formatEN":"Update website groups [name]"}
func (b *BaseApi) UpdateWebGroup(c *gin.Context) {
	var req request.WebsiteGroupUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := websiteGroupService.UpdateGroup(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website Group
// @Summary Delete website group
// @Description 删除网站组
// @Accept json
// @Param request body request.WebsiteResourceReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/groups/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"website_groups","output_colume":"name","output_value":"name"}],"formatZH":"删除网站组 [name]","formatEN":"Delete website group [name]"}
func (b *BaseApi) DeleteWebGroup(c *gin.Context) {
	var req request.WebsiteResourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := websiteGroupService.DeleteGroup(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
