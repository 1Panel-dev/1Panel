package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags System Setting
// @Summary Create system snapshot
// @Accept json
// @Param request body dto.SnapshotCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot [post]
// @x-panel-log {"bodyKeys":["from", "description"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建系统快照 [description] 到 [from]","formatEN":"Create system backup [description] to [from]"}
func (b *BaseApi) CreateSnapshot(c *gin.Context) {
	var req dto.SnapshotCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := snapshotService.SnapshotCreate(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Import system snapshot
// @Accept json
// @Param request body dto.SnapshotImport true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/import [post]
// @x-panel-log {"bodyKeys":["from", "names"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"从 [from] 同步系统快照 [names]","formatEN":"Sync system snapshots [names] from [from]"}
func (b *BaseApi) ImportSnapshot(c *gin.Context) {
	var req dto.SnapshotImport
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := snapshotService.SnapshotImport(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Load Snapshot status
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200 {object} dto.SnapshotStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/status [post]
func (b *BaseApi) LoadSnapShotStatus(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := snapshotService.LoadSnapShotStatus(req.ID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags System Setting
// @Summary Update snapshot description
// @Accept json
// @Param request body dto.UpdateDescription true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/description/update [post]
// @x-panel-log {"bodyKeys":["id","description"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"snapshots","output_column":"name","output_value":"name"}],"formatZH":"快照 [name] 描述信息修改 [description]","formatEN":"The description of the snapshot [name] is modified => [description]"}
func (b *BaseApi) UpdateSnapDescription(c *gin.Context) {
	var req dto.UpdateDescription
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := snapshotService.UpdateDescription(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Page system snapshot
// @Accept json
// @Param request body dto.PageSnapshot true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/search [post]
func (b *BaseApi) SearchSnapshot(c *gin.Context) {
	var req dto.PageSnapshot
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, accounts, err := snapshotService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: accounts,
	})
}

// @Tags System Setting
// @Summary Load system snapshot size
// @Accept json
// @Param request body dto.PageSnapshot true "request"
// @Success 200 {array} dto.SnapshotFile
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/size [post]
func (b *BaseApi) LoadSnapshotSize(c *gin.Context) {
	var req dto.PageSnapshot
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := snapshotService.LoadSize(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags System Setting
// @Summary Recover system backup
// @Accept json
// @Param request body dto.SnapshotRecover true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/recover [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"snapshots","output_column":"name","output_value":"name"}],"formatZH":"从系统快照 [name] 恢复","formatEN":"Recover from system backup [name]"}
func (b *BaseApi) RecoverSnapshot(c *gin.Context) {
	var req dto.SnapshotRecover
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := snapshotService.SnapshotRecover(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Rollback system backup
// @Accept json
// @Param request body dto.SnapshotRecover true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/rollback [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"snapshots","output_column":"name","output_value":"name"}],"formatZH":"从系统快照 [name] 回滚","formatEN":"Rollback from system backup [name]"}
func (b *BaseApi) RollbackSnapshot(c *gin.Context) {
	var req dto.SnapshotRecover
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := snapshotService.SnapshotRollback(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Delete system backup
// @Accept json
// @Param request body dto.SnapshotBatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/snapshot/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"snapshots","output_column":"name","output_value":"name"}],"formatZH":"删除系统快照 [name]","formatEN":"Delete system backup [name]"}
func (b *BaseApi) DeleteSnapshot(c *gin.Context) {
	var req dto.SnapshotBatchDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := snapshotService.Delete(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
