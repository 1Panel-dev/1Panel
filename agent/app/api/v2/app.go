package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// @Tags App
// @Summary List apps
// @Accept json
// @Param request body request.AppSearch true "request"
// @Success 200 {object} response.AppRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/search [post]
func (b *BaseApi) SearchApp(c *gin.Context) {
	var req request.AppSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	list, err := appService.PageApp(c, req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags App
// @Summary Sync remote app list
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/sync/remote [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"应用商店同步","formatEN":"App store synchronization"}
func (b *BaseApi) SyncApp(c *gin.Context) {
	var req dto.OperateWithTask
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := appService.GetAppUpdate()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	if !res.CanUpdate {
		if res.IsSyncing {
			helper.SuccessWithMsg(c, i18n.GetMsgByKey("AppStoreIsSyncing"))
		} else {
			helper.SuccessWithMsg(c, i18n.GetMsgByKey("AppStoreIsUpToDate"))
		}
		return
	}
	if err = appService.SyncAppListFromRemote(req.TaskID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Sync local  app list
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/sync/local [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"应用商店同步","formatEN":"App store synchronization"}
func (b *BaseApi) SyncLocalApp(c *gin.Context) {
	var req dto.OperateWithTask
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	go appService.SyncAppListFromLocal(req.TaskID)
	helper.Success(c)
}

// @Tags App
// @Summary Search app by key
// @Accept json
// @Param key path string true "app key"
// @Success 200 {object} response.AppDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/:key [get]
func (b *BaseApi) GetApp(c *gin.Context) {
	appKey, err := helper.GetStrParamByKey(c, "key")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	appDTO, err := appService.GetApp(c, appKey)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, appDTO)
}

// @Tags App
// @Summary Search app detail by appid
// @Accept json
// @Param appId path integer true "app id"
// @Param version path string true "app 版本"
// @Param version path string true "app 类型"
// @Success 200 {object} response.AppDetailDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/detail/:appId/:version/:type [get]
func (b *BaseApi) GetAppDetail(c *gin.Context) {
	appID, err := helper.GetIntParamByKey(c, "appId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	version := c.Param("version")
	appType := c.Param("type")
	appDetailDTO, err := appService.GetAppDetail(appID, version, appType)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}

// @Tags App
// @Summary Get app detail by id
// @Accept json
// @Param appId path integer true "id"
// @Success 200 {object} response.AppDetailDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/details/:id [get]
func (b *BaseApi) GetAppDetailByID(c *gin.Context) {
	appDetailID, err := helper.GetIntParamByKey(c, "id")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	appDetailDTO, err := appService.GetAppDetailByID(appDetailID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}

// @Tags App
// @Summary Install app
// @Accept json
// @Param request body request.AppInstallCreate true "request"
// @Success 200 {object} model.AppInstall
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/install [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"安装应用 [name]","formatEN":"Install app [name]"}
func (b *BaseApi) InstallApp(c *gin.Context) {
	var req request.AppInstallCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	install, err := appService.Install(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, install)
}

func (b *BaseApi) GetAppTags(c *gin.Context) {
	tags, err := appService.GetAppTags(c)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, tags)
}

// @Tags App
// @Summary Get app list update
// @Success 200 {object} response.AppUpdateRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/checkupdate [get]
func (b *BaseApi) GetAppListUpdate(c *gin.Context) {
	res, err := appService.GetAppUpdate()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags App
// @Summary Get app icon by app_id
// @Accept json
// @Param appId path integer true "app id"
// @Success 200 {file} file "app icon"
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/icon/:appId [get]
func (b *BaseApi) GetAppIcon(c *gin.Context) {
	appID, err := helper.GetIntParamByKey(c, "appID")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	iconBytes, err := appService.GetAppIcon(appID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	c.Data(http.StatusOK, "image/png", iconBytes)
}
