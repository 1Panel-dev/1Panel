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

// List app installed
// @Tags App
// @Summary Search app list installed
// @Description 获取已安装应用列表
// @Accept json
// @Param request body request.AppInstalledSearch true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /apps/installed [post]
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
		list, err := appInstallService.Search(req)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, list)
	}
}

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

func (b *BaseApi) LoadPassword(c *gin.Context) {
	key, ok := c.Params.Get("key")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error key in path"))
		return
	}
	password, err := appInstallService.LoadPassword(key)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, password)
}

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

func (b *BaseApi) SyncInstalled(c *gin.Context) {
	if err := appInstallService.SyncAll(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, "")
}

func (b *BaseApi) SearchInstalledBackup(c *gin.Context) {
	var req request.AppBackupSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	total, list, err := appInstallService.PageInstallBackups(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

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

func (b *BaseApi) DeleteAppBackup(c *gin.Context) {
	var req request.AppBackupDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := appInstallService.DeleteBackup(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) GetServices(c *gin.Context) {
	key := c.Param("key")
	services, err := appInstallService.GetServices(key)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, services)
}

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
