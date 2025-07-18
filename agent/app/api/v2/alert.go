package v2

import (
	"errors"
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) PageAlert(c *gin.Context) {
	var req dto.AlertSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, alerts, err := alertService.PageAlert(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: alerts,
	})
}

func (b *BaseApi) GetAlerts(c *gin.Context) {
	alerts, err := alertService.GetAlerts()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, alerts)
}

func (b *BaseApi) CreateAlert(c *gin.Context) {
	var req dto.AlertCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := alertService.CreateAlert(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) DeleteAlert(c *gin.Context) {
	var req dto.DeleteRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := alertService.DeleteAlert(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) UpdateAlert(c *gin.Context) {
	var req dto.AlertUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := alertService.UpdateAlert(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) GetAlert(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, errors.New("no such id in request param"))
		return
	}
	alert, err := alertService.GetAlert(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, alert)
}

func (b *BaseApi) UpdateAlertStatus(c *gin.Context) {
	var req dto.AlertUpdateStatus
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := alertService.UpdateStatus(req.ID, req.Status); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) GetDisks(c *gin.Context) {
	alerts, err := alertService.GetDisks()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, alerts)
}

func (b *BaseApi) PageAlertLogs(c *gin.Context) {
	var req dto.AlertLogSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, alertLogs, err := alertService.PageAlertLogs(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: alertLogs,
	})
}

func (b *BaseApi) CleanAlertLogs(c *gin.Context) {
	if err := alertService.CleanAlertLogs(); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) GetClams(c *gin.Context) {
	clams, err := alertService.GetClams()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, clams)
}

func (b *BaseApi) GetCronJobs(c *gin.Context) {
	var req dto.CronJobReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	cronJobs, err := alertService.GetCronJobs(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, cronJobs)
}

func (b *BaseApi) GetAlertConfig(c *gin.Context) {
	config, err := alertService.GetAlertConfig()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, config)
}

func (b *BaseApi) UpdateAlertConfig(c *gin.Context) {
	var req dto.AlertConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := alertService.UpdateAlertConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) DeleteAlertConfig(c *gin.Context) {
	var req dto.DeleteRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := alertService.DeleteAlertConfig(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) TestAlertConfig(c *gin.Context) {
	var req dto.AlertConfigTest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	flag, err := alertService.TestAlertConfig(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, flag)
}
