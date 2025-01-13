package v1

import (
	"encoding/base64"
	"fmt"
	"path"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Backup Account
// @Summary Create backup account
// @Accept json
// @Param request body dto.BackupOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建备份账号 [type]","formatEN":"create backup account [type]"}
func (b *BaseApi) CreateBackup(c *gin.Context) {
	var req dto.BackupOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if len(req.Credential) != 0 {
		credential, err := base64.StdEncoding.DecodeString(req.Credential)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Credential = string(credential)
	}
	if len(req.AccessKey) != 0 {
		accessKey, err := base64.StdEncoding.DecodeString(req.AccessKey)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.AccessKey = string(accessKey)
	}

	if err := backupService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Refresh OneDrive token
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/refresh/onedrive [post]
func (b *BaseApi) RefreshOneDriveToken(c *gin.Context) {
	backupService.Run()
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary List buckets
// @Accept json
// @Param request body dto.ForBuckets true "request"
// @Success 200 {array} object
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/search [post]
func (b *BaseApi) ListBuckets(c *gin.Context) {
	var req dto.ForBuckets
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if len(req.Credential) != 0 {
		credential, err := base64.StdEncoding.DecodeString(req.Credential)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Credential = string(credential)
	}
	if len(req.AccessKey) != 0 {
		accessKey, err := base64.StdEncoding.DecodeString(req.AccessKey)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.AccessKey = string(accessKey)
	}

	buckets, err := backupService.GetBuckets(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, buckets)
}

// @Tags Backup Account
// @Summary Load OneDrive info
// @Accept json
// @Success 200 {object} dto.OneDriveInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/onedrive [get]
func (b *BaseApi) LoadOneDriveInfo(c *gin.Context) {
	data, err := backupService.LoadOneDriveInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Backup Account
// @Summary Delete backup account
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"backup_accounts","output_column":"type","output_value":"types"}],"formatZH":"删除备份账号 [types]","formatEN":"delete backup account [types]"}
func (b *BaseApi) DeleteBackup(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := backupService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Page backup records
// @Accept json
// @Param request body dto.RecordSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/record/search [post]
func (b *BaseApi) SearchBackupRecords(c *gin.Context) {
	var req dto.RecordSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := backupService.SearchRecordsWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Backup Account
// @Summary Load backup records size
// @Accept json
// @Param request body dto.RecordSearch true "request"
// @Success 200 {array} dto.BackupFile
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/record/size [post]
func (b *BaseApi) LoadBackupSize(c *gin.Context) {
	var req dto.RecordSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	list, err := backupService.LoadSize(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Backup Account
// @Summary Page backup records by cronjob
// @Accept json
// @Param request body dto.RecordSearchByCronjob true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/record/search/bycronjob [post]
func (b *BaseApi) SearchBackupRecordsByCronjob(c *gin.Context) {
	var req dto.RecordSearchByCronjob
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := backupService.SearchRecordsByCronjobWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Backup Account
// @Summary Load backup records size for cronjob
// @Accept json
// @Param request body dto.RecordSearchByCronjob true "request"
// @Success 200 {array} dto.BackupFile
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/record/size/bycronjob [post]
func (b *BaseApi) LoadBackupSizeByCronjob(c *gin.Context) {
	var req dto.RecordSearchByCronjob
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	list, err := backupService.LoadSizeByCronjob(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Backup Account
// @Summary Download backup record
// @Accept json
// @Param request body dto.DownloadRecord true "request"
// @Success 200 {string} filePath
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/record/download [post]
// @x-panel-log {"bodyKeys":["source","fileName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"下载备份记录 [source][fileName]","formatEN":"download backup records [source][fileName]"}
func (b *BaseApi) DownloadRecord(c *gin.Context) {
	var req dto.DownloadRecord
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	filePath, err := backupService.DownloadRecord(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, filePath)
}

// @Tags Backup Account
// @Summary Delete backup record
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/record/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"backup_records","output_column":"file_name","output_value":"files"}],"formatZH":"删除备份记录 [files]","formatEN":"delete backup records [files]"}
func (b *BaseApi) DeleteBackupRecord(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := backupService.BatchDeleteRecord(req.Ids); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Update backup account
// @Accept json
// @Param request body dto.BackupOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/update [post]
// @x-panel-log {"bodyKeys":["type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新备份账号 [types]","formatEN":"update backup account [types]"}
func (b *BaseApi) UpdateBackup(c *gin.Context) {
	var req dto.BackupOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.Credential) != 0 {
		credential, err := base64.StdEncoding.DecodeString(req.Credential)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Credential = string(credential)
	}
	if len(req.AccessKey) != 0 {
		accessKey, err := base64.StdEncoding.DecodeString(req.AccessKey)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.AccessKey = string(accessKey)
	}

	if err := backupService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary List backup accounts
// @Success 200 {array} dto.BackupInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/search [get]
func (b *BaseApi) ListBackup(c *gin.Context) {
	data, err := backupService.List()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Backup Account
// @Summary List files from backup accounts
// @Accept json
// @Param request body dto.BackupSearchFile true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/search/files [post]
func (b *BaseApi) LoadFilesFromBackup(c *gin.Context) {
	var req dto.BackupSearchFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data := backupService.ListFiles(req)
	helper.SuccessWithData(c, data)
}

// @Tags Backup Account
// @Summary Backup system data
// @Accept json
// @Param request body dto.CommonBackup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/backup [post]
// @x-panel-log {"bodyKeys":["type","name","detailName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"备份 [type] 数据 [name][detailName]","formatEN":"backup [type] data [name][detailName]"}
func (b *BaseApi) Backup(c *gin.Context) {
	var req dto.CommonBackup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	switch req.Type {
	case "app":
		if _, err := backupService.AppBackup(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "mysql", "mariadb":
		if err := backupService.MysqlBackup(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case constant.AppPostgresql:
		if err := backupService.PostgresqlBackup(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "website":
		if err := backupService.WebsiteBackup(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "redis":
		if err := backupService.RedisBackup(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Recover system data
// @Accept json
// @Param request body dto.CommonRecover true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/recover [post]
// @x-panel-log {"bodyKeys":["type","name","detailName","file"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"从 [file] 恢复 [type] 数据 [name][detailName]","formatEN":"recover [type] data [name][detailName] from [file]"}
func (b *BaseApi) Recover(c *gin.Context) {
	var req dto.CommonRecover
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	downloadPath, err := backupService.DownloadRecord(dto.DownloadRecord{Source: req.Source, FileDir: path.Dir(req.File), FileName: path.Base(req.File)})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, fmt.Errorf("download file failed, err: %v", err))
		return
	}
	req.File = downloadPath
	switch req.Type {
	case "mysql", "mariadb":
		if err := backupService.MysqlRecover(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case constant.AppPostgresql:
		if err := backupService.PostgresqlRecover(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "website":
		if err := backupService.WebsiteRecover(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "redis":
		if err := backupService.RedisRecover(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "app":
		if err := backupService.AppRecover(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Backup Account
// @Summary Recover system data by upload
// @Accept json
// @Param request body dto.CommonRecover true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/backup/recover/byupload [post]
// @x-panel-log {"bodyKeys":["type","name","detailName","file"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"从 [file] 恢复 [type] 数据 [name][detailName]","formatEN":"recover [type] data [name][detailName] from [file]"}
func (b *BaseApi) RecoverByUpload(c *gin.Context) {
	var req dto.CommonRecover
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	switch req.Type {
	case "mysql", "mariadb":
		if err := backupService.MysqlRecoverByUpload(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case constant.AppPostgresql:
		if err := backupService.PostgresqlRecoverByUpload(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "app":
		if err := backupService.AppRecover(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	case "website":
		if err := backupService.WebsiteRecover(req); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	}
	helper.SuccessWithData(c, nil)
}
