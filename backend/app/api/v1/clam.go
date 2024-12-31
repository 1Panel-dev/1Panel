package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Clam
// @Summary Create clam
// @Accept json
// @Param request body dto.ClamCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam [post]
// @x-panel-log {"bodyKeys":["name","path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建扫描规则 [name][path]","formatEN":"create clam [name][path]"}
func (b *BaseApi) CreateClam(c *gin.Context) {
	var req dto.ClamCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := clamService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Clam
// @Summary Update clam
// @Accept json
// @Param request body dto.ClamUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/update [post]
// @x-panel-log {"bodyKeys":["name","path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改扫描规则 [name][path]","formatEN":"update clam [name][path]"}
func (b *BaseApi) UpdateClam(c *gin.Context) {
	var req dto.ClamUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := clamService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Clam
// @Summary Update clam status
// @Accept json
// @Param request body dto.ClamUpdateStatus true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/status/update [post]
// @x-panel-log {"bodyKeys":["id","status"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"clams","output_column":"name","output_value":"name"}],"formatZH":"修改扫描规则 [name] 状态为 [status]","formatEN":"change the status of clam [name] to [status]."}
func (b *BaseApi) UpdateClamStatus(c *gin.Context) {
	var req dto.ClamUpdateStatus
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := clamService.UpdateStatus(req.ID, req.Status); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Clam
// @Summary Page clam
// @Accept json
// @Param request body dto.SearchClamWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/search [post]
func (b *BaseApi) SearchClam(c *gin.Context) {
	var req dto.SearchClamWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := clamService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Clam
// @Summary Load clam base info
// @Accept json
// @Success 200 {object} dto.ClamBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/base [get]
func (b *BaseApi) LoadClamBaseInfo(c *gin.Context) {
	info, err := clamService.LoadBaseInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, info)
}

// @Tags Clam
// @Summary Operate Clam
// @Accept json
// @Param request body dto.Operate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] Clam","formatEN":"[operation] FTP"}
func (b *BaseApi) OperateClam(c *gin.Context) {
	var req dto.Operate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := clamService.Operate(req.Operation); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Clam
// @Summary Clean clam record
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/record/clean [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":true,"db":"clams","output_column":"name","output_value":"name"}],"formatZH":"清空扫描报告 [name]","formatEN":"clean clam record [name]"}
func (b *BaseApi) CleanClamRecord(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := clamService.CleanRecord(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Clam
// @Summary Page clam record
// @Accept json
// @Param request body dto.ClamLogSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/record/search [post]
func (b *BaseApi) SearchClamRecord(c *gin.Context) {
	var req dto.ClamLogSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := clamService.LoadRecords(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Clam
// @Summary Load clam record detail
// @Accept json
// @Param request body dto.ClamLogReq true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/record/log [post]
func (b *BaseApi) LoadClamRecordLog(c *gin.Context) {
	var req dto.ClamLogReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	content, err := clamService.LoadRecordLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, content)
}

// @Tags Clam
// @Summary Load clam file
// @Accept json
// @Param request body dto.ClamFileReq true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/file/search [post]
func (b *BaseApi) SearchClamFile(c *gin.Context) {
	var req dto.ClamFileReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	content, err := clamService.LoadFile(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, content)
}

// @Tags Clam
// @Summary Update clam file
// @Accept json
// @Param request body dto.UpdateByNameAndFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/file/update [post]
func (b *BaseApi) UpdateFile(c *gin.Context) {
	var req dto.UpdateByNameAndFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := clamService.UpdateFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Clam
// @Summary Delete clam
// @Accept json
// @Param request body dto.ClamDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"clams","output_column":"name","output_value":"names"}],"formatZH":"删除扫描规则 [names]","formatEN":"delete clam [names]"}
func (b *BaseApi) DeleteClam(c *gin.Context) {
	var req dto.ClamDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := clamService.Delete(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Clam
// @Summary Handle clam scan
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/clam/handle [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":true,"db":"clams","output_column":"name","output_value":"name"}],"formatZH":"执行病毒扫描 [name]","formatEN":"handle clam scan [name]"}
func (b *BaseApi) HandleClamScan(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := clamService.HandleOnce(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
