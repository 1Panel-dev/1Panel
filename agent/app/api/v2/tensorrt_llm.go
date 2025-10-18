package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) PageTensorRTLLMs(c *gin.Context) {
	var req request.TensorRTLLMSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	list := tensorrtLLMService.Page(req)
	helper.SuccessWithData(c, list)
}

func (b *BaseApi) CreateTensorRTLLM(c *gin.Context) {
	var req request.TensorRTLLMCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := tensorrtLLMService.Create(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) UpdateTensorRTLLM(c *gin.Context) {
	var req request.TensorRTLLMUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := tensorrtLLMService.Update(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) DeleteTensorRTLLM(c *gin.Context) {
	var req request.TensorRTLLMDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := tensorrtLLMService.Delete(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) OperateTensorRTLLM(c *gin.Context) {
	var req request.TensorRTLLMOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := tensorrtLLMService.Operate(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
