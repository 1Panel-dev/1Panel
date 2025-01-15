package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags OpenResty
// @Summary Load OpenResty conf
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
// @Router /openresty [get]
func (b *BaseApi) GetNginx(c *gin.Context) {
	fileInfo, err := nginxService.GetNginxConfig()
	if err != nil {
		helper.InternalServer(c, err)
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
// @Router /openresty/scope [post]
func (b *BaseApi) GetNginxConfigByScope(c *gin.Context) {
	var req request.NginxScopeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	params, err := nginxService.GetConfigByScope(req)
	if err != nil {
		helper.InternalServer(c, err)
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
// @Router /openresty/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新 nginx 配置 [domain]","formatEN":"Update nginx conf [domain]"}
func (b *BaseApi) UpdateNginxConfigByScope(c *gin.Context) {
	var req request.NginxConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := nginxService.UpdateConfigByScope(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags OpenResty
// @Summary Load OpenResty status info
// @Success 200 {object} response.NginxStatus
// @Security ApiKeyAuth
// @Router /openresty/status [get]
func (b *BaseApi) GetNginxStatus(c *gin.Context) {
	res, err := nginxService.GetStatus()
	if err != nil {
		helper.InternalServer(c, err)
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
// @Router /openresty/file [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 nginx 配置","formatEN":"Update nginx conf"}
func (b *BaseApi) UpdateNginxFile(c *gin.Context) {
	var req request.NginxConfigFileUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := nginxService.UpdateConfigFile(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags OpenResty
// @Summary Build OpenResty
// @Accept json
// @Param request body request.NginxBuildReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /openresty/build [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"构建 OpenResty","formatEN":"Build OpenResty"}
func (b *BaseApi) BuildNginx(c *gin.Context) {
	var req request.NginxBuildReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := nginxService.Build(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags OpenResty
// @Summary Update OpenResty module
// @Accept json
// @Param request body request.NginxModuleUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /openresty/module/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 OpenResty 模块","formatEN":"Update OpenResty module"}
func (b *BaseApi) UpdateNginxModule(c *gin.Context) {
	var req request.NginxModuleUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := nginxService.UpdateModule(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags OpenResty
// @Summary Get OpenResty modules
// @Success 200 {array} response.NginxModule
// @Security ApiKeyAuth
// @Router /openresty/modules [get]
func (b *BaseApi) GetNginxModules(c *gin.Context) {
	modules, err := nginxService.GetModules()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, modules)
}
