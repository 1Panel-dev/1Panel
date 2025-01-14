package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/gin-gonic/gin"
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags App
// @Summary Sync app list
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"应用商店同步","formatEN":"App store synchronization"}
func (b *BaseApi) SyncApp(c *gin.Context) {
	go appService.SyncAppListFromLocal()
	res, err := appService.GetAppUpdate()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
	go func() {
		if err := appService.SyncAppListFromRemote(); err != nil {
			global.LOG.Errorf("Synchronization with the App Store failed [%s]", err.Error())
		}
	}()
	helper.SuccessWithData(c, "")
}

// @Tags App
// @Summary Search app by key
// @Accept json
// @Param key path string true "app key"
// @Success 200 {object} response.AppDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/{key} [get]
func (b *BaseApi) GetApp(c *gin.Context) {
	appKey, err := helper.GetStrParamByKey(c, "key")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	appDTO, err := appService.GetApp(c, appKey)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, appDTO)
}

// @Tags App
// @Summary Search app detail by appid
// @Accept json
// @Param appId path integer true "app id"
// @Param version path string true "app version"
// @Param type path string true "app type"
// @Success 200 {object} response.AppDetailDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/detail/{appId}/{version}/{type} [get]
func (b *BaseApi) GetAppDetail(c *gin.Context) {
	appID, err := helper.GetIntParamByKey(c, "appId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	version := c.Param("version")
	appType := c.Param("type")
	appDetailDTO, err := appService.GetAppDetail(appID, version, appType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}

// @Tags App
// @Summary Get app detail by id
// @Accept json
// @Param id path integer true "id"
// @Success 200 {object} response.AppDetailDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/details/{id} [get]
func (b *BaseApi) GetAppDetailByID(c *gin.Context) {
	appDetailID, err := helper.GetIntParamByKey(c, "id")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	appDetailDTO, err := appService.GetAppDetailByID(appDetailID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}

// @Tags App
// @Summary Get Ignore App
// @Accept json
// @Success 200 {object} response.IgnoredApp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/ignored [get]
func (b *BaseApi) GetIgnoredApp(c *gin.Context) {
	res, err := appService.GetIgnoredApp()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
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
	tx, ctx := helper.GetTxAndContext()
	install, err := appService.Install(ctx, req)
	if err != nil {
		tx.Rollback()
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	tx.Commit()
	helper.SuccessWithData(c, install)
}

func (b *BaseApi) GetAppTags(c *gin.Context) {
	tags, err := appService.GetAppTags(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}
