package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) GetContainerImageCheckUpdateStatus(c *gin.Context) {
	status, err := service.GetContainerImageCheckUpdateStatus()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, status)
}
func (b *BaseApi) ContainerImageCheckUpdate(c *gin.Context) {
	taskID, err := service.StartContainerImageCheckUpdate()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	status, err := service.GetContainerImageCheckUpdateStatus()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	status.TaskID = taskID
	helper.SuccessWithData(c, status)
}
func (b *BaseApi) SetContainerImageCheckUpdateSchedule(c *gin.Context) {
	var req struct {
		Enabled *bool `json:"enabled" validate:"required"`
	}
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := service.SetContainerImageCheckUpdateSchedule(*req.Enabled); err != nil {
		helper.InternalServer(c, err)
		return
	}
	b.GetContainerImageCheckUpdateStatus(c)
}
