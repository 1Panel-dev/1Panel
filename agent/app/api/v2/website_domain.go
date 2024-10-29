package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Website Domain
// @Summary Delete website domain
// @Description 删除网站域名
// @Accept json
// @Param request body request.WebsiteDomainDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/domains/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_domains","output_column":"domain","output_value":"domain"}],"formatZH":"删除域名 [domain]","formatEN":"Delete domain [domain]"}
func (b *BaseApi) DeleteWebDomain(c *gin.Context) {
	var req request.WebsiteDomainDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.DeleteWebsiteDomain(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website Domain
// @Summary Create website domain
// @Description 创建网站域名
// @Accept json
// @Param request body request.WebsiteDomainCreate true "request"
// @Success 200 {object} model.WebsiteDomain
// @Security ApiKeyAuth
// @Router /websites/domains [post]
// @x-panel-log {"bodyKeys":["domain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建域名 [domain]","formatEN":"Create domain [domain]"}
func (b *BaseApi) CreateWebDomain(c *gin.Context) {
	var req request.WebsiteDomainCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	domain, err := websiteService.CreateWebsiteDomain(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, domain)
}

// @Tags Website Domain
// @Summary Search website domains by websiteId
// @Description 通过网站 id 查询域名
// @Accept json
// @Param websiteId path integer true "request"
// @Success 200 {array} model.WebsiteDomain
// @Security ApiKeyAuth
// @Router /websites/domains/:websiteId [get]
func (b *BaseApi) GetWebDomains(c *gin.Context) {
	websiteId, err := helper.GetIntParamByKey(c, "websiteId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	list, err := websiteService.GetWebsiteDomain(websiteId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// 写一个 update website domain 的接口
// @Tags Website Domain
// @Summary Update website domain
// @Description 更新网站域名
// @Accept json
// @Param request body request.WebsiteDomainUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/domains/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_domains","output_column":"domain","output_value":"domain"}],"formatZH":"更新域名 [domain]","formatEN":"Update domain [domain]"}
func (b *BaseApi) UpdateWebDomain(c *gin.Context) {
	var req request.WebsiteDomainUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateWebsiteDomain(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}
