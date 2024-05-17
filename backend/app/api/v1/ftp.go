package v1

import (
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags FTP
// @Summary Page FTP user
// @Description 获取 FTP 账户列表分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /toolbox/ftp/search [post]
func (b *BaseApi) SearchFtp(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := ftpService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags FTP
// @Summary Create FTP user
// @Description 创建 FTP 账户
// @Accept json
// @Param request body dto.FtpCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/ftp [post]
// @x-panel-log {"bodyKeys":["user", "path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 FTP 账户 [user][path]","formatEN":"create FTP [user][path]"}
func (b *BaseApi) CreateFtp(c *gin.Context) {
	var req dto.FtpCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.Password) != 0 {
		pass, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Password = string(pass)
	}
	if err := ftpService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags FTP
// @Summary Delete FTP user
// @Description 删除 FTP 账户
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/ftp/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"ftps","output_column":"user","output_value":"users"}],"formatZH":"删除 FTP 账户 [users]","formatEN":"delete FTP users [users]"}
func (b *BaseApi) DeleteFtp(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := ftpService.Delete(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags FTP
// @Summary Sync FTP user
// @Description 同步 FTP 账户
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/ftp/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"同步 FTP 账户","formatEN":"sync FTP users"}
func (b *BaseApi) SyncFtp(c *gin.Context) {
	if err := ftpService.Sync(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags FTP
// @Summary Update FTP user
// @Description 修改 FTP 账户
// @Accept json
// @Param request body dto.FtpUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/ftp/update [post]
// @x-panel-log {"bodyKeys":["user", "path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改 FTP 账户 [user][path]","formatEN":"update FTP [user][path]"}
func (b *BaseApi) UpdateFtp(c *gin.Context) {
	var req dto.FtpUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.Password) != 0 {
		pass, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Password = string(pass)
	}
	if err := ftpService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
