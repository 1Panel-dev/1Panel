package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags TaskLog
// @Summary Page task logs
// @Description 获取任务日志列表
// @Accept json
// @Param request body dto.SearchTaskLogReq true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
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
