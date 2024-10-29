package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Database Common
// @Summary Load base info
// @Description 获取数据库基础信息
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {object} dto.DBBaseInfo
// @Security ApiKeyAuth
// @Router /databases/common/info [post]
func (b *BaseApi) LoadDBBaseInfo(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := dbCommonService.LoadBaseInfo(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Common
// @Summary Load Database conf
// @Description 获取数据库配置文件
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/common/load/file [post]
func (b *BaseApi) LoadDBFile(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content, err := dbCommonService.LoadDatabaseFile(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, content)
}

// @Tags Database Common
// @Summary Update conf by upload file
// @Description 上传替换配置文件
// @Accept json
// @Param request body dto.DBConfUpdateByFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/common/update/conf [post]
// @x-panel-log {"bodyKeys":["type","database"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 [type] 数据库 [database] 配置信息","formatEN":"update the [type] [database] database configuration information"}
func (b *BaseApi) UpdateDBConfByFile(c *gin.Context) {
	var req dto.DBConfUpdateByFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := dbCommonService.UpdateConfByFile(req); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
