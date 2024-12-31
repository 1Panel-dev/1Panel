package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Database Common
// @Summary Load base info
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {object} dto.DBBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/common/info [post]
func (b *BaseApi) LoadDBBaseInfo(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := dbCommonService.LoadBaseInfo(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Common
// @Summary Load Database conf
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/common/load/file [post]
func (b *BaseApi) LoadDBFile(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content, err := dbCommonService.LoadDatabaseFile(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, content)
}

// @Tags Database Common
// @Summary Update conf by upload file
// @Accept json
// @Param request body dto.DBConfUpdateByFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/common/update/conf [post]
// @x-panel-log {"bodyKeys":["type","database"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 [type] 数据库 [database] 配置信息","formatEN":"update the [type] [database] database configuration information"}
func (b *BaseApi) UpdateDBConfByFile(c *gin.Context) {
	var req dto.DBConfUpdateByFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := dbCommonService.UpdateConfByFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
