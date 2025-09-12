package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags App
// @Summary Page app installed
// @Accept json
// @Param request body request.AppInstalledSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/search [post]
func (b *BaseApi) SearchAppInstalled(c *gin.Context) {
	var req request.AppInstalledSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if req.All {
		list, err := appInstallService.SearchForWebsite(req)
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{
			Items: list,
			Total: int64(len(list)),
		})
	} else {
		total, list, err := appInstallService.Page(req)
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{
			Items: list,
			Total: total,
		})
	}
}

// @Tags App
// @Summary List app installed
// @Accept json
// @Success 200 {array} dto.AppInstallInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/list [get]
func (b *BaseApi) ListAppInstalled(c *gin.Context) {
	list, err := appInstallService.GetInstallList()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags App
// @Summary Check app installed
// @Accept json
// @Param request body request.AppInstalledInfo true "request"
// @Success 200 {object} response.AppInstalledCheck
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/check [post]
func (b *BaseApi) CheckAppInstalled(c *gin.Context) {
	var req request.AppInstalledInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	checkData, err := appInstallService.CheckExist(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, checkData)
}

// @Tags App
// @Summary Search app port by key
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {integer} port
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/loadport [post]
func (b *BaseApi) LoadPort(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	port, err := appInstallService.LoadPort(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, port)
}

// @Tags App
// @Summary Search app password by key
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {object} response.DatabaseConn
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/conninfo [POST]
func (b *BaseApi) LoadConnInfo(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	conn, err := appInstallService.LoadConnInfo(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, conn)
}

// @Tags App
// @Summary Check before delete
// @Accept json
// @Param appInstallId path integer true "App install id"
// @Success 200 {array} dto.AppResource
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/delete/check/:appInstallId [get]
func (b *BaseApi) DeleteCheck(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	checkData, err := appInstallService.DeleteCheck(appInstallId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, checkData)
}

// Sync app installed
// @Tags App
// @Summary Sync app installed
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"同步已安装应用列表","formatEN":"Sync the list of installed apps"}
func (b *BaseApi) SyncInstalled(c *gin.Context) {
	if err := appInstallService.SyncAll(false); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Operate installed app
// @Accept json
// @Param request body request.AppInstalledOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/op [post]
// @x-panel-log {"bodyKeys":["installId","operate"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"installId","isList":false,"db":"app_installs","output_column":"app_id","output_value":"appId"},{"input_column":"id","input_value":"installId","isList":false,"db":"app_installs","output_column":"name","output_value":"appName"},{"input_column":"id","input_value":"appId","isList":false,"db":"apps","output_column":"key","output_value":"appKey"}],"formatZH":"[operate] 应用 [appKey][appName]","formatEN":"[operate] App [appKey][appName]"}
func (b *BaseApi) OperateInstalled(c *gin.Context) {
	var req request.AppInstalledOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.Operate(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Search app service by key
// @Accept json
// @Param key path string true "request"
// @Success 200 {array} response.AppService
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/services/:key [get]
func (b *BaseApi) GetServices(c *gin.Context) {
	key := c.Param("key")
	services, err := appInstallService.GetServices(key)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, services)
}

// @Tags App
// @Summary Search app update version by install id
// @Accept json
// @Param appInstallId path integer true "request"
// @Success 200 {array} dto.AppVersion
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/update/versions [post]
func (b *BaseApi) GetUpdateVersions(c *gin.Context) {
	var req request.AppUpdateVersion
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	versions, err := appInstallService.GetUpdateVersions(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, versions)
}

// @Tags App
// @Summary Change app port
// @Accept json
// @Param request body request.PortUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/port/change [post]
// @x-panel-log {"bodyKeys":["key","name","port"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"应用端口修改 [key]-[name] => [port]","formatEN":"Application port update [key]-[name] => [port]"}
func (b *BaseApi) ChangeAppPort(c *gin.Context) {
	var req request.PortUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.ChangeAppPort(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Search default config by key
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/conf [post]
func (b *BaseApi) GetDefaultConfig(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content, err := appInstallService.GetDefaultConfigByKey(req.Type, req.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, content)
}

// @Tags App
// @Summary Search params by appInstallId
// @Accept json
// @Param appInstallId path string true "request"
// @Success 200 {object} response.AppConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/params/:appInstallId [get]
func (b *BaseApi) GetParams(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	content, err := appInstallService.GetParams(appInstallId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, content)
}

// @Tags App
// @Summary Change app params
// @Accept json
// @Param request body request.AppInstalledUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/params/update [post]
// @x-panel-log {"bodyKeys":["installId"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"应用参数修改 [installId]","formatEN":"Application param update [installId]"}
func (b *BaseApi) UpdateInstalled(c *gin.Context) {
	var req request.AppInstalledUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.Update(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Update app config
// @Accept json
// @Param request body request.AppConfigUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/config/update [post]
// @x-panel-log {"bodyKeys":["installID","webUI"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"应用配置更新 [installID]","formatEN":"Application config update [installID]"}
func (b *BaseApi) UpdateAppConfig(c *gin.Context) {
	var req request.AppConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.UpdateAppConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Get app install info
// @Accept json
// @Param appInstallId path integer true "App install id"
// @Success 200 {object} dto.AppInstallInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/info/:appInstallId [get]
func (b *BaseApi) GetAppInstallInfo(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	info, err := appInstallService.GetAppInstallInfo(appInstallId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, info)
}
