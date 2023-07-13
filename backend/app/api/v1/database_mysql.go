package v1

import (
	"context"
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags Database Mysql
// @Summary Create mysql database
// @Description 创建 mysql 数据库
// @Accept json
// @Param request body dto.MysqlDBCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建 mysql 数据库 [name]","formatEN":"create mysql database [name]"}
func (b *BaseApi) CreateMysql(c *gin.Context) {
	var req dto.MysqlDBCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Password = string(password)
	}

	if _, err := mysqlService.Create(context.Background(), req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Mysql
// @Summary Update mysql database description
// @Description 更新 mysql 数据库库描述信息
// @Accept json
// @Param request body dto.UpdateDescription true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/description/update [post]
// @x-panel-log {"bodyKeys":["id","description"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"mysql 数据库 [name] 描述信息修改 [description]","formatEN":"The description of the mysql database [name] is modified => [description]"}
func (b *BaseApi) UpdateMysqlDescription(c *gin.Context) {
	var req dto.UpdateDescription
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := mysqlService.UpdateDescription(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Mysql
// @Summary Change mysql password
// @Description 修改 mysql 密码
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/change/password [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"更新数据库 [name] 密码","formatEN":"Update database [name] password"}
func (b *BaseApi) ChangeMysqlPassword(c *gin.Context) {
	var req dto.ChangeDBInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if len(req.Value) != 0 {
		value, err := base64.StdEncoding.DecodeString(req.Value)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Value = string(value)
	}

	if err := mysqlService.ChangePassword(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Mysql
// @Summary Change mysql access
// @Description 修改 mysql 访问权限
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/change/access [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"更新数据库 [name] 访问权限","formatEN":"Update database [name] access"}
func (b *BaseApi) ChangeMysqlAccess(c *gin.Context) {
	var req dto.ChangeDBInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := mysqlService.ChangeAccess(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Mysql
// @Summary Update mysql variables
// @Description mysql 性能调优
// @Accept json
// @Param request body dto.MysqlVariablesUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/variables/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"调整 mysql 数据库性能参数","formatEN":"adjust mysql database performance parameters"}
func (b *BaseApi) UpdateMysqlVariables(c *gin.Context) {
	var req []dto.MysqlVariablesUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := mysqlService.UpdateVariables(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Mysql
// @Summary Update mysql conf by upload file
// @Description 上传替换 mysql 配置文件
// @Accept json
// @Param request body dto.MysqlConfUpdateByFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/conffile/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新 mysql 数据库配置信息","formatEN":"update the mysql database configuration information"}
func (b *BaseApi) UpdateMysqlConfByFile(c *gin.Context) {
	var req dto.MysqlConfUpdateByFile
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := mysqlService.UpdateConfByFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Database Mysql
// @Summary Page mysql databases
// @Description 获取 mysql 数据库列表分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /databases/search [post]
func (b *BaseApi) SearchMysql(c *gin.Context) {
	var req dto.SearchWithPage
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := mysqlService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Database Mysql
// @Summary List mysql database names
// @Description 获取 mysql 数据库列表
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Router /databases/options [get]
func (b *BaseApi) ListDBName(c *gin.Context) {
	list, err := mysqlService.ListDBName()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Database Mysql
// @Summary Check before delete mysql database
// @Description Mysql 数据库删除前检查
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Router /databases/del/check [post]
func (b *BaseApi) DeleteCheckMysql(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	apps, err := mysqlService.DeleteCheck(req.ID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, apps)
}

// @Tags Database Mysql
// @Summary Delete mysql database
// @Description 删除 mysql 数据库
// @Accept json
// @Param request body dto.MysqlDBDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"删除 mysql 数据库 [name]","formatEN":"delete mysql database [name]"}
func (b *BaseApi) DeleteMysql(c *gin.Context) {
	var req dto.MysqlDBDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	tx, ctx := helper.GetTxAndContext()
	if err := mysqlService.Delete(ctx, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		tx.Rollback()
		return
	}
	tx.Commit()
	helper.SuccessWithData(c, nil)
}

// @Tags Database Mysql
// @Summary Load mysql base info
// @Description 获取 mysql 基础信息
// @Success 200 {object} dto.DBBaseInfo
// @Security ApiKeyAuth
// @Router /databases/baseinfo [get]
func (b *BaseApi) LoadBaseinfo(c *gin.Context) {
	data, err := mysqlService.LoadBaseInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Mysql
// @Summary Load mysql remote access
// @Description 获取 mysql 远程访问权限
// @Success 200 {boolean} isRemote
// @Security ApiKeyAuth
// @Router /databases/remote [get]
func (b *BaseApi) LoadRemoteAccess(c *gin.Context) {
	isRemote, err := mysqlService.LoadRemoteAccess()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, isRemote)
}

// @Tags Database Mysql
// @Summary Load mysql status info
// @Description 获取 mysql 状态信息
// @Success 200 {object} dto.MysqlStatus
// @Security ApiKeyAuth
// @Router /databases/status [get]
func (b *BaseApi) LoadStatus(c *gin.Context) {
	data, err := mysqlService.LoadStatus()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Mysql
// @Summary Load mysql variables info
// @Description 获取 mysql 性能参数信息
// @Success 200 {object} dto.MysqlVariables
// @Security ApiKeyAuth
// @Router /databases/variables [get]
func (b *BaseApi) LoadVariables(c *gin.Context) {
	data, err := mysqlService.LoadVariables()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}
