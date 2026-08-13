package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) ListDockerPortGuard(c *gin.Context) {
	data, err := dockerPortGuardService.List(c.Request.Context())
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

func (b *BaseApi) SyncDockerPortGuard(c *gin.Context) {
	if err := dockerPortGuardService.Reconcile(c.Request.Context()); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) OperateDockerPortGuard(c *gin.Context) {
	var request dto.DockerPortGuardOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := dockerPortGuardService.Operate(c.Request.Context(), request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) DeleteDockerPortGuardPolicies(c *gin.Context) {
	var request dto.DockerPortGuardPolicyBatchDelete
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := dockerPortGuardService.DeleteBatch(c.Request.Context(), request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) UpsertDockerPortGuardPolicies(c *gin.Context) {
	var request dto.DockerPortGuardPolicyBatch
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := dockerPortGuardService.UpsertBatch(c.Request.Context(), request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
