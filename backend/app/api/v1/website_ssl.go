package v1

import (
	"reflect"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Website SSL
// @Summary Page website ssl
// @Description 获取网站 ssl 列表分页
// @Accept json
// @Param request body request.WebsiteSSLSearch true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ssl/search [post]
func (b *BaseApi) PageWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if !reflect.DeepEqual(req.PageInfo, dto.PageInfo{}) {
		total, accounts, err := websiteSSLService.Page(req)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{
			Total: total,
			Items: accounts,
		})
	} else {
		list, err := websiteSSLService.Search(req)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, list)
	}
}

// @Tags Website SSL
// @Summary Create website ssl
// @Description 创建网站 ssl
// @Accept json
// @Param request body request.WebsiteSSLCreate true "request"
// @Success 200 {object} request.WebsiteSSLCreate
// @Security ApiKeyAuth
// @Router /websites/ssl [post]
// @x-panel-log {"bodyKeys":["primaryDomain"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建网站 ssl [primaryDomain]","formatEN":"Create website ssl [primaryDomain]"}
func (b *BaseApi) CreateWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	res, err := websiteSSLService.Create(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website SSL
// @Summary Reset website ssl
// @Description 重置网站 ssl
// @Accept json
// @Param request body request.WebsiteSSLRenew true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ssl/renew [post]
// @x-panel-log {"bodyKeys":["SSLId"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"SSLId","isList":false,"db":"website_ssls","output_column":"primary_domain","output_value":"domain"}],"formatZH":"重置 ssl [domain]","formatEN":"Renew ssl [domain]"}
func (b *BaseApi) RenewWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLRenew
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := websiteSSLService.Renew(req.SSLID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website SSL
// @Summary Resolve website ssl
// @Description 解析网站 ssl
// @Accept json
// @Param request body request.WebsiteDNSReq true "request"
// @Success 200 {array} response.WebsiteDNSRes
// @Security ApiKeyAuth
// @Router /websites/ssl/resolve [post]
func (b *BaseApi) GetDNSResolve(c *gin.Context) {
	var req request.WebsiteDNSReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	res, err := websiteSSLService.GetDNSResolve(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website SSL
// @Summary Delete website ssl
// @Description 删除网站 ssl
// @Accept json
// @Param request body request.WebsiteResourceReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ssl/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_ssls","output_column":"primary_domain","output_value":"domain"}],"formatZH":"删除 ssl [domain]","formatEN":"Delete ssl [domain]"}
func (b *BaseApi) DeleteWebsiteSSL(c *gin.Context) {
	var req request.WebsiteResourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := websiteSSLService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website SSL
// @Summary Search website ssl by website id
// @Description 通过网站 id 查询 ssl
// @Accept json
// @Param websiteId path integer true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ssl/website/:websiteId [get]
func (b *BaseApi) GetWebsiteSSLByWebsiteId(c *gin.Context) {
	websiteId, err := helper.GetIntParamByKey(c, "websiteId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	websiteSSL, err := websiteSSLService.GetWebsiteSSL(websiteId)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, websiteSSL)
}

// @Tags Website SSL
// @Summary Search website ssl by id
// @Description 通过 id 查询 ssl
// @Accept json
// @Param id path integer true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ssl/:id [get]
func (b *BaseApi) GetWebsiteSSLById(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	websiteSSL, err := websiteSSLService.GetSSL(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, websiteSSL)
}

// @Tags Website SSL
// @Summary Update ssl
// @Description 更新 ssl
// @Accept json
// @Param request body request.WebsiteSSLUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ssl/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_ssls","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新证书设置 [domain]","formatEN":"Update ssl config [domain]"}
func (b *BaseApi) UpdateWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := websiteSSLService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
