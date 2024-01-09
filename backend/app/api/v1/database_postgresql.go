package v1

import (
	"context"
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Database Postgresql
// @Summary Create postgresql database
// @Description 创建 postgresql 数据库
// @Accept json
// @Param request body dto.PostgresqlDBCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/pg [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 postgresql 数据库 [name]","formatEN":"create postgresql database [name]"}
func (b *BaseApi) CreatePostgresql(c *gin.Context) {
	var req dto.PostgresqlDBCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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

	if _, err := postgresqlService.Create(context.Background(), req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Postgresql
// @Summary Bind postgresql user
// @Description 绑定 postgresql 数据库用户
// @Accept json
// @Param request body dto.PostgresqlBindUser true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/pg/bind [post]
// @x-panel-log {"bodyKeys":["name", "username"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"绑定 postgresql 数据库 [name] 用户 [username]","formatEN":"bind postgresql database [name] user [username]"}
func (b *BaseApi) BindPostgresqlUser(c *gin.Context) {
	var req dto.PostgresqlBindUser
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := postgresqlService.BindUser(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Postgresql
// @Summary Update postgresql database description
// @Description 更新 postgresql 数据库库描述信息
// @Accept json
// @Param request body dto.UpdateDescription true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/pg/description [post]
// @x-panel-log {"bodyKeys":["id","description"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_postgresqls","output_column":"name","output_value":"name"}],"formatZH":"postgresql 数据库 [name] 描述信息修改 [description]","formatEN":"The description of the postgresql database [name] is modified => [description]"}
func (b *BaseApi) UpdatePostgresqlDescription(c *gin.Context) {
	var req dto.UpdateDescription
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := postgresqlService.UpdateDescription(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Postgresql
// @Summary Change postgresql privileges
// @Description 修改 postgresql 用户权限
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/pg/privileges [post]
// @x-panel-log {"bodyKeys":["database", "username"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新数据库 [database] 用户 [username] 权限","formatEN":"Update [user] privileges of database [database]"}
func (b *BaseApi) ChangePostgresqlPrivileges(c *gin.Context) {
	var req dto.PostgresqlPrivileges
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := postgresqlService.ChangePrivileges(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Postgresql
// @Summary Change postgresql password
// @Description 修改 postgresql 密码
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/pg/password [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_postgresqls","output_column":"name","output_value":"name"}],"formatZH":"更新数据库 [name] 密码","formatEN":"Update database [name] password"}
func (b *BaseApi) ChangePostgresqlPassword(c *gin.Context) {
	var req dto.ChangeDBInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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

	if err := postgresqlService.ChangePassword(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Postgresql
// @Summary Page postgresql databases
// @Description 获取 postgresql 数据库列表分页
// @Accept json
// @Param request body dto.PostgresqlDBSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /databases/pg/search [post]
func (b *BaseApi) SearchPostgresql(c *gin.Context) {
	var req dto.PostgresqlDBSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := postgresqlService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Database Postgresql
// @Summary List postgresql database names
// @Description 获取 postgresql 数据库列表
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200 {array} dto.PostgresqlOption
// @Security ApiKeyAuth
// @Router /databases/pg/options [get]
func (b *BaseApi) ListPostgresqlDBName(c *gin.Context) {
	list, err := postgresqlService.ListDBOption()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Database Postgresql
// @Summary Load postgresql database from remote
// @Description 从服务器获取
// @Accept json
// @Param request body dto.PostgresqlLoadDB true "request"
// @Security ApiKeyAuth
// @Router /databases/pg/:database/load [post]
func (b *BaseApi) LoadPostgresqlDBFromRemote(c *gin.Context) {
	database, err := helper.GetStrParamByKey(c, "database")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := postgresqlService.LoadFromRemote(database); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Database Postgresql
// @Summary Check before delete postgresql database
// @Description Postgresql 数据库删除前检查
// @Accept json
// @Param request body dto.PostgresqlDBDeleteCheck true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Router /databases/pg/del/check [post]
func (b *BaseApi) DeleteCheckPostgresql(c *gin.Context) {
	var req dto.PostgresqlDBDeleteCheck
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	apps, err := postgresqlService.DeleteCheck(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, apps)
}

// @Tags Database Postgresql
// @Summary Delete postgresql database
// @Description 删除 postgresql 数据库
// @Accept json
// @Param request body dto.PostgresqlDBDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/pg/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_postgresqls","output_column":"name","output_value":"name"}],"formatZH":"删除 postgresql 数据库 [name]","formatEN":"delete postgresql database [name]"}
func (b *BaseApi) DeletePostgresql(c *gin.Context) {
	var req dto.PostgresqlDBDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	tx, ctx := helper.GetTxAndContext()
	if err := postgresqlService.Delete(ctx, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		tx.Rollback()
		return
	}
	tx.Commit()
	helper.SuccessWithData(c, nil)
}
