package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags File
// @Summary List RecycleBin files
// @Description 获取回收站文件列表
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/recycle/search [post]
func (b *BaseApi) SearchRecycleBinFile(c *gin.Context) {
	var req dto.PageInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := recycleBinService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags File
// @Summary Reduce RecycleBin files
// @Description 还原回收站文件
// @Accept json
// @Param request body request.RecycleBinReduce true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/recycle/reduce [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"还原回收站文件 [name]","formatEN":"Reduce RecycleBin file [name]"}
func (b *BaseApi) ReduceRecycleBinFile(c *gin.Context) {
	var req request.RecycleBinReduce
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := recycleBinService.Reduce(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags File
// @Summary Clear RecycleBin files
// @Description 清空回收站文件
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Router /files/recycle/clear [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清空回收站","formatEN":"清空回收站"}
func (b *BaseApi) ClearRecycleBinFile(c *gin.Context) {
	if err := recycleBinService.Clear(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags File
// @Summary Get RecycleBin status
// @Description 获取回收站状态
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Router /files/recycle/status [get]
func (b *BaseApi) GetRecycleStatus(c *gin.Context) {
	settingInfo, err := settingService.GetSettingInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, settingInfo.FileRecycleBin)
}
