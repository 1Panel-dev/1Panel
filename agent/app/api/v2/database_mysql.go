package v2

import (
	"context"
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Database Mysql
// @Summary Create mysql database
// @Accept json
// @Param request body dto.MysqlDBCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 mysql 数据库 [name]","formatEN":"create mysql database [name]"}
func (b *BaseApi) CreateMysql(c *gin.Context) {
	var req dto.MysqlDBCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		req.Password = string(password)
	}

	if _, err := mysqlService.Create(context.Background(), req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Bind user of mysql database
// @Accept json
// @Param request body dto.BindUser true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/bind [post]
// @x-panel-log {"bodyKeys":["database", "username"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"绑定 mysql 数据库名 [database] [username]","formatEN":"bind mysql database [database] [username]"}
func (b *BaseApi) BindUser(c *gin.Context) {
	var req dto.BindUser
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		req.Password = string(password)
	}

	if err := mysqlService.BindUser(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Update mysql database description
// @Accept json
// @Param request body dto.UpdateDescription true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/description/update [post]
// @x-panel-log {"bodyKeys":["id","description"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"mysql 数据库 [name] 描述信息修改 [description]","formatEN":"The description of the mysql database [name] is modified => [description]"}
func (b *BaseApi) UpdateMysqlDescription(c *gin.Context) {
	var req dto.UpdateDescription
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := mysqlService.UpdateDescription(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Change mysql password
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/change/password [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"更新数据库 [name] 密码","formatEN":"Update database [name] password"}
func (b *BaseApi) ChangeMysqlPassword(c *gin.Context) {
	var req dto.ChangeDBInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.Value) != 0 {
		value, err := base64.StdEncoding.DecodeString(req.Value)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		req.Value = string(value)
	}

	if err := mysqlService.ChangePassword(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Change mysql access
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/change/access [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"更新数据库 [name] 访问权限","formatEN":"Update database [name] access"}
func (b *BaseApi) ChangeMysqlAccess(c *gin.Context) {
	var req dto.ChangeDBInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := mysqlService.ChangeAccess(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Update mysql variables
// @Accept json
// @Param request body dto.MysqlVariablesUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/variables/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"调整 mysql 数据库性能参数","formatEN":"adjust mysql database performance parameters"}
func (b *BaseApi) UpdateMysqlVariables(c *gin.Context) {
	var req dto.MysqlVariablesUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := mysqlService.UpdateVariables(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Page mysql databases
// @Accept json
// @Param request body dto.MysqlDBSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/search [post]
func (b *BaseApi) SearchMysql(c *gin.Context) {
	var req dto.MysqlDBSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := mysqlService.SearchWithPage(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Database Mysql
// @Summary List mysql database names
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200 {array} dto.MysqlOption
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/options [get]
func (b *BaseApi) ListDBName(c *gin.Context) {
	list, err := mysqlService.ListDBOption()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Database Mysql
// @Summary Load mysql database from remote
// @Accept json
// @Param request body dto.MysqlLoadDB true "request"
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/load [post]
func (b *BaseApi) LoadDBFromRemote(c *gin.Context) {
	var req dto.MysqlLoadDB
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := mysqlService.LoadFromRemote(req); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Check before delete mysql database
// @Accept json
// @Param request body dto.MysqlDBDeleteCheck true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/del/check [post]
func (b *BaseApi) DeleteCheckMysql(c *gin.Context) {
	var req dto.MysqlDBDeleteCheck
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	apps, err := mysqlService.DeleteCheck(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, apps)
}

// @Tags Database Mysql
// @Summary Delete mysql database
// @Accept json
// @Param request body dto.MysqlDBDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mysqls","output_column":"name","output_value":"name"}],"formatZH":"删除 mysql 数据库 [name]","formatEN":"delete mysql database [name]"}
func (b *BaseApi) DeleteMysql(c *gin.Context) {
	var req dto.MysqlDBDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	tx, ctx := helper.GetTxAndContext()
	if err := mysqlService.Delete(ctx, req); err != nil {
		helper.InternalServer(c, err)
		tx.Rollback()
		return
	}
	tx.Commit()
	helper.SuccessWithOutData(c)
}

// @Tags Database Mysql
// @Summary Load mysql remote access
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {boolean} isRemote
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/remote [post]
func (b *BaseApi) LoadRemoteAccess(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	isRemote, err := mysqlService.LoadRemoteAccess(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, isRemote)
}

// @Tags Database Mysql
// @Summary Load mysql status info
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {object} dto.MysqlStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/status [post]
func (b *BaseApi) LoadStatus(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := mysqlService.LoadStatus(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Mysql
// @Summary Load mysql variables info
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {object} dto.MysqlVariables
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/variables [post]
func (b *BaseApi) LoadVariables(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := mysqlService.LoadVariables(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, data)
}
