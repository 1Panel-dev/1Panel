package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags OpenResty
// @Summary Load OpenResty conf
// @Success 200 {object} response.NginxFile
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /openresty [get]
func (b *BaseApi) GetNginx(c *gin.Context) {
	fileInfo, err := nginxService.GetNginxConfig()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, fileInfo)
}

// @Tags OpenResty
// @Summary Load partial OpenResty conf
// @Accept json
// @Param request body request.NginxScopeReq true "request"
// @Success 200 {array} response.NginxParam
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /openresty/scope [post]
func (b *BaseApi) GetNginxConfigByScope(c *gin.Context) {
	var req request.NginxScopeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	params, err := nginxService.GetConfigByScope(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, params)
}

// @Tags OpenResty
// @Summary Update OpenResty conf
// @Accept json
// @Param request body request.NginxConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /openresty/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新 nginx 配置 [domain]","formatEN":"Update nginx conf [domain]"}
func (b *BaseApi) UpdateNginxConfigByScope(c *gin.Context) {
	var req request.NginxConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := nginxService.UpdateConfigByScope(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags OpenResty
// @Summary Load OpenResty status info
// @Success 200 {object} response.NginxStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /openresty/status [get]
func (b *BaseApi) GetNginxStatus(c *gin.Context) {
	res, err := nginxService.GetStatus()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags OpenResty
// @Summary Update OpenResty conf by upload file
// @Accept json
// @Param request body request.NginxConfigFileUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /openresty/file [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 nginx 配置","formatEN":"Update nginx conf"}
func (b *BaseApi) UpdateNginxFile(c *gin.Context) {
	var req request.NginxConfigFileUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := nginxService.UpdateConfigFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags OpenResty
// @Summary Clear OpenResty proxy cache
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /openresty/clear [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清理 Openresty 代理缓存","formatEN":"Clear nginx proxy cache"}
func (b *BaseApi) ClearNginxProxyCache(c *gin.Context) {
	if err := nginxService.ClearProxyCache(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
