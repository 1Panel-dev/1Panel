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
// @Description 获取应用列表
// @Accept json
// @Param request body request.AppSearch true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/search [post]
func (b *BaseApi) SearchApp(c *gin.Context) {
	var req request.AppSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	list, err := appService.PageApp(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags App
// @Summary Sync app list
// @Description 同步应用列表
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"应用商店同步","formatEN":"App store synchronization"}
func (b *BaseApi) SyncApp(c *gin.Context) {
	go appService.SyncAppListFromLocal()
	res, err := appService.GetAppUpdate()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	if !res.CanUpdate {
		helper.SuccessWithMsg(c, i18n.GetMsgByKey("AppStoreIsUpToDate"))
		return
	}
	go func() {
		global.LOG.Infof("sync app list start ...")
		if err := appService.SyncAppListFromRemote(); err != nil {
			global.LOG.Errorf("sync app list error [%s]", err.Error())
		} else {
			global.LOG.Infof("sync app list success!")
		}
	}()
	helper.SuccessWithData(c, "")
}

// @Tags App
// @Summary Search app by key
// @Description 通过 key 获取应用信息
// @Accept json
// @Param key path string true "app key"
// @Success 200 {object} response.AppDTO
// @Security ApiKeyAuth
// @Router /apps/:key [get]
func (b *BaseApi) GetApp(c *gin.Context) {
	appKey, err := helper.GetStrParamByKey(c, "key")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	appDTO, err := appService.GetApp(appKey)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, appDTO)
}

// @Tags App
// @Summary Search app detail by appid
// @Description 通过 appid 获取应用详情
// @Accept json
// @Param appId path integer true "app id"
// @Param version path string true "app 版本"
// @Param version path string true "app 类型"
// @Success 200 {object} response.AppDetailDTO
// @Security ApiKeyAuth
// @Router /apps/detail/:appId/:version/:type [get]
func (b *BaseApi) GetAppDetail(c *gin.Context) {
	appId, err := helper.GetIntParamByKey(c, "appId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	version := c.Param("version")
	appType := c.Param("type")
	appDetailDTO, err := appService.GetAppDetail(appId, version, appType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}

// @Tags App
// @Summary Get app detail by id
// @Description 通过 id 获取应用详情
// @Accept json
// @Param appId path integer true "id"
// @Success 200 {object} response.AppDetailDTO
// @Security ApiKeyAuth
// @Router /apps/details/:id [get]
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
// @Description 获取忽略的应用版本
// @Accept json
// @Success 200 {object} response.IgnoredApp
// @Security ApiKeyAuth
// @Router /apps/ingored [get]
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
// @Description 安装应用
// @Accept json
// @Param request body request.AppInstallCreate true "request"
// @Success 200 {object} model.AppInstall
// @Security ApiKeyAuth
// @Router /apps/install [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[{"input_column":"name","input_value":"name","isList":false,"db":"app_installs","output_column":"app_id","output_value":"appId"},{"info":"appId","isList":false,"db":"apps","output_column":"key","output_value":"appKey"}],"formatZH":"安装应用 [appKey]-[name]","formatEN":"Install app [appKey]-[name]"}
func (b *BaseApi) InstallApp(c *gin.Context) {
	var req request.AppInstallCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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
	tags, err := appService.GetAppTags()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, tags)
}

// @Tags App
// @Summary Get app list update
// @Description 获取应用更新版本
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/checkupdate [get]
func (b *BaseApi) GetAppListUpdate(c *gin.Context) {
	res, err := appService.GetAppUpdate()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}
