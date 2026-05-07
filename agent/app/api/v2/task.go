package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/gin-gonic/gin"
)

// @Tags TaskLog
// @Summary Page task logs
// @Accept json
// @Param request body dto.SearchTaskLogReq true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/tasks/search [post]
func (b *BaseApi) PageTasks(c *gin.Context) {
	var req dto.SearchTaskLogReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := taskService.Page(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags TaskLog
// @Summary Read task log by Line
// @Param request body request.TaskLogReadReq true "request"
// @Success 200 {object} response.FileLineContent
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/tasks/read [post]
func (b *BaseApi) ReadTaskLogByLine(c *gin.Context) {
	var req request.TaskLogReadReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	var res *response.FileLineContent
	res, err := taskService.ReadByLine(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	if res.TotalLines > 100 {
		helper.SuccessWithDataGzipped(c, res)
	} else {
		helper.SuccessWithData(c, res)
	}
}

// @Tags TaskLog
// @Summary Get the number of executing tasks
// @Success 200 {integer} int64
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/tasks/executing/count [get]
func (b *BaseApi) CountExecutingTasks(c *gin.Context) {
	count, err := taskService.CountExecutingTask()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, count)
}
