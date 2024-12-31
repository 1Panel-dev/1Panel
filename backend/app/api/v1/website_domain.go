package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Website Domain
// @Summary Delete website domain
// @Accept json
// @Param request body request.WebsiteDomainDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/domains/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_domains","output_column":"domain","output_value":"domain"}],"formatZH":"删除域名 [domain]","formatEN":"Delete domain [domain]"}
func (b *BaseApi) DeleteWebDomain(c *gin.Context) {
	var req request.WebsiteDomainDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.DeleteWebsiteDomain(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website Domain
// @Summary Create website domain
// @Accept json
// @Param request body request.WebsiteDomainCreate true "request"
// @Success 200 {array} model.WebsiteDomain
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/domains [post]
// @x-panel-log {"bodyKeys":["domain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建域名 [domain]","formatEN":"Create domain [domain]"}
func (b *BaseApi) CreateWebDomain(c *gin.Context) {
	var req request.WebsiteDomainCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	domain, err := websiteService.CreateWebsiteDomain(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, domain)
}

// @Tags Website Domain
// @Summary Search website domains by websiteId
// @Accept json
// @Param websiteId path integer true "request"
// @Success 200 {array} model.WebsiteDomain
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/domains/{websiteId} [get]
func (b *BaseApi) GetWebDomains(c *gin.Context) {
	websiteId, err := helper.GetIntParamByKey(c, "websiteId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	list, err := websiteService.GetWebsiteDomain(websiteId)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}
