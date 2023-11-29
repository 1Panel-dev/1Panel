package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Website CA
// @Summary Page website ca
// @Description 获取网站 ca 列表分页
// @Accept json
// @Param request body request.WebsiteCASearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /websites/ca/search [post]
func (b *BaseApi) PageWebsiteCA(c *gin.Context) {
	var req request.WebsiteCASearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, cas, err := websiteCAService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: cas,
	})
}

// @Tags Website CA
// @Summary Create website ca
// @Description 创建网站 ca
// @Accept json
// @Param request body request.WebsiteCACreate true "request"
// @Success 200 {object} request.WebsiteCACreate
// @Security ApiKeyAuth
// @Router /websites/ca [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建网站 ca [name]","formatEN":"Create website ca [name]"}
func (b *BaseApi) CreateWebsiteCA(c *gin.Context) {
	var req request.WebsiteCACreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteCAService.Create(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website CA
// @Summary Get website ca
// @Description 获取网站 ca
// @Accept json
// @Param id path int true "id"
// @Success 200 {object} response.WebsiteCADTO
// @Security ApiKeyAuth
// @Router /websites/ca/{id} [get]
func (b *BaseApi) GetWebsiteCA(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		return
	}
	res, err := websiteCAService.GetCA(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website CA
// @Summary Delete website ca
// @Description 删除网站 ca
// @Accept json
// @Param request body request.WebsiteCommonReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ca/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_cas","output_column":"name","output_value":"name"}],"formatZH":"删除网站 ca [name]","formatEN":"Delete website ca [name]"}
func (b *BaseApi) DeleteWebsiteCA(c *gin.Context) {
	var req request.WebsiteCommonReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteCAService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website CA
// @Summary Obtain SSL
// @Description 自签 SSL 证书
// @Accept json
// @Param request body request.WebsiteCAObtain true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ca/obtain [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_cas","output_column":"name","output_value":"name"}],"formatZH":"自签 SSL 证书 [name]","formatEN":"Obtain SSL [name]"}
func (b *BaseApi) ObtainWebsiteCA(c *gin.Context) {
	var req request.WebsiteCAObtain
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if _, err := websiteCAService.ObtainSSL(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website CA
// @Summary Obtain SSL
// @Description 续签 SSL 证书
// @Accept json
// @Param request body request.WebsiteCAObtain true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/ca/obtain [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_cas","output_column":"name","output_value":"name"}],"formatZH":"自签 SSL 证书 [name]","formatEN":"Obtain SSL [name]"}
func (b *BaseApi) RenewWebsiteCA(c *gin.Context) {
	var req request.WebsiteCARenew
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if _, err := websiteCAService.ObtainSSL(request.WebsiteCAObtain{
		SSLID: req.SSLID,
		Renew: true,
		Unit:  "year",
		Time:  1,
	}); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
