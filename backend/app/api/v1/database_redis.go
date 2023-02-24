package v1

import (
	"bufio"
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/gin-gonic/gin"
)

// @Tags Database Redis
// @Summary Load redis status info
// @Description 获取 redis 状态信息
// @Success 200 {object} dto.RedisStatus
// @Security ApiKeyAuth
// @Router /databases/redis/status [get]
func (b *BaseApi) LoadRedisStatus(c *gin.Context) {
	data, err := redisService.LoadStatus()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Redis
// @Summary Load redis conf
// @Description 获取 redis 配置信息
// @Success 200 {object} dto.RedisConf
// @Security ApiKeyAuth
// @Router /databases/redis/conf [get]
func (b *BaseApi) LoadRedisConf(c *gin.Context) {
	data, err := redisService.LoadConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Redis
// @Summary Load redis persistence conf
// @Description 获取 redis 持久化配置
// @Success 200 {object} dto.RedisPersistence
// @Security ApiKeyAuth
// @Router /databases/redis/persistence/conf [get]
func (b *BaseApi) LoadPersistenceConf(c *gin.Context) {
	data, err := redisService.LoadPersistenceConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Database Redis
// @Summary Update redis conf
// @Description 更新 redis 配置信息
// @Accept json
// @Param request body dto.RedisConfUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/redis/conf/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新 redis 数据库配置信息","formatEN":"update the redis database configuration information"}
func (b *BaseApi) UpdateRedisConf(c *gin.Context) {
	var req dto.RedisConfUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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
// @Description 更新 redis 密码
// @Accept json
// @Param request body dto.ChangeDBInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/redis/password [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"修改 redis 数据库密码","formatEN":"change the password of the redis database"}
func (b *BaseApi) ChangeRedisPassword(c *gin.Context) {
	var req dto.ChangeDBInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := redisService.ChangePassword(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Redis
// @Summary Update redis persistence conf
// @Description 更新 redis 持久化配置
// @Accept json
// @Param request body dto.RedisConfPersistenceUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/redis/persistence/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"redis 数据库持久化配置更新","formatEN":"redis database persistence configuration update"}
func (b *BaseApi) UpdateRedisPersistenceConf(c *gin.Context) {
	var req dto.RedisConfPersistenceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := redisService.UpdatePersistenceConf(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Redis
// @Summary Backup redis
// @Description 备份 redis 数据库
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/redis/backup [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"备份 redis 数据库","formatEN":"backup redis database"}
func (b *BaseApi) RedisBackup(c *gin.Context) {
	if err := redisService.Backup(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Redis
// @Summary Recover redis
// @Description 恢复 redis 数据库
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/redis/recover [post]
// @x-panel-log {"bodyKeys":["fileDir","fileName"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"redis 数据库从 [fileDir]/[fileName] 恢复","formatEN":"redis database recover from [fileDir]/[fileName]"}
func (b *BaseApi) RedisRecover(c *gin.Context) {
	var req dto.RedisBackupRecover
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := redisService.Recover(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Database Redis
// @Summary Page redis backups
// @Description 获取 redis 备份记录分页
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /databases/redis/backup/search [post]
func (b *BaseApi) RedisBackupList(c *gin.Context) {
	var req dto.PageInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := redisService.SearchBackupListWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Database Redis
// @Summary Update redis conf by file
// @Description 上传更新 redis 配置信息
// @Accept json
// @Param request body dto.RedisConfUpdateByFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /databases/redis/conffile/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新 redis 数据库配置信息","formatEN":"update the redis database configuration information"}
func (b *BaseApi) UpdateRedisConfByFile(c *gin.Context) {
	var req dto.RedisConfUpdateByFile
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	redisInfo, err := redisService.LoadConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	path := fmt.Sprintf("%s/redis/%s/conf/redis.conf", constant.AppInstallDir, redisInfo.Name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	if req.RestartNow {
		composeDir := fmt.Sprintf("%s/redis/%s/docker-compose.yml", constant.AppInstallDir, redisInfo.Name)
		if _, err := compose.Restart(composeDir); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	}

	helper.SuccessWithData(c, nil)
}
