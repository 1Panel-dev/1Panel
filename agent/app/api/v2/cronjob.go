package v2

import (
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/gin-gonic/gin"
)

// @Tags Cronjob
// @Summary Create cronjob
// @Accept json
// @Param request body dto.CronjobOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs [post]
// @x-panel-log {"bodyKeys":["type","name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建计划任务 [type][name]","formatEN":"create cronjob [type][name]"}
func (b *BaseApi) CreateCronjob(c *gin.Context) {
	var req dto.CronjobOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.Create(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Cronjob
// @Summary Load cronjob info
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/load/info [post]
func (b *BaseApi) LoadCronjobInfo(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := cronjobService.LoadInfo(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Cronjob
// @Summary Export cronjob list
// @Accept json
// @Param request body dto.OperateByIDs true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/export [post]
func (b *BaseApi) ExportCronjob(c *gin.Context) {
	var req dto.OperateByIDs
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	content, err := cronjobService.Export(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	http.ServeContent(c.Writer, c.Request, "", time.Now(), strings.NewReader(content))
}

// @Tags Cronjob
// @Summary Import cronjob list
// @Accept json
// @Param request body dto.CronjobImport true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/import [post]
func (b *BaseApi) ImportCronjob(c *gin.Context) {
	var req dto.CronjobImport
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.Import(req.Cronjobs); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Cronjob
// @Summary Load script options
// @Success 200 {array} dto.ScriptOptions
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/script/options [get]
func (b *BaseApi) LoadScriptOptions(c *gin.Context) {
	helper.SuccessWithData(c, cronjobService.LoadScriptOptions())
}

// @Tags Cronjob
// @Summary Load cronjob spec time
// @Accept json
// @Param request body dto.CronjobSpec true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/next [post]
func (b *BaseApi) LoadNextHandle(c *gin.Context) {
	var req dto.CronjobSpec
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	list, err := cronjobService.LoadNextHandle(req.Spec)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Cronjob
// @Summary Page cronjobs
// @Accept json
// @Param request body dto.PageCronjob true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/search [post]
func (b *BaseApi) SearchCronjob(c *gin.Context) {
	var req dto.PageCronjob
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := cronjobService.SearchWithPage(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Cronjob
// @Summary Page job records
// @Accept json
// @Param request body dto.SearchRecord true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/search/records [post]
func (b *BaseApi) SearchJobRecords(c *gin.Context) {
	var req dto.SearchRecord
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	loc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	req.StartTime = req.StartTime.In(loc)
	req.EndTime = req.EndTime.In(loc)

	total, list, err := cronjobService.SearchRecords(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Cronjob
// @Summary Load Cronjob record log
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/records/log [post]
func (b *BaseApi) LoadRecordLog(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content := cronjobService.LoadRecordLog(req)
	helper.SuccessWithData(c, content)
}

// @Tags Cronjob
// @Summary Clean job records
// @Accept json
// @Param request body dto.CronjobClean true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/records/clean [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"cronjobs","output_column":"name","output_value":"name"}],"formatZH":"清空计划任务记录 [name]","formatEN":"clean cronjob [name] records"}
func (b *BaseApi) CleanRecord(c *gin.Context) {
	var req dto.CronjobClean
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.CleanRecord(req); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Cronjob
// @Summary Delete cronjob
// @Accept json
// @Param request body dto.CronjobBatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"cronjobs","output_column":"name","output_value":"names"}],"formatZH":"删除计划任务 [names]","formatEN":"delete cronjob [names]"}
func (b *BaseApi) DeleteCronjob(c *gin.Context) {
	var req dto.CronjobBatchDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.Delete(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Cronjob
// @Summary Update cronjob
// @Accept json
// @Param request body dto.CronjobOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"cronjobs","output_column":"name","output_value":"name"}],"formatZH":"更新计划任务 [name]","formatEN":"update cronjob [name]"}
func (b *BaseApi) UpdateCronjob(c *gin.Context) {
	var req dto.CronjobOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.Update(req.ID, req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Cronjob
// @Summary Update cronjob group
// @Accept json
// @Param request body dto.ChangeGroup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/group/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"cronjobs","output_column":"name","output_value":"name"}],"formatZH":"更新计划任务分组 [name]","formatEN":"update cronjob group [name]"}
func (b *BaseApi) UpdateCronjobGroup(c *gin.Context) {
	var req dto.ChangeGroup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.UpdateGroup(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Cronjob
// @Summary Update cronjob status
// @Accept json
// @Param request body dto.CronjobUpdateStatus true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/status [post]
// @x-panel-log {"bodyKeys":["id","status"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"cronjobs","output_column":"name","output_value":"name"}],"formatZH":"修改计划任务 [name] 状态为 [status]","formatEN":"change the status of cronjob [name] to [status]."}
func (b *BaseApi) UpdateCronjobStatus(c *gin.Context) {
	var req dto.CronjobUpdateStatus
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.UpdateStatus(req.ID, req.Status); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Cronjob
// @Summary Download cronjob records
// @Accept json
// @Param request body dto.CronjobDownload true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/download [post]
// @x-panel-log {"bodyKeys":["recordID"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"recordID","isList":false,"db":"job_records","output_column":"file","output_value":"file"}],"formatZH":"下载计划任务记录 [file]","formatEN":"download the cronjob record [file]"}
func (b *BaseApi) TargetDownload(c *gin.Context) {
	var req dto.CronjobDownload
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	filePath, err := cronjobService.Download(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	c.File(filePath)
}

// @Tags Cronjob
// @Summary Handle cronjob once
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /cronjobs/handle [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"cronjobs","output_column":"name","output_value":"name"}],"formatZH":"手动执行计划任务 [name]","formatEN":"manually execute the cronjob [name]"}
func (b *BaseApi) HandleOnce(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := cronjobService.HandleOnce(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
