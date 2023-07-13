package v1

import (
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/gin-gonic/gin"
)

// @Tags Host
// @Summary Create host
// @Description 创建主机
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts [post]
// @x-panel-log {"bodyKeys":["name","addr"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建主机 [name][addr]","formatEN":"create host [name][addr]"}
func (b *BaseApi) CreateHost(c *gin.Context) {
	var req dto.HostOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if req.AuthMode == "password" && len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		passwordItem, err := encrypt.StringEncrypt(string(password))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Password = passwordItem
		req.PrivateKey = ""
		req.PassPhrase = ""
	}
	if req.AuthMode == "key" && len(req.PrivateKey) != 0 {
		privateKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		keyItem, err := encrypt.StringEncrypt(string(privateKey))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Password = keyItem

		if len(req.PassPhrase) != 0 {
			pass, err := encrypt.StringEncrypt(req.PassPhrase)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
				return
			}
			req.PassPhrase = pass
		}
		req.Password = ""
	}

	host, err := hostService.Create(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, host)
}

// @Tags Host
// @Summary Test host conn by info
// @Description 测试主机连接
// @Accept json
// @Param request body dto.HostConnTest true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts/test/byinfo [post]
func (b *BaseApi) TestByInfo(c *gin.Context) {
	var req dto.HostConnTest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	connStatus := hostService.TestByInfo(req)
	helper.SuccessWithData(c, connStatus)
}

// @Tags Host
// @Summary Test host conn by host id
// @Description 测试主机连接
// @Accept json
// @Param id path integer true "request"
// @Success 200 {boolean} connStatus
// @Security ApiKeyAuth
// @Router /hosts/test/byid/:id [post]
func (b *BaseApi) TestByID(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	connStatus := hostService.TestLocalConn(id)
	helper.SuccessWithData(c, connStatus)
}

// @Tags Host
// @Summary Load host tree
// @Description 加载主机树
// @Accept json
// @Param request body dto.SearchForTree true "request"
// @Success 200 {array} dto.HostTree
// @Security ApiKeyAuth
// @Router /hosts/tree [post]
func (b *BaseApi) HostTree(c *gin.Context) {
	var req dto.SearchForTree
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	data, err := hostService.SearchForTree(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Host
// @Summary Page host
// @Description 获取主机列表分页
// @Accept json
// @Param request body dto.SearchHostWithPage true "request"
// @Success 200 {array} dto.HostTree
// @Security ApiKeyAuth
// @Router /hosts/search [post]
func (b *BaseApi) SearchHost(c *gin.Context) {
	var req dto.SearchHostWithPage
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := hostService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Host
// @Summary Delete host
// @Description 删除主机
// @Accept json
// @Param request body dto.BatchDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"ids","isList":true,"db":"hosts","output_column":"addr","output_value":"addrs"}],"formatZH":"删除主机 [addrs]","formatEN":"delete host [addrs]"}
func (b *BaseApi) DeleteHost(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := hostService.Delete(req.Ids); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Host
// @Summary Update host
// @Description 更新主机
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts/update [post]
// @x-panel-log {"bodyKeys":["name","addr"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新主机信息 [name][addr]","formatEN":"update host [name][addr]"}
func (b *BaseApi) UpdateHost(c *gin.Context) {
	var req dto.HostOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if req.AuthMode == "password" && len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		passwordItem, err := encrypt.StringEncrypt(string(password))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Password = passwordItem
	}
	if req.AuthMode == "key" && len(req.PrivateKey) != 0 {
		privateKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		keyItem, err := encrypt.StringEncrypt(string(privateKey))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.PrivateKey = keyItem

		if len(req.PassPhrase) != 0 {
			pass, err := encrypt.StringEncrypt(req.PassPhrase)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
				return
			}
			req.PassPhrase = pass
		}
	}

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["group_id"] = req.GroupID
	upMap["addr"] = req.Addr
	upMap["port"] = req.Port
	upMap["user"] = req.User
	upMap["auth_mode"] = req.AuthMode
	upMap["remember_password"] = req.RememberPassword
	if req.AuthMode == "password" {
		upMap["password"] = req.Password
		upMap["private_key"] = ""
		upMap["pass_phrase"] = ""
	} else {
		upMap["password"] = ""
		upMap["private_key"] = req.PrivateKey
		upMap["pass_phrase"] = req.PassPhrase
	}
	upMap["description"] = req.Description
	if err := hostService.Update(req.ID, upMap); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Host
// @Summary Update host group
// @Description 切换分组
// @Accept json
// @Param request body dto.ChangeHostGroup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /hosts/update/group [post]
// @x-panel-log {"bodyKeys":["id","group"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"hosts","output_column":"addr","output_value":"addr"}],"formatZH":"切换主机[addr]分组 => [group]","formatEN":"change host [addr] group => [group]"}
func (b *BaseApi) UpdateHostGroup(c *gin.Context) {
	var req dto.ChangeHostGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	upMap := make(map[string]interface{})
	upMap["group_id"] = req.GroupID
	if err := hostService.Update(req.ID, upMap); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
