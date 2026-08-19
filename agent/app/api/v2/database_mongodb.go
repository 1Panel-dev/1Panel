package v2

import (
	"context"
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Database Mongodb
// @Summary Create mongodb database
// @Accept json
// @Param request body dto.MongodbDBCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 mongodb 数据库 [name]","formatEN":"create mongodb database [name]"}
func (b *BaseApi) CreateMongodb(c *gin.Context) {
	var req dto.MongodbDBCreate
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

	if _, err := mongodbService.Create(context.Background(), req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Database Mongodb
// @Summary Page mongodb databases
// @Accept json
// @Param request body dto.MongodbDBSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/search [post]
func (b *BaseApi) SearchMongodb(c *gin.Context) {
	var req dto.MongodbDBSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := mongodbService.SearchWithPage(req, helper.IsDemoRequest(c))
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Database Mongodb
// @Summary Update mongodb database description
// @Accept json
// @Param request body dto.UpdateDescription true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/description [post]
// @x-panel-log {"bodyKeys":["id","description"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mongodbs","output_column":"name","output_value":"name"}],"formatZH":"mongodb 数据库 [name] 描述信息修改 [description]","formatEN":"The description of the mongodb database [name] is modified => [description]"}
func (b *BaseApi) UpdateMongodbDescription(c *gin.Context) {
	var req dto.UpdateDescription
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := mongodbService.UpdateDescription(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Database Mongodb
// @Summary Load mongodb database from remote
// @Accept json
// @Param request body dto.MongodbLoadDB true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/load [post]
func (b *BaseApi) LoadMongodbFromRemote(c *gin.Context) {
	var req dto.MongodbLoadDB
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := mongodbService.LoadFromRemote(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Database Mongodb
// @Summary Bind mongodb database user info
// @Accept json
// @Param request body dto.MongodbBind true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/bind [post]
// @x-panel-log {"bodyKeys":["database", "name", "username"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"绑定 mongodb 数据库 [database] [name] 用户 [username]","formatEN":"bind mongodb database [database] [name] user [username]"}
func (b *BaseApi) BindMongodbUser(c *gin.Context) {
	var req dto.MongodbBind
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

	if err := mongodbService.BindUser(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Database Mongodb
// @Summary Change mongodb database password
// @Accept json
// @Param request body dto.MongodbPassword true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/password [post]
// @x-panel-log {"bodyKeys":["database", "name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 mongodb 数据库 [database] [name] 密码","formatEN":"update mongodb database [database] [name] password"}
func (b *BaseApi) ChangeMongodbPassword(c *gin.Context) {
	var req dto.MongodbPassword
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

	if err := mongodbService.ChangePassword(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Database Mongodb
// @Summary Change mongodb root password
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/root/password [post]
// @x-panel-log {"bodyKeys":["database"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 mongodb 数据库 [database] root 密码","formatEN":"update mongodb database [database] root password"}
func (b *BaseApi) ChangeMongodbRootPassword(c *gin.Context) {
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

	if err := mongodbService.ChangeRootPassword(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Database Mongodb
// @Summary Load mongodb privileges
// @Accept json
// @Param request body dto.MongodbPrivilegesLoad true "request"
// @Success 200 {string} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/privileges [post]
func (b *BaseApi) LoadMongodbPrivileges(c *gin.Context) {
	var req dto.MongodbPrivilegesLoad
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	permission, err := mongodbService.LoadPrivileges(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, permission)
}

// @Tags Database Mongodb
// @Summary Change mongodb privileges
// @Accept json
// @Param request body dto.MongodbPrivileges true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/privileges/change [post]
// @x-panel-log {"bodyKeys":["database", "username"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 mongodb 数据库 [database] 用户 [username] 权限","formatEN":"update mongodb database [database] user [username] privileges"}
func (b *BaseApi) ChangeMongodbPrivileges(c *gin.Context) {
	var req dto.MongodbPrivileges
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := mongodbService.ChangePrivileges(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Database Mongodb
// @Summary Check before delete mongodb database
// @Accept json
// @Param request body dto.MongodbDBDeleteCheck true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/del/check [post]
func (b *BaseApi) DeleteCheckMongodb(c *gin.Context) {
	var req dto.MongodbDBDeleteCheck
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	apps, err := mongodbService.DeleteCheck(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, apps)
}

// @Tags Database Mongodb
// @Summary Delete mongodb database
// @Accept json
// @Param request body dto.MongodbDBDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/mongodb/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"database_mongodbs","output_column":"name","output_value":"name"}],"formatZH":"删除 mongodb 数据库 [name]","formatEN":"delete mongodb database [name]"}
func (b *BaseApi) DeleteMongodb(c *gin.Context) {
	var req dto.MongodbDBDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	tx, ctx := helper.GetTxAndContext()
	if err := mongodbService.Delete(ctx, req); err != nil {
		helper.InternalServer(c, err)
		tx.Rollback()
		return
	}
	tx.Commit()
	helper.Success(c)
}
