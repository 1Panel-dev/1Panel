package v1

import (
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Database Redis
// @Summary Load redis status info
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200 {object} dto.RedisStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/redis/status [post]
func (b *BaseApi) LoadRedisStatus(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}
	data, err := redisService.LoadStatus(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Redis
// @Summary Load redis conf
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200 {object} dto.RedisConf
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/redis/conf [post]
func (b *BaseApi) LoadRedisConf(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}
	data, err := redisService.LoadConf(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Redis
// @Summary Load redis persistence conf
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200 {object} dto.RedisPersistence
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/redis/persistence/conf [post]
func (b *BaseApi) LoadPersistenceConf(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}
	data, err := redisService.LoadPersistenceConf(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

func (b *BaseApi) CheckHasCli(c *gin.Context) {
	helper.SuccessWithData(c, redisService.CheckHasCli())
}

// @Tags Database Redis
// @Summary Install redis-cli
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/redis/install/cli [post]
func (b *BaseApi) InstallCli(c *gin.Context) {
	if err := redisService.InstallCli(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithOutData(c)
}

// @Tags Database Redis
// @Summary Update redis conf
// @Accept json
// @Param request body dto.RedisConfUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/redis/conf/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 redis 数据库配置信息","formatEN":"update the redis database configuration information"}
func (b *BaseApi) UpdateRedisConf(c *gin.Context) {
	var req dto.RedisConfUpdate
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}

	if err := redisService.UpdateConf(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Redis
// @Summary Change redis password
// @Accept json
// @Param request body dto.ChangeRedisPass true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/redis/password [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改 redis 数据库密码","formatEN":"change the password of the redis database"}
func (b *BaseApi) ChangeRedisPassword(c *gin.Context) {
	var req dto.ChangeRedisPass
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

	if err := redisService.ChangePassword(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Redis
// @Summary Update redis persistence conf
// @Accept json
// @Param request body dto.RedisConfPersistenceUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /databases/redis/persistence/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"redis 数据库持久化配置更新","formatEN":"redis database persistence configuration update"}
func (b *BaseApi) UpdateRedisPersistenceConf(c *gin.Context) {
	var req dto.RedisConfPersistenceUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := redisService.UpdatePersistenceConf(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
