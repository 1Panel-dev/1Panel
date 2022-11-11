package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) PageWebsiteSSL(c *gin.Context) {
	var req dto.PageInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	total, accounts, err := websiteSSLService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: accounts,
	})
}

//func (b *BaseApi) CreateWebsiteSSL(c *gin.Context) {
//	var req dto.WebsiteSSLCreate
//	if err := c.ShouldBindJSON(&req); err != nil {
//		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
//		return
//	}
//	if _, err := WebsiteSSLService.Create(req); err != nil {
//		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
//		return
//	}
//	helper.SuccessWithData(c, nil)
//}
//
//func (b *BaseApi) UpdateWebsiteSSL(c *gin.Context) {
//	var req dto.WebsiteSSLUpdate
//	if err := c.ShouldBindJSON(&req); err != nil {
//		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
//		return
//	}
//	if _, err := WebsiteSSLService.Update(req); err != nil {
//		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
//		return
//	}
//	helper.SuccessWithData(c, nil)
//}
//
//func (b *BaseApi) DeleteWebsiteSSL(c *gin.Context) {
//
//	id, err := helper.GetParamID(c)
//	if err != nil {
//		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
//		return
//	}
//
//	if err := WebsiteSSLService.Delete(id); err != nil {
//		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
//		return
//	}
//	helper.SuccessWithData(c, nil)
//}
