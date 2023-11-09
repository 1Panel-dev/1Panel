package v1

import (
	"bufio"
	"os"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Fail2ban
// @Summary Load fail2ban base info
// @Description 获取 Fail2ban 基础信息
// @Success 200 {object} dto.Fail2banBaseInfo
// @Security ApiKeyAuth
// @Router /toolbox/fail2ban/base [get]
func (b *BaseApi) LoadFail2banBaseInfo(c *gin.Context) {
	data, err := fail2banService.LoadBaseInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Fail2ban
// @Summary Page fail2ban ip list
// @Description 获取 Fail2ban ip 列表
// @Accept json
// @Param request body dto.Fail2banSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /toolbox/fail2ban/search [post]
func (b *BaseApi) SearchFail2ban(c *gin.Context) {
	var req dto.Fail2banSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := fail2banService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Fail2ban
// @Summary Operate fail2ban
// @Description 修改 Fail2ban 状态
// @Accept json
// @Param request body dto.Operate true "request"
// @Security ApiKeyAuth
// @Router /toolbox/fail2ban/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] Fail2ban","formatEN":"[operation] Fail2ban"}
func (b *BaseApi) OperateFail2ban(c *gin.Context) {
	var req dto.Operate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := fail2banService.Operate(req.Operation); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Fail2ban
// @Summary Operate sshd of fail2ban
// @Description 配置 sshd
// @Accept json
// @Param request body dto.Operate true "request"
// @Security ApiKeyAuth
// @Router /toolbox/fail2ban/operate/sshd [post]
// @x-panel-log {"bodyKeys":["operate", "ips"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] ips: [ips]","formatEN":"[operate] ips: [ips]"}
func (b *BaseApi) OperateSSHD(c *gin.Context) {
	var req dto.Fail2banSet
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := fail2banService.OperateSSHD(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Fail2ban
// @Summary Update fail2ban conf
// @Description 修改 Fail2ban 配置
// @Accept json
// @Param request body dto.Fail2banUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/fail2ban/update [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改 Fail2ban 配置 [key] => [value]","formatEN":"update fail2ban conf [key] => [value]"}
func (b *BaseApi) UpdateFail2banConf(c *gin.Context) {
	var req dto.Fail2banUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := fail2banService.UpdateConf(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Fail2ban
// @Summary Update fail2ban conf by file
// @Description 通过文件修改 fail2ban 配置
// @Accept json
// @Param request body dto.UpdateByFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/fail2ban/update/byconf [post]
func (b *BaseApi) UpdateFail2banConfByFile(c *gin.Context) {
	var req dto.UpdateByFile
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}
	path := "/etc/fail2ban/jail.local"
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	helper.SuccessWithData(c, nil)
}
