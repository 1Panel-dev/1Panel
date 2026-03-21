package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Mail
// @Summary List mail domains
// @Success 200 {array} response.MailDomainRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/domains [get]
func (b *BaseApi) ListMailDomains(c *gin.Context) {
	domains, err := mailService.ListDomains()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, domains)
}

// @Tags Mail
// @Summary Create mail domain
// @Accept json
// @Param request body request.MailDomainCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/domains [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建邮件域 [name]","formatEN":"Create mail domain [name]"}
func (b *BaseApi) CreateMailDomain(c *gin.Context) {
	var req request.MailDomainCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.CreateDomain(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Delete mail domain
// @Accept json
// @Param request body request.MailDomainDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/domains/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"mail_domains","output_column":"name","output_value":"name"}],"formatZH":"删除邮件域 [name]","formatEN":"Delete mail domain [name]"}
func (b *BaseApi) DeleteMailDomain(c *gin.Context) {
	var req request.MailDomainDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.DeleteDomain(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Search mail accounts
// @Accept json
// @Param request body request.MailAccountSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/accounts/search [post]
func (b *BaseApi) SearchMailAccounts(c *gin.Context) {
	var req request.MailAccountSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, accounts, err := mailService.SearchAccounts(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{Total: total, Items: accounts})
}

// @Tags Mail
// @Summary Create mail account
// @Accept json
// @Param request body request.MailAccountCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/accounts [post]
// @x-panel-log {"bodyKeys":["username"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建邮箱账号 [username]","formatEN":"Create mail account [username]"}
func (b *BaseApi) CreateMailAccount(c *gin.Context) {
	var req request.MailAccountCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.CreateAccount(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Update mail account
// @Accept json
// @Param request body request.MailAccountUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/accounts/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"mail_accounts","output_column":"email","output_value":"email"}],"formatZH":"更新邮箱账号 [email]","formatEN":"Update mail account [email]"}
func (b *BaseApi) UpdateMailAccount(c *gin.Context) {
	var req request.MailAccountUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.UpdateAccount(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Delete mail account
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/accounts/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"mail_accounts","output_column":"email","output_value":"email"}],"formatZH":"删除邮箱账号 [email]","formatEN":"Delete mail account [email]"}
func (b *BaseApi) DeleteMailAccount(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.DeleteAccount(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Search mail aliases
// @Accept json
// @Param request body request.MailAliasSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/aliases/search [post]
func (b *BaseApi) SearchMailAliases(c *gin.Context) {
	var req request.MailAliasSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, aliases, err := mailService.SearchAliases(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{Total: total, Items: aliases})
}

// @Tags Mail
// @Summary Create mail alias
// @Accept json
// @Param request body request.MailAliasCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/aliases [post]
// @x-panel-log {"bodyKeys":["source","target"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建邮箱别名 [source] → [target]","formatEN":"Create mail alias [source] → [target]"}
func (b *BaseApi) CreateMailAlias(c *gin.Context) {
	var req request.MailAliasCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.CreateAlias(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Delete mail alias
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/aliases/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"mail_aliases","output_column":"source","output_value":"source"}],"formatZH":"删除邮箱别名 [source]","formatEN":"Delete mail alias [source]"}
func (b *BaseApi) DeleteMailAlias(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.DeleteAlias(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Setup mail (enable/disable)
// @Accept json
// @Param request body request.MailSetup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/setup [post]
// @x-panel-log {"bodyKeys":["enable"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"邮件服务设置 [enable]","formatEN":"Mail setup [enable]"}
func (b *BaseApi) SetupMail(c *gin.Context) {
	var req request.MailSetup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := mailService.Setup(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Mail
// @Summary Get mail status
// @Success 200 {object} response.MailStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mail/status [get]
func (b *BaseApi) GetMailStatus(c *gin.Context) {
	status, err := mailService.GetStatus()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, status)
}
