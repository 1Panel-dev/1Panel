package v2

import (
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/gin-gonic/gin"
)

// @Tags FTP
// @Summary Load FTP base info
// @Success 200 {object} dto.FtpBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/ftp/base [get]
func (b *BaseApi) LoadFtpBaseInfo(c *gin.Context) {
	data, err := ftpService.LoadBaseInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags FTP
// @Summary Load FTP operation log
// @Accept json
// @Param request body dto.FtpLogSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/ftp/log/search [post]
func (b *BaseApi) LoadFtpLogInfo(c *gin.Context) {
	var req dto.FtpLogSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := ftpService.LoadLog(req)
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
// @Summary Operate FTP
// @Accept json
// @Param request body dto.Operate true "request"
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/ftp/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] FTP","formatEN":"[operation] FTP"}
func (b *BaseApi) OperateFtp(c *gin.Context) {
	var req dto.Operate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := ftpService.Operate(req.Operation); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithOutData(c)
}

// @Tags FTP
// @Summary Page FTP user
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.FtpCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
			helper.BadRequest(c, err)
			return
		}
		req.Password = string(pass)
	}
	if _, err := ftpService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags FTP
// @Summary Delete FTP user
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.FtpUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
			helper.BadRequest(c, err)
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
