package v1

import (
	"github.com/1Panel-dev/1Panel/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) GetOperationList(c *gin.Context) {
	pagenation, isOK := helper.GeneratePaginationFromReq(c)
	if !isOK {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, constant.ErrPageGenerate)
		return
	}

	total, list, err := operationService.Page(pagenation.Page, pagenation.PageSize)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}
