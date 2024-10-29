package v2

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Backup Account
// @Summary Create backup account
// @Description 创建备份账号
// @Accept json
// @Param request body dto.BackupOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /core/backup [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建备份账号 [type]","formatEN":"create backup account [type]"}
func (b *BaseApi) CreateBackup(c *gin.Context) {
	var req dto.BackupOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := backupService.Create(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Refresh OneDrive token
// @Description 刷新 OneDrive token
// @Success 200
// @Security ApiKeyAuth
// @Router /core/backup/refresh/onedrive [post]
func (b *BaseApi) RefreshOneDriveToken(c *gin.Context) {
	backupService.Run()
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary List buckets
// @Description 获取 bucket 列表
// @Accept json
// @Param request body dto.ForBuckets true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Router /core/backup/search [post]
func (b *BaseApi) ListBuckets(c *gin.Context) {
	var req dto.ForBuckets
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	buckets, err := backupService.GetBuckets(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, buckets)
}

// @Tags Backup Account
// @Summary Load OneDrive info
// @Description 获取 OneDrive 信息
// @Accept json
// @Success 200 {object} dto.OneDriveInfo
// @Security ApiKeyAuth
// @Router /core/backup/onedrive [get]
func (b *BaseApi) LoadOneDriveInfo(c *gin.Context) {
	data, err := backupService.LoadOneDriveInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Backup Account
// @Summary Load backup account options
// @Description 获取备份账号选项
// @Accept json
// @Success 200 {array} dto.BackupOption
// @Security ApiKeyAuth
// @Router /core/backup/options [get]
func (b *BaseApi) LoadBackupOptions(c *gin.Context) {
	list, err := backupService.LoadBackupOptions()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Backup Account
// @Summary Delete backup account
// @Description 删除备份账号
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /core/backup/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"backup_accounts","output_column":"type","output_value":"types"}],"formatZH":"删除备份账号 [types]","formatEN":"delete backup account [types]"}
func (b *BaseApi) DeleteBackup(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := backupService.Delete(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Update backup account
// @Description 更新备份账号信息
// @Accept json
// @Param request body dto.BackupOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /core/backup/update [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新备份账号 [types]","formatEN":"update backup account [types]"}
func (b *BaseApi) UpdateBackup(c *gin.Context) {
	var req dto.BackupOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := backupService.Update(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Search backup accounts with page
// @Description 获取备份账号列表
// @Accept json
// @Param request body dto.SearchPageWithType true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /core/backup/search [post]
func (b *BaseApi) SearchBackup(c *gin.Context) {
	var req dto.SearchPageWithType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := backupService.SearchWithPage(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Backup Account
// @Summary get local backup dir
// @Description 获取本地备份目录
// @Success 200
// @Security ApiKeyAuth
// @Router /core/backup/local [get]
func (b *BaseApi) GetLocalDir(c *gin.Context) {
	dir, err := backupService.GetLocalDir()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dir)
}
