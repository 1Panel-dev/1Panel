package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Website
// @Summary Page websites
// @Accept json
// @Param request body request.WebsiteSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/search [post]
func (b *BaseApi) PageWebsite(c *gin.Context) {
	var req request.WebsiteSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, websites, err := websiteService.PageWebsite(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: websites,
	})
}

// @Tags Website
// @Summary List websites
// @Success 200 {array} response.WebsiteDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/list [get]
func (b *BaseApi) GetWebsites(c *gin.Context) {
	websites, err := websiteService.GetWebsites()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, websites)
}

// @Tags Website
// @Summary List website names
// @Success 200 {array} response.WebsiteOption
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/options [post]
func (b *BaseApi) GetWebsiteOptions(c *gin.Context) {
	var req request.WebsiteOptionReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	websites, err := websiteService.GetWebsiteOptions(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, websites)
}

// @Tags Website
// @Summary Create website
// @Accept json
// @Param request body request.WebsiteCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites [post]
// @x-panel-log {"bodyKeys":["primaryDomain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建网站 [primaryDomain]","formatEN":"Create website [primaryDomain]"}
func (b *BaseApi) CreateWebsite(c *gin.Context) {
	var req request.WebsiteCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.CreateWebsite(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Operate website
// @Accept json
// @Param request body request.WebsiteOp true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/operate [post]
// @x-panel-log {"bodyKeys":["id", "operate"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[operate] 网站 [domain]","formatEN":"[operate] website [domain]"}
func (b *BaseApi) OpWebsite(c *gin.Context) {
	var req request.WebsiteOp
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.OpWebsite(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Delete website
// @Accept json
// @Param request body request.WebsiteDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"删除网站 [domain]","formatEN":"Delete website [domain]"}
func (b *BaseApi) DeleteWebsite(c *gin.Context) {
	var req request.WebsiteDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.DeleteWebsite(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update website
// @Accept json
// @Param request body request.WebsiteUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/update [post]
// @x-panel-log {"bodyKeys":["primaryDomain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新网站 [primaryDomain]","formatEN":"Update website [primaryDomain]"}
func (b *BaseApi) UpdateWebsite(c *gin.Context) {
	var req request.WebsiteUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateWebsite(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Search website by id
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.WebsiteDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/:id [get]
func (b *BaseApi) GetWebsite(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	website, err := websiteService.GetWebsite(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, website)
}

// @Tags Website Nginx
// @Summary Search website nginx by id
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/:id/config/:type [get]
func (b *BaseApi) GetWebsiteNginx(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	configType := c.Param("type")

	fileInfo, err := websiteService.GetWebsiteNginxConfig(id, configType)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, fileInfo)
}

// @Tags Website Nginx
// @Summary Load nginx conf
// @Accept json
// @Param request body request.NginxScopeReq true "request"
// @Success 200 {object} response.WebsiteNginxConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/config [post]
func (b *BaseApi) GetNginxConfig(c *gin.Context) {
	var req request.NginxScopeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	config, err := websiteService.GetNginxConfigByScope(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, config)
}

// @Tags Website Nginx
// @Summary Update nginx conf
// @Accept json
// @Param request body request.NginxConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/config/update [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"nginx 配置修改 [domain]","formatEN":"Nginx conf update [domain]"}
func (b *BaseApi) UpdateNginxConfig(c *gin.Context) {
	var req request.NginxConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateNginxConfigByScope(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website HTTPS
// @Summary Load https conf
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.WebsiteHTTPS
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/:id/https [get]
func (b *BaseApi) GetHTTPSConfig(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	res, err := websiteService.GetWebsiteHTTPS(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website HTTPS
// @Summary Update https conf
// @Accept json
// @Param request body request.WebsiteHTTPSOp true "request"
// @Success 200 {object} response.WebsiteHTTPS
// @Security ApiKeyAuth
// @Security Timestamp
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
		helper.InternalServer(c, err)
		return
	}
	tx.Commit()
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Check before create website
// @Accept json
// @Param request body request.WebsiteInstallCheckReq true "request"
// @Success 200 {array} response.WebsitePreInstallCheck
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/check [post]
func (b *BaseApi) CreateWebsiteCheck(c *gin.Context) {
	var req request.WebsiteInstallCheckReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := websiteService.PreInstallCheck(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Website Nginx
// @Summary Update website nginx conf
// @Accept json
// @Param request body request.WebsiteNginxUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/nginx/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[domain] Nginx 配置修改","formatEN":"[domain] Nginx conf update"}
func (b *BaseApi) UpdateWebsiteNginxConfig(c *gin.Context) {
	var req request.WebsiteNginxUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateNginxConfigFile(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Operate website log
// @Accept json
// @Param request body request.WebsiteLogReq true "request"
// @Success 200 {object} response.WebsiteLog
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/log [post]
// @x-panel-log {"bodyKeys":["id", "operate"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"[domain][operate] 日志","formatEN":"[domain][operate] logs"}
func (b *BaseApi) OpWebsiteLog(c *gin.Context) {
	var req request.WebsiteLogReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.OpWebsiteLog(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Change default server
// @Accept json
// @Param request body request.WebsiteDefaultUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/default/server [post]
// @x-panel-log {"bodyKeys":["id", "operate"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"修改默认 server => [domain]","formatEN":"Change default server => [domain]"}
func (b *BaseApi) ChangeDefaultServer(c *gin.Context) {
	var req request.WebsiteDefaultUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.ChangeDefaultServer(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website PHP
// @Summary Update php version
// @Accept json
// @Param request body request.WebsitePHPVersionReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/php/version [post]
// @x-panel-log {"bodyKeys":["websiteId"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteId","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"php 版本变更 [domain]","formatEN":"php version update [domain]"}
func (b *BaseApi) ChangePHPVersion(c *gin.Context) {
	var req request.WebsitePHPVersionReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.ChangePHPVersion(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get rewrite conf
// @Accept json
// @Param request body request.NginxRewriteReq true "request"
// @Success 200 {object} response.NginxRewriteRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/rewrite [post]
func (b *BaseApi) GetRewriteConfig(c *gin.Context) {
	var req request.NginxRewriteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetRewriteConfig(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update rewrite conf
// @Accept json
// @Param request body request.NginxRewriteUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/rewrite/update [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"伪静态配置修改 [domain]","formatEN":"Nginx conf rewrite update [domain]"}
func (b *BaseApi) UpdateRewriteConfig(c *gin.Context) {
	var req request.NginxRewriteUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateRewriteConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update Site Dir
// @Accept json
// @Param request body request.WebsiteUpdateDir true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dir/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新网站 [domain] 目录","formatEN":"Update  domain [domain] dir"}
func (b *BaseApi) UpdateSiteDir(c *gin.Context) {
	var req request.WebsiteUpdateDir
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateSiteDir(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update Site Dir permission
// @Accept json
// @Param request body request.WebsiteUpdateDirPermission true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dir/permission [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新网站 [domain] 目录权限","formatEN":"Update  domain [domain] dir permission"}
func (b *BaseApi) UpdateSiteDirPermission(c *gin.Context) {
	var req request.WebsiteUpdateDirPermission
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateSitePermission(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get proxy conf
// @Accept json
// @Param request body request.WebsiteProxyReq true "request"
// @Success 200 {array} request.WebsiteProxyConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/proxies [post]
func (b *BaseApi) GetProxyConfig(c *gin.Context) {
	var req request.WebsiteProxyReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetProxies(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update proxy conf
// @Accept json
// @Param request body request.WebsiteProxyConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/proxies/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"修改网站 [domain] 反向代理配置 ","formatEN":"Update domain [domain] proxy config"}
func (b *BaseApi) UpdateProxyConfig(c *gin.Context) {
	var req request.WebsiteProxyConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.OperateProxy(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update proxy file
// @Accept json
// @Param request body request.NginxProxyUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/proxy/file [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新反向代理文件 [domain]","formatEN":"Nginx conf proxy file update [domain]"}
func (b *BaseApi) UpdateProxyConfigFile(c *gin.Context) {
	var req request.NginxProxyUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateProxyFile(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get AuthBasic conf
// @Accept json
// @Param request body request.NginxAuthReq true "request"
// @Success 200 {object} response.NginxAuthRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/auths [post]
func (b *BaseApi) GetAuthConfig(c *gin.Context) {
	var req request.NginxAuthReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetAuthBasics(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Get AuthBasic conf
// @Accept json
// @Param request body request.NginxAuthUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/auths/update [post]
func (b *BaseApi) UpdateAuthConfig(c *gin.Context) {
	var req request.NginxAuthUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateAuthBasic(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get AuthBasic conf
// @Accept json
// @Param request body request.NginxAuthReq true "request"
// @Success 200 {object} response.NginxPathAuthRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/auths/path [post]
func (b *BaseApi) GetPathAuthConfig(c *gin.Context) {
	var req request.NginxAuthReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetPathAuthBasics(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Get AuthBasic conf
// @Accept json
// @Param request body request.NginxPathAuthUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/auths/path/update [post]
func (b *BaseApi) UpdatePathAuthConfig(c *gin.Context) {
	var req request.NginxPathAuthUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdatePathAuthBasic(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get AntiLeech conf
// @Accept json
// @Param request body request.NginxCommonReq true "request"
// @Success 200 {object} response.NginxAntiLeechRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/leech [post]
func (b *BaseApi) GetAntiLeech(c *gin.Context) {
	var req request.NginxCommonReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetAntiLeech(req.WebsiteID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update AntiLeech
// @Accept json
// @Param request body request.NginxAntiLeechUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/leech/update [post]
func (b *BaseApi) UpdateAntiLeech(c *gin.Context) {
	var req request.NginxAntiLeechUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateAntiLeech(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update redirect conf
// @Accept json
// @Param request body request.NginxRedirectReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/redirect/update [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"修改网站 [domain] 重定向理配置 ","formatEN":"Update domain [domain] redirect config"}
func (b *BaseApi) UpdateRedirectConfig(c *gin.Context) {
	var req request.NginxRedirectReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.OperateRedirect(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get redirect conf
// @Accept json
// @Param request body request.WebsiteProxyReq true "request"
// @Success 200 {array} response.NginxRedirectConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/redirect [post]
func (b *BaseApi) GetRedirectConfig(c *gin.Context) {
	var req request.WebsiteRedirectReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.GetRedirect(req.WebsiteID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update redirect file
// @Accept json
// @Param request body request.NginxRedirectUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/redirect/file [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"更新重定向文件 [domain]","formatEN":"Nginx conf redirect file update [domain]"}
func (b *BaseApi) UpdateRedirectConfigFile(c *gin.Context) {
	var req request.NginxRedirectUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateRedirectFile(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get website dir
// @Accept json
// @Param request body request.WebsiteCommonReq true "request"
// @Success 200 {object} response.WebsiteDirConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dir [post]
func (b *BaseApi) GetDirConfig(c *gin.Context) {
	var req request.WebsiteCommonReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteService.LoadWebsiteDirConfig(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Get default html
// @Accept json
// @Success 200 {object} response.WebsiteHtmlRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/default/html/:type [get]
func (b *BaseApi) GetDefaultHtml(c *gin.Context) {
	resourceType, err := helper.GetStrParamByKey(c, "type")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	fileInfo, err := websiteService.GetDefaultHtml(resourceType)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, fileInfo)
}

// @Tags Website
// @Summary Update default html
// @Accept json
// @Param request body request.WebsiteHtmlUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/default/html/update [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新默认 html","formatEN":"Update default html"}
func (b *BaseApi) UpdateDefaultHtml(c *gin.Context) {
	var req request.WebsiteHtmlUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateDefaultHtml(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get website upstreams
// @Accept json
// @Param request body request.WebsiteCommonReq true "request"
// @Success 200 {array} dto.NginxUpstream
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/lbs [get]
func (b *BaseApi) GetLoadBalances(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	res, err := websiteService.GetLoadBalances(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Create website upstream
// @Accept json
// @Param request body request.WebsiteLBCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/lbs/create [post]
func (b *BaseApi) CreateLoadBalance(c *gin.Context) {
	var req request.WebsiteLBCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.CreateLoadBalance(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Delete website upstream
// @Accept json
// @Param request body request.WebsiteLBDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/lbs/delete [post]
func (b *BaseApi) DeleteLoadBalance(c *gin.Context) {
	var req request.WebsiteLBDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.DeleteLoadBalance(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update website upstream
// @Accept json
// @Param request body request.WebsiteLBUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/lbs/update [post]
func (b *BaseApi) UpdateLoadBalance(c *gin.Context) {
	var req request.WebsiteLBUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateLoadBalance(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Update website upstream file
// @Accept json
// @Param request body request.WebsiteLBUpdateFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/lbs/file [post]
func (b *BaseApi) UpdateLoadBalanceFile(c *gin.Context) {
	var req request.WebsiteLBUpdateFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateLoadBalanceFile(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

func (b *BaseApi) ChangeWebsiteGroup(c *gin.Context) {
	var req dto.UpdateGroup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.ChangeGroup(req.Group, req.NewGroup); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary update website proxy cache config
// @Accept json
// @Param request body request.NginxProxyCacheUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/proxy/config [post]
func (b *BaseApi) UpdateProxyCache(c *gin.Context) {
	var req request.NginxProxyCacheUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.UpdateProxyCache(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Summary Get website proxy cache config
// @Accept json
// @Param id path int true "id"
// @Success 200 {object} response.NginxProxyCache
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/proxy/config/{id} [get]
func (b *BaseApi) GetProxyCache(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	res, err := websiteService.GetProxyCache(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Set Real IP
// @Accept json
// @Param request body request.WebsiteRealIP true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/realip [post]
// @x-panel-log {"bodyKeys":["websiteID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"websiteID","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"修改 [domain] 网站真实IP配置 ","formatEN":"Modify the real IP configuration of [domain] website"}
func (b *BaseApi) SetRealIPConfig(c *gin.Context) {
	var req request.WebsiteRealIP
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.SetRealIPConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Get Real IP Config
// @Accept json
// @Param id path int true "id"
// @Success 200 {object} response.WebsiteRealIP
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/realip/config/{id} [get]
func (b *BaseApi) GetRealIPConfig(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	res, err := websiteService.GetRealIPConfig(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Get website resource
// @Accept json
// @Param id path int true "id"
// @Success 200 {object} response.Resource
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/resource/{id} [get]
func (b *BaseApi) GetWebsiteResource(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	res, err := websiteService.GetWebsiteResource(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Get databases
// @Accept json
// @Success 200 {object} response.Database
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/databases [get]
func (b *BaseApi) GetWebsiteDatabase(c *gin.Context) {
	res, err := websiteService.ListDatabases()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Change website database
// @Accept json
// @Param request body request.ChangeDatabase true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/databases [post]
func (b *BaseApi) ChangeWebsiteDatabase(c *gin.Context) {
	var req request.ChangeDatabase
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.ChangeDatabase(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary Operate custom rewrite
// @Accept json
// @Param request body request.CustomRewriteOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/rewrite/custom [post]
func (b *BaseApi) OperateCustomRewrite(c *gin.Context) {
	var req request.CustomRewriteOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.OperateCustomRewrite(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Website
// @Summary List custom rewrite
// @Accept json
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/rewrite/custom [get]
func (b *BaseApi) ListCustomRewrite(c *gin.Context) {
	res, err := websiteService.ListCustomRewrite()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Clear Website proxy cache
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/proxy/clear [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清理 Openresty 代理缓存","formatEN":"Clear nginx proxy cache"}
func (b *BaseApi) ClearProxyCache(c *gin.Context) {
	var req request.NginxCommonReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteService.ClearProxyCache(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}
