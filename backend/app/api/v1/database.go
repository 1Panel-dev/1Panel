package v1

import (
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Database
// @Summary Create database
// @Accept json
// @Param request body dto.DatabaseCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db [post]
// @x-panel-log {"bodyKeys":["name", "type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建远程数据库 [name][type]","formatEN":"create database [name][type]"}
func (b *BaseApi) CreateDatabase(c *gin.Context) {
	var req dto.DatabaseCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if req.SSL {
		key, _ := base64.StdEncoding.DecodeString(req.ClientKey)
		req.ClientKey = string(key)
		cert, _ := base64.StdEncoding.DecodeString(req.ClientCert)
		req.ClientCert = string(cert)
		ca, _ := base64.StdEncoding.DecodeString(req.RootCert)
		req.RootCert = string(ca)
	}

	if err := databaseService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database
// @Summary Check database
// @Accept json
// @Param request body dto.DatabaseCreate true "request"
// @Success 200 {boolean} isOk
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db/check [post]
// @x-panel-log {"bodyKeys":["name", "type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"检测远程数据库 [name][type] 连接性","formatEN":"check if database [name][type] is connectable"}
func (b *BaseApi) CheckDatabase(c *gin.Context) {
	var req dto.DatabaseCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if req.SSL {
		clientKey, _ := base64.StdEncoding.DecodeString(req.ClientKey)
		req.ClientKey = string(clientKey)
		clientCert, _ := base64.StdEncoding.DecodeString(req.ClientCert)
		req.ClientCert = string(clientCert)
		rootCert, _ := base64.StdEncoding.DecodeString(req.RootCert)
		req.RootCert = string(rootCert)
	}

	helper.SuccessWithData(c, databaseService.CheckDatabase(req))
}

// @Tags Database
// @Summary Page databases
// @Accept json
// @Param request body dto.DatabaseSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db/search [post]
func (b *BaseApi) SearchDatabase(c *gin.Context) {
	var req dto.DatabaseSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := databaseService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Database
// @Summary List databases
// @Param type path string true "type"
// @Success 200 {array} dto.DatabaseOption
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db/list/{type} [get]
func (b *BaseApi) ListDatabase(c *gin.Context) {
	dbType, err := helper.GetStrParamByKey(c, "type")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	list, err := databaseService.List(dbType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Database
// @Summary Retrieve database list based on type
// @Param type path string true "type"
// @Success 200 {array} dto.DatabaseItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db/item/{type} [get]
func (b *BaseApi) LoadDatabaseItems(c *gin.Context) {
	dbType, err := helper.GetStrParamByKey(c, "type")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	list, err := databaseService.LoadItems(dbType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Database
// @Summary Get databases
// @Param name path string true "name"
// @Success 200 {object} dto.DatabaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db/{name} [get]
func (b *BaseApi) GetDatabase(c *gin.Context) {
	name, err := helper.GetStrParamByKey(c, "name")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	data, err := databaseService.Get(name)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database
// @Summary Check before delete remote database
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /db/remote/del/check [post]
func (b *BaseApi) DeleteCheckDatabase(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	apps, err := databaseService.DeleteCheck(req.ID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, apps)
}

// @Tags Database
// @Summary Delete database
// @Accept json
// @Param request body dto.DatabaseDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"databases","output_column":"name","output_value":"names"}],"formatZH":"删除远程数据库 [names]","formatEN":"delete database [names]"}
func (b *BaseApi) DeleteDatabase(c *gin.Context) {
	var req dto.DatabaseDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := databaseService.Delete(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database
// @Summary Update database
// @Accept json
// @Param request body dto.DatabaseUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/db/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新远程数据库 [name]","formatEN":"update database [name]"}
func (b *BaseApi) UpdateDatabase(c *gin.Context) {
	var req dto.DatabaseUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if req.SSL {
		cKey, _ := base64.StdEncoding.DecodeString(req.ClientKey)
		req.ClientKey = string(cKey)
		cCert, _ := base64.StdEncoding.DecodeString(req.ClientCert)
		req.ClientCert = string(cCert)
		ca, _ := base64.StdEncoding.DecodeString(req.RootCert)
		req.RootCert = string(ca)
	}

	if err := databaseService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
