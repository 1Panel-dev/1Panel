package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// Page website dns account
// @Tags Website DNS
// @Summary Search website dns account with page
// @Description 获取网站 dns 列表分页
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /websites/dns/search [post]
func (b *BaseApi) PageWebsiteDnsAccount(c *gin.Context) {
	var req dto.PageInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	total, accounts, err := websiteDnsAccountService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: accounts,
	})
}

// Create website dns account
// @Tags Website DNS
// @Summary Create website dns account
// @Description 创建网站 dns
// @Accept json
// @Param request body request.WebsiteDnsAccountCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/dns [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建网站 dns [name]","formatEN":"Create website dns [name]"}
func (b *BaseApi) CreateWebsiteDnsAccount(c *gin.Context) {
	var req request.WebsiteDnsAccountCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if _, err := websiteDnsAccountService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Update website dns account
// @Tags Website DNS
// @Summary Update website dns account
// @Description 更新网站 dns
// @Accept json
// @Param request body request.WebsiteDnsAccountUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/dns/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新网站 dns [name]","formatEN":"Update website dns [name]"}
func (b *BaseApi) UpdateWebsiteDnsAccount(c *gin.Context) {
	var req request.WebsiteDnsAccountUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if _, err := websiteDnsAccountService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Delete website dns account
// @Tags Website DNS
// @Summary Delete website dns account
// @Description 删除网站 dns
// @Accept json
// @Param request body request.WebsiteResourceReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/dns/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"website_dns_accounts","output_colume":"name","output_value":"name"}],"formatZH":"删除网站 dns [name]","formatEN":"Delete website dns [name]"}
func (b *BaseApi) DeleteWebsiteDnsAccount(c *gin.Context) {
	var req request.WebsiteResourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := websiteDnsAccountService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
