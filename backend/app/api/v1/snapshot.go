package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags System Setting
// @Summary Create system backup
// @Description 创建系统快照
// @Accept json
// @Param request body dto.SnapshotCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/snapshot [post]
// @x-panel-log {"bodyKeys":["from", "description"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建系统快照 [description] 到 [from]","formatEN":"Create system backup [description] to [from]"}
func (b *BaseApi) CreateSnapshot(c *gin.Context) {
	var req dto.SnapshotCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := snapshotService.SnapshotCreate(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Page system snapshot
// @Description 获取系统快照列表分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /settings/snapshot/search [post]
func (b *BaseApi) SearchSnapshot(c *gin.Context) {
	var req dto.SearchWithPage
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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
// @Summary Recover system backup
// @Description 从系统快照恢复
// @Accept json
// @Param request body dto.SnapshotRecover true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/snapshot/recover [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"snapshots","output_colume":"name","output_value":"name"}],"formatZH":"从系统快照 [name] 恢复","formatEN":"Recover from system backup [name]"}
func (b *BaseApi) RecoverSnapshot(c *gin.Context) {
	var req dto.SnapshotRecover
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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
// @Description 从系统快照回滚
// @Accept json
// @Param request body dto.SnapshotRecover true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/snapshot/rollback [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"snapshots","output_colume":"name","output_value":"name"}],"formatZH":"从系统快照 [name] 回滚","formatEN":"Rollback from system backup [name]"}
func (b *BaseApi) RollbackSnapshot(c *gin.Context) {
	var req dto.SnapshotRecover
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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
// @Description 删除系统快照
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/snapshot/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"ids","isList":true,"db":"snapshots","output_colume":"name","output_value":"name"}],"formatZH":"删除系统快照 [name]","formatEN":"Delete system backup [name]"}
func (b *BaseApi) DeleteSnapshot(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := snapshotService.Delete(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
