package v1

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags Cronjob
// @Summary Create cronjob
// @Description 创建计划任务
// @Accept json
// @Param request body dto.CronjobCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /cronjobs [post]
// @x-panel-log {"bodyKeys":["type","name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建计划任务 [type][name]","formatEN":"create cronjob [type][name]"}
func (b *BaseApi) CreateCronjob(c *gin.Context) {
	var req dto.CronjobCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := cronjobService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Cronjob
// @Summary Page cronjobs
// @Description 获取计划任务分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /cronjobs/search [post]
func (b *BaseApi) SearchCronjob(c *gin.Context) {
	var req dto.SearchWithPage
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := cronjobService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Cronjob
// @Summary Page job records
// @Description 获取计划任务记录
// @Accept json
// @Param request body dto.SearchRecord true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /cronjobs/search/records [post]
func (b *BaseApi) SearchJobRecords(c *gin.Context) {
	var req dto.SearchRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	req.StartTime = req.StartTime.Add(8 * time.Hour)
	req.EndTime = req.EndTime.Add(8 * time.Hour)

	total, list, err := cronjobService.SearchRecords(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Cronjob
// @Summary Clean job records
// @Description 清空计划任务记录
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /cronjobs/records/clean [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"cronjobs","output_colume":"name","output_value":"name"}],"formatZH":"清空计划任务记录 [name]","formatEN":"clean cronjob [name] records"}
func (b *BaseApi) CleanRecord(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := cronjobService.CleanRecord(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Cronjob
// @Summary Delete cronjob
// @Description 删除计划任务
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /cronjobs/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"ids","isList":true,"db":"cronjobs","output_colume":"name","output_value":"names"}],"formatZH":"删除计划任务 [names]","formatEN":"delete cronjob [names]"}
func (b *BaseApi) DeleteCronjob(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := cronjobService.Delete(req.Ids); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Cronjob
// @Summary Update cronjob
// @Description 更新计划任务
// @Accept json
// @Param request body dto.CronjobUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /cronjobs/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"cronjobs","output_colume":"name","output_value":"name"}],"formatZH":"更新计划任务 [name]","formatEN":"update cronjob [name]"}
func (b *BaseApi) UpdateCronjob(c *gin.Context) {
	var req dto.CronjobUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := cronjobService.Update(req.ID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Cronjob
// @Summary Update cronjob status
// @Description 更新计划任务状态
// @Accept json
// @Param request body dto.CronjobUpdateStatus true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /cronjobs/status [post]
// @x-panel-log {"bodyKeys":["id","status"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"cronjobs","output_colume":"name","output_value":"name"}],"formatZH":"修改计划任务 [name] 状态为 [status]","formatEN":"change the status of cronjob [name] to [status]."}
func (b *BaseApi) UpdateCronjobStatus(c *gin.Context) {
	var req dto.CronjobUpdateStatus
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := cronjobService.UpdateStatus(req.ID, req.Status); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Cronjob
// @Summary Download cronjob records
// @Description 下载计划任务记录
// @Accept json
// @Param request body dto.CronjobDownload true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /cronjobs/download [post]
// @x-panel-log {"bodyKeys":["recordID"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"recordID","isList":false,"db":"job_records","output_colume":"file","output_value":"file"}],"formatZH":"下载计划任务记录 [file]","formatEN":"download the cronjob record [file]"}
func (b *BaseApi) TargetDownload(c *gin.Context) {
	var req dto.CronjobDownload
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	filePath, err := cronjobService.Download(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, filePath)
}

// @Tags Cronjob
// @Summary Handle cronjob once
// @Description 手动执行计划任务
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /cronjobs/handle [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"cronjobs","output_colume":"name","output_value":"name"}],"formatZH":"手动执行计划任务 [name]","formatEN":"manually execute the cronjob [name]"}
func (b *BaseApi) HandleOnce(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := cronjobService.HandleOnce(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
