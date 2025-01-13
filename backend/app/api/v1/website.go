package v1

import (
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
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
// @Success 200 {array} response.WebsiteDTO
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Success 200 {array} response.WebsiteOption
// @Security ApiKeyAuth
// @Security Timestamp
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
	if len(req.FtpPassword) != 0 {
		pass, err := base64.StdEncoding.DecodeString(req.FtpPassword)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.FtpPassword = string(pass)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Search website by id
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.WebsiteDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/{id} [get]
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
// @Accept json
// @Param id path integer true "request"
// @Param type path string true "type"
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/{id}/config/{type} [get]
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website HTTPS
// @Summary Load https conf
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.WebsiteHTTPS
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/{id}/https [get]
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
// @Accept json
// @Param id path integer true "request"
// @Param request body request.WebsiteHTTPSOp true "request"
// @Success 200 {object} response.WebsiteHTTPS
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/{id}/https [post]
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Website
// @Summary Load website php conf
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} response.PHPConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/php/config/{id} [get]
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
// @Accept json
// @Param request body request.WebsitePHPConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body request.WebsitePHPFileUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Delete proxy conf
// @Accept json
// @Param request body request.WebsiteProxyDel true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/proxies/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"websites","output_column":"primary_domain","output_value":"domain"}],"formatZH":"删除网站 [domain] 反向代理配置","formatEN":"Delete domain [domain] proxy config"}
func (b *BaseApi) DeleteProxyConfig(c *gin.Context) {
	var req request.WebsiteProxyDel
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := websiteService.DeleteProxy(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Update AntiLeech conf
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website
// @Summary Get default html
// @Accept json
// @Param type path string true "type"
// @Success 200 {object} response.WebsiteHtmlRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/default/html/{type} [get]
func (b *BaseApi) GetDefaultHtml(c *gin.Context) {
	resourceType, err := helper.GetStrParamByKey(c, "type")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	fileInfo, err := websiteService.GetDefaultHtml(resourceType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
