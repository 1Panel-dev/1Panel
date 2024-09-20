package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/gin-gonic/gin"
)

// @Tags OpenResty
// @Summary Load OpenResty conf
// @Description 获取 OpenResty 配置信息
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
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
// @Description 获取部分 OpenResty 配置信息
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, params)
}

// @Tags OpenResty
// @Summary Update OpenResty conf
// @Description 更新 OpenResty 配置信息
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags OpenResty
// @Summary Load OpenResty status info
// @Description 获取 OpenResty 状态信息
// @Success 200 {object} response.NginxStatus
// @Security ApiKeyAuth
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
// @Description 上传更新 OpenResty 配置文件
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags OpenResty
// @Summary Clear OpenResty proxy cache
// @Description 清理 OpenResty 代理缓存
// @Success 200
// @Security ApiKeyAuth
// @Router /openresty/clear [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清理 Openresty 代理缓存","formatEN":"Clear nginx proxy cache"}
func (b *BaseApi) ClearNginxProxyCache(c *gin.Context) {
	if err := nginxService.ClearProxyCache(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags OpenResty
// @Summary Build OpenResty
// @Description 构建 OpenResty
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags OpenResty
// @Summary Update OpenResty module
// @Description 更新 OpenResty 模块
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags OpenResty
// @Summary Get OpenResty modules
// @Description 获取 OpenResty 模块
// @Success 200 {array} response.NginxModule
// @Security ApiKeyAuth
// @Router /openresty/modules [get]
func (b *BaseApi) GetNginxModules(c *gin.Context) {
	modules, err := nginxService.GetModules()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, modules)
}
