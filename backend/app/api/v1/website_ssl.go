package v1

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Website SSL
// @Summary Page website ssl
// @Accept json
// @Param request body request.WebsiteSSLSearch true "request"
// @Success 200 {array} response.WebsiteSSLDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/search [post]
func (b *BaseApi) PageWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLSearch
	if err := helper.CheckBind(&req, c); err != nil {
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
// @Accept json
// @Param request body request.WebsiteSSLCreate true "request"
// @Success 200 {object} request.WebsiteSSLCreate
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl [post]
// @x-panel-log {"bodyKeys":["primaryDomain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建网站 ssl [primaryDomain]","formatEN":"Create website ssl [primaryDomain]"}
func (b *BaseApi) CreateWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Summary Apply  ssl
// @Accept json
// @Param request body request.WebsiteSSLApply true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/obtain [post]
// @x-panel-log {"bodyKeys":["ID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ID","isList":false,"db":"website_ssls","output_column":"primary_domain","output_value":"domain"}],"formatZH":"申请证书  [domain]","formatEN":"apply ssl [domain]"}
func (b *BaseApi) ApplyWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLApply
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteSSLService.ObtainSSL(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website SSL
// @Summary Resolve website ssl
// @Accept json
// @Param request body request.WebsiteDNSReq true "request"
// @Success 200 {array} response.WebsiteDNSRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/resolve [post]
func (b *BaseApi) GetDNSResolve(c *gin.Context) {
	var req request.WebsiteDNSReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Accept json
// @Param request body request.WebsiteBatchDelReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"website_ssls","output_column":"primary_domain","output_value":"domain"}],"formatZH":"删除 ssl [domain]","formatEN":"Delete ssl [domain]"}
func (b *BaseApi) DeleteWebsiteSSL(c *gin.Context) {
	var req request.WebsiteBatchDelReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteSSLService.Delete(req.IDs); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website SSL
// @Summary Search website ssl by website id
// @Accept json
// @Param websiteId path integer true "request"
// @Success 200 {object} response.WebsiteSSLDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/website/{websiteId} [get]
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
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.WebsiteSSLDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/{id} [get]
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
// @Summary Update Website ssl
// @Accept json
// @Param request body request.WebsiteSSLUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_ssls","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新证书设置 [domain]","formatEN":"Update ssl config [domain]"}
func (b *BaseApi) UpdateWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteSSLService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website SSL
// @Summary Upload Website ssl
// @Accept json
// @Param request body request.WebsiteSSLUpload true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ssl/upload [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"上传 ssl [type]","formatEN":"Upload ssl [type]"}
func (b *BaseApi) UploadWebsiteSSL(c *gin.Context) {
	var req request.WebsiteSSLUpload
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteSSLService.Upload(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website SSL
// @Summary Download SSL file
// @Accept json
// @Param request body request.WebsiteResourceReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router  /websites/ssl/download [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_ssls","output_column":"primary_domain","output_value":"domain"}],"formatZH":"下载证书文件 [domain]","formatEN":"download ssl file [domain]"}
func (b *BaseApi) DownloadWebsiteSSL(c *gin.Context) {
	var req request.WebsiteResourceReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	file, err := websiteSSLService.DownloadFile(req.ID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(info.Name()))
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
}
