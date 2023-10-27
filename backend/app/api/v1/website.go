package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Website
// @Summary Page websites
// @Description 获取网站列表分页
// @Accept json
// @Param request body request.WebsiteSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /websites/search [post]
func (b *BaseApi) PageWebsite(c *gin.Context) {
	var req request.WebsiteSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, websites, err := websiteService.PageWebsite(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: websites,
	})
}

// @Tags Website
// @Summary List websites
// @Description 获取网站列表
// @Success 200 {array} response.WebsiteDTO
// @Security ApiKeyAuth
// @Router /websites/list [get]
func (b *BaseApi) GetWebsites(c *gin.Context) {
	websites, err := websiteService.GetWebsites()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, websites)
}

// @Tags Website
// @Summary List website names
// @Description 获取网站列表
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Router /websites/options [get]
func (b *BaseApi) GetWebsiteOptions(c *gin.Context) {
	websites, err := websiteService.GetWebsiteOptions()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, websites)
}

// @Tags Website
// @Summary Create website
// @Description 创建网站
// @Accept json
// @Param request body request.WebsiteCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites [post]
// @x-panel-log {"bodyKeys":["primaryDomain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建网站 [primaryDomain]","formatEN":"Create website [primaryDomain]"}
func (b *BaseApi) CreateWebsite(c *gin.Context) {
	var req request.WebsiteCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.CreateWebsite(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Operate website
// @Description 操作网站
// @Accept json
// @Param request body request.WebsiteOp true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/operate [post]
// @x-panel-log {"bodyKeys":["id", "operate"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[operate] 网站 [domain]","formatEN":"[operate] website [domain]"}
func (b *BaseApi) OpWebsite(c *gin.Context) {
	var req request.WebsiteOp
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.OpWebsite(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Delete website
// @Description 删除网站
// @Accept json
// @Param request body request.WebsiteDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"删除网站 [domain]","formatEN":"Delete website [domain]"}
func (b *BaseApi) DeleteWebsite(c *gin.Context) {
	var req request.WebsiteDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.DeleteWebsite(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Update website
// @Description 更新网站
// @Accept json
// @Param request body request.WebsiteUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/update [post]
// @x-panel-log {"bodyKeys":["primaryDomain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新网站 [primaryDomain]","formatEN":"Update website [primaryDomain]"}
func (b *BaseApi) UpdateWebsite(c *gin.Context) {
	var req request.WebsiteUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateWebsite(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Search website by id
// @Description 通过 id 查询网站
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.WebsiteDTO
// @Security ApiKeyAuth
// @Router /websites/:id [get]
func (b *BaseApi) GetWebsite(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	website, err := websiteService.GetWebsite(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, website)
}

// @Tags Website Nginx
// @Summary Search website nginx by id
// @Description 通过 id 查询网站 nginx
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
// @Router /websites/:id/config/:type [get]
func (b *BaseApi) GetWebsiteNginx(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	configType := c.Param("type")

	fileInfo, err := websiteService.GetWebsiteNginxConfig(id, configType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, fileInfo)
}

// @Tags Website Nginx
// @Summary Load nginx conf
// @Description 获取 nginx 配置
// @Accept json
// @Param request body request.NginxScopeReq true "request"
// @Success 200 {object} response.WebsiteNginxConfig
// @Security ApiKeyAuth
// @Router /websites/config [post]
func (b *BaseApi) GetNginxConfig(c *gin.Context) {
	var req request.NginxScopeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	config, err := websiteService.GetNginxConfigByScope(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, config)
}

// @Tags Website Nginx
// @Summary Update nginx conf
// @Description 更新 nginx 配置
// @Accept json
// @Param request body request.NginxConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/config/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"nginx 配置修改 [domain]","formatEN":"Nginx conf update [domain]"}
func (b *BaseApi) UpdateNginxConfig(c *gin.Context) {
	var req request.NginxConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateNginxConfigByScope(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website HTTPS
// @Summary Load https conf
// @Description 获取 https 配置
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.WebsiteHTTPS
// @Security ApiKeyAuth
// @Router /websites/:id/https [get]
func (b *BaseApi) GetHTTPSConfig(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	res, err := websiteService.GetWebsiteHTTPS(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website HTTPS
// @Summary Update https conf
// @Description 更新 https 配置
// @Accept json
// @Param request body request.WebsiteHTTPSOp true "request"
// @Success 200 {object} response.WebsiteHTTPS
// @Security ApiKeyAuth
// @Router /websites/:id/https [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新网站 [domain] https 配置","formatEN":"Update website https [domain] conf"}
func (b *BaseApi) UpdateHTTPSConfig(c *gin.Context) {
	var req request.WebsiteHTTPSOp
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	tx, ctx := helper.GetTxAndContext()
	res, err := websiteService.OpWebsiteHTTPS(ctx, req)
	if err != nil {
		tx.Rollback()
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	tx.Commit()
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Check before create website
// @Description 网站创建前检查
// @Accept json
// @Param request body request.WebsiteInstallCheckReq true "request"
// @Success 200 {array} response.WebsitePreInstallCheck
// @Security ApiKeyAuth
// @Router /websites/check [post]
func (b *BaseApi) CreateWebsiteCheck(c *gin.Context) {
	var req request.WebsiteInstallCheckReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := websiteService.PreInstallCheck(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Website WAF
// @Summary Load websit waf conf
// @Description 获取网站 waf 配置
// @Accept json
// @Param request body request.WebsiteWafReq true "request"
// @Success 200 {object} response.WebsiteWafConfig
// @Security ApiKeyAuth
// @Router /websites/waf/config [post]
func (b *BaseApi) GetWebsiteWafConfig(c *gin.Context) {
	var req request.WebsiteWafReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := websiteService.GetWafConfig(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Website WAF
// @Summary Update website waf conf
// @Description 更新 网站 waf 配置
// @Accept json
// @Param request body request.WebsiteWafUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/waf/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"WAF 配置修改 [domain]","formatEN":"WAF conf update [domain]"}
func (b *BaseApi) UpdateWebsiteWafConfig(c *gin.Context) {
	var req request.WebsiteWafUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateWafConfig(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website WAF
// @Summary Update website waf  file
// @Description 更新 网站 waf 配置文件
// @Accept json
// @Param request body request.WebsiteWafUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/waf/file/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"WAF 配置文件修改 [domain]","formatEN":"WAF conf file update [domain]"}
func (b *BaseApi) UpdateWebsiteWafFile(c *gin.Context) {
	var req request.WebsiteWafFileUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateWafFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website Nginx
// @Summary Update website nginx conf
// @Description 更新 网站 nginx 配置
// @Accept json
// @Param request body request.WebsiteNginxUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/nginx/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[domain] Nginx 配置修改","formatEN":"[domain] Nginx conf update"}
func (b *BaseApi) UpdateWebsiteNginxConfig(c *gin.Context) {
	var req request.WebsiteNginxUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateNginxConfigFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Operate website log
// @Description 操作网站日志
// @Accept json
// @Param request body request.WebsiteLogReq true "request"
// @Success 200 {object} response.WebsiteLog
// @Security ApiKeyAuth
// @Router /websites/log [post]
// @x-panel-log {"bodyKeys":["id", "operate"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[domain][operate] 日志","formatEN":"[domain][operate] logs"}
func (b *BaseApi) OpWebsiteLog(c *gin.Context) {
	var req request.WebsiteLogReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.OpWebsiteLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Change default server
// @Description 操作网站日志
// @Accept json
// @Param request body request.WebsiteDefaultUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/default/server [post]
// @x-panel-log {"bodyKeys":["id", "operate"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"修改默认 server => [domain]","formatEN":"Change default server => [domain]"}
func (b *BaseApi) ChangeDefaultServer(c *gin.Context) {
	var req request.WebsiteDefaultUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.ChangeDefaultServer(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Load websit php conf
// @Description 获取网站 php 配置
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.PHPConfig
// @Security ApiKeyAuth
// @Router /websites/php/config/:id [get]
func (b *BaseApi) GetWebsitePHPConfig(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	data, err := websiteService.GetPHPConfig(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Website PHP
// @Summary Update website php conf
// @Description 更新 网站 PHP 配置
// @Accept json
// @Param request body request.WebsitePHPConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/php/config [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[domain] PHP 配置修改","formatEN":"[domain] PHP conf update"}
func (b *BaseApi) UpdateWebsitePHPConfig(c *gin.Context) {
	var req request.WebsitePHPConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdatePHPConfig(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website PHP
// @Summary Update php conf
// @Description 更新 php 配置文件
// @Accept json
// @Param request body request.WebsitePHPFileUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/php/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"php 配置修改 [domain]","formatEN":"Nginx conf update [domain]"}
func (b *BaseApi) UpdatePHPFile(c *gin.Context) {
	var req request.WebsitePHPFileUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdatePHPConfigFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website PHP
// @Summary Update php version
// @Description 变更 php 版本
// @Accept json
// @Param request body request.WebsitePHPVersionReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/php/version [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"php 版本变更 [domain]","formatEN":"php version update [domain]"}
func (b *BaseApi) ChangePHPVersion(c *gin.Context) {
	var req request.WebsitePHPVersionReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.ChangePHPVersion(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get rewrite conf
// @Description 获取伪静态配置
// @Accept json
// @Param request body request.NginxRewriteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/rewrite [post]
func (b *BaseApi) GetRewriteConfig(c *gin.Context) {
	var req request.NginxRewriteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetRewriteConfig(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update rewrite conf
// @Description 更新伪静态配置
// @Accept json
// @Param request body request.NginxRewriteUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/rewrite/update [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"伪静态配置修改 [domain]","formatEN":"Nginx conf rewrite update [domain]"}
func (b *BaseApi) UpdateRewriteConfig(c *gin.Context) {
	var req request.NginxRewriteUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateRewriteConfig(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Update Site Dir
// @Description 更新网站目录
// @Accept json
// @Param request body request.WebsiteUpdateDir true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/dir/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新网站 [domain] 目录","formatEN":"Update  domain [domain] dir"}
func (b *BaseApi) UpdateSiteDir(c *gin.Context) {
	var req request.WebsiteUpdateDir
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateSiteDir(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update Site Dir permission
// @Description 更新网站目录权限
// @Accept json
// @Param request body request.WebsiteUpdateDirPermission true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/dir/permission [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新网站 [domain] 目录权限","formatEN":"Update  domain [domain] dir permission"}
func (b *BaseApi) UpdateSiteDirPermission(c *gin.Context) {
	var req request.WebsiteUpdateDirPermission
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateSitePermission(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get proxy conf
// @Description 获取反向代理配置
// @Accept json
// @Param request body request.WebsiteProxyReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/proxies [post]
func (b *BaseApi) GetProxyConfig(c *gin.Context) {
	var req request.WebsiteProxyReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetProxies(req.ID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update proxy conf
// @Description 修改反向代理配置
// @Accept json
// @Param request body request.WebsiteProxyConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/proxies/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"修改网站 [domain] 反向代理配置 ","formatEN":"Update domain [domain] proxy config"}
func (b *BaseApi) UpdateProxyConfig(c *gin.Context) {
	var req request.WebsiteProxyConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.OperateProxy(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update proxy file
// @Description 更新反向代理文件
// @Accept json
// @Param request body request.NginxProxyUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/proxy/file [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新反向代理文件 [domain]","formatEN":"Nginx conf proxy file update [domain]"}
func (b *BaseApi) UpdateProxyConfigFile(c *gin.Context) {
	var req request.NginxProxyUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateProxyFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get AuthBasic conf
// @Description 获取密码访问配置
// @Accept json
// @Param request body request.NginxAuthReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/auths [post]
func (b *BaseApi) GetAuthConfig(c *gin.Context) {
	var req request.NginxAuthReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetAuthBasics(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Get AuthBasic conf
// @Description 更新密码访问配置
// @Accept json
// @Param request body request.NginxAuthUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/auths/update [post]
func (b *BaseApi) UpdateAuthConfig(c *gin.Context) {
	var req request.NginxAuthUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateAuthBasic(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get AntiLeech conf
// @Description 获取防盗链配置
// @Accept json
// @Param request body request.NginxCommonReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/leech [post]
func (b *BaseApi) GetAntiLeech(c *gin.Context) {
	var req request.NginxCommonReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetAntiLeech(req.WebsiteID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update AntiLeech
// @Description 更新防盗链配置
// @Accept json
// @Param request body request.NginxAntiLeechUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/leech/update [post]
func (b *BaseApi) UpdateAntiLeech(c *gin.Context) {
	var req request.NginxAntiLeechUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateAntiLeech(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update redirect conf
// @Description 修改重定向配置
// @Accept json
// @Param request body request.NginxRedirectReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/redirect/update [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"修改网站 [domain] 重定向理配置 ","formatEN":"Update domain [domain] redirect config"}
func (b *BaseApi) UpdateRedirectConfig(c *gin.Context) {
	var req request.NginxRedirectReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.OperateRedirect(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get redirect conf
// @Description 获取重定向配置
// @Accept json
// @Param request body request.WebsiteProxyReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/redirect [post]
func (b *BaseApi) GetRedirectConfig(c *gin.Context) {
	var req request.WebsiteRedirectReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetRedirect(req.WebsiteID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update redirect file
// @Description 更新重定向文件
// @Accept json
// @Param request body request.NginxRedirectUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/redirect/file [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新重定向文件 [domain]","formatEN":"Nginx conf redirect file update [domain]"}
func (b *BaseApi) UpdateRedirectConfigFile(c *gin.Context) {
	var req request.NginxRedirectUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateRedirectFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get website dir
// @Description 获取网站目录配置
// @Accept json
// @Param request body request.WebsiteCommonReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /websites/dir [post]
func (b *BaseApi) GetDirConfig(c *gin.Context) {
	var req request.WebsiteCommonReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.LoadWebsiteDirConfig(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}
