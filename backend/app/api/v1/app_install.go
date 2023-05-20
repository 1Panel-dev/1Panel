package v1

import (
	"errors"
	"reflect"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags App
// @Summary List app installed
// @Description 获取已安装应用列表
// @Accept json
// @Param request body request.AppInstalledSearch true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/installed/search [post]
func (b *BaseApi) SearchAppInstalled(c *gin.Context) {
	var req request.AppInstalledSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if !reflect.DeepEqual(req.PageInfo, dto.PageInfo{}) {
		total, list, err := appInstallService.Page(req)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{
			Items: list,
			Total: total,
		})
	} else {
		list, err := appInstallService.SearchForWebsite(req)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, list)
	}
}

// @Tags App
// @Summary Check app installed
// @Description 检查应用安装情况
// @Accept json
// @Param key path string true "request"
// @Success 200 {object} response.AppInstalledCheck
// @Security ApiKeyAuth
// @Router /apps/installed/check/:key [get]
func (b *BaseApi) CheckAppInstalled(c *gin.Context) {
	key, ok := c.Params.Get("key")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error key in path"))
		return
	}
	checkData, err := appInstallService.CheckExist(key)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, checkData)
}

// @Tags App
// @Summary Search app port by key
// @Description 获取应用端口
// @Accept json
// @Param key path string true "request"
// @Success 200 {integer} port
// @Security ApiKeyAuth
// @Router /apps/installed/loadport/:key [get]
func (b *BaseApi) LoadPort(c *gin.Context) {
	key, ok := c.Params.Get("key")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error key in path"))
		return
	}
	port, err := appInstallService.LoadPort(key)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, port)
}

// @Tags App
// @Summary Search app password by key
// @Description 获取应用连接信息
// @Accept json
// @Param key path string true "request"
// @Success 200 {string} response.DatabaseConn
// @Security ApiKeyAuth
// @Router /apps/installed/conninfo/:key [get]
func (b *BaseApi) LoadConnInfo(c *gin.Context) {
	key, ok := c.Params.Get("key")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error key in path"))
		return
	}
	conn, err := appInstallService.LoadConnInfo(key)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, conn)
}

// @Tags App
// @Summary Check before delete
// @Description 删除前检查
// @Accept json
// @Param appInstallId path integer true "App install id"
// @Success 200 {anrry} dto.AppResource
// @Security ApiKeyAuth
// @Router /apps/installed/delete/check/:appInstallId [get]
func (b *BaseApi) DeleteCheck(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	checkData, err := appInstallService.DeleteCheck(appInstallId)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, checkData)
}

// Sync app installed
// @Tags App
// @Summary Sync app installed
// @Description 同步已安装应用列表
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/installed/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"同步已安装应用列表","formatEN":"Sync the list of installed apps"}
func (b *BaseApi) SyncInstalled(c *gin.Context) {
	if err := appInstallService.SyncAll(false); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, "")
}

// @Tags App
// @Summary Operate installed app
// @Description 操作已安装应用
// @Accept json
// @Param request body request.AppInstalledOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/installed/op [post]
// @x-panel-log {"bodyKeys":["installId","operate"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"installId","isList":false,"db":"app_installs","output_colume":"app_id","output_value":"appId"},{"input_colume":"id","input_value":"installId","isList":false,"db":"app_installs","output_colume":"name","output_value":"appName"},{"input_colume":"id","input_value":"appId","isList":false,"db":"apps","output_colume":"key","output_value":"appKey"}],"formatZH":"[appKey] 应用 [appName] [operate]","formatEN":"[appKey] App [appName] [operate]"}
func (b *BaseApi) OperateInstalled(c *gin.Context) {
	var req request.AppInstalledOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := appInstallService.Operate(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags App
// @Summary Search app service by key
// @Description 通过 key 获取应用 service
// @Accept json
// @Param key path string true "request"
// @Success 200 {anrry} response.AppService
// @Security ApiKeyAuth
// @Router /apps/services/:key [get]
func (b *BaseApi) GetServices(c *gin.Context) {
	key := c.Param("key")
	services, err := appInstallService.GetServices(key)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, services)
}

// @Tags App
// @Summary Search app update version by install id
// @Description 通过 install id 获取应用更新版本
// @Accept json
// @Param appInstallId path integer true "request"
// @Success 200 {anrry} dto.AppVersion
// @Security ApiKeyAuth
// @Router /apps/installed/:appInstallId/versions [get]
func (b *BaseApi) GetUpdateVersions(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	versions, err := appInstallService.GetUpdateVersions(appInstallId)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, versions)
}

// @Tags App
// @Summary Change app port
// @Description 修改应用端口
// @Accept json
// @Param request body request.PortUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/installed/port/change [post]
// @x-panel-log {"bodyKeys":["key","name","port"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"应用端口修改 [key]-[name] => [port]","formatEN":"Application port update [key]-[name] => [port]"}
func (b *BaseApi) ChangeAppPort(c *gin.Context) {
	var req request.PortUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := appInstallService.ChangeAppPort(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags App
// @Summary Search default config by key
// @Description 通过 key 获取应用默认配置
// @Accept json
// @Param key path string true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Router /apps/installed/conf/:key [get]
func (b *BaseApi) GetDefaultConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	content, err := appInstallService.GetDefaultConfigByKey(key)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, content)
}

// @Tags App
// @Summary Search params by appInstallId
// @Description 通过 install id 获取应用参数
// @Accept json
// @Param appInstallId path string true "request"
// @Success 200 {object} response.AppParam
// @Security ApiKeyAuth
// @Router /apps/installed/params/:appInstallId [get]
func (b *BaseApi) GetParams(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInternalServer, nil)
		return
	}
	content, err := appInstallService.GetParams(appInstallId)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, content)
}

// @Tags App
// @Summary Change app params
// @Description 修改应用参数
// @Accept json
// @Param request body request.AppInstalledUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/installed/params/update [post]
// @x-panel-log {"bodyKeys":["installId"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"应用参数修改 [installId]","formatEN":"Application param update [installId]"}
func (b *BaseApi) UpdateInstalled(c *gin.Context) {
	var req request.AppInstalledUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := appInstallService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
