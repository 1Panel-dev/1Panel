package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/gin-gonic/gin"
)

// @Tags Host
// @Summary Create host
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts [post]
func (b *BaseApi) CreateHost(c *gin.Context) {
	var req dto.HostOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	host, err := hostService.Create(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, host)
}

// @Tags Host
// @Summary Test by info
// @Accept json
// @Param request body dto.HostConnTest true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/test/byinfo [post]
func (b *BaseApi) TestByInfo(c *gin.Context) {
	var req dto.HostConnTest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	helper.SuccessWithData(c, hostService.TestByInfo(req))
}

// @Tags Host
// @Summary Test by ID
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/test/byid [post]
func (b *BaseApi) TestByID(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	helper.SuccessWithData(c, hostService.TestLocalConn(req.ID))
}

// @Tags Host
// @Summary Host tree
// @Accept json
// @Param request body dto.SearchForTree true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/tree [post]
func (b *BaseApi) HostTree(c *gin.Context) {
	var req dto.SearchForTree
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := hostService.SearchForTree(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Host
// @Summary Search host
// @Accept json
// @Param request body dto.SearchPageWithGroup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/search [post]
func (b *BaseApi) SearchHost(c *gin.Context) {
	var req dto.SearchPageWithGroup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := hostService.SearchWithPage(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{Items: list, Total: total})
}

// @Tags Host
// @Summary Delete host
// @Accept json
// @Param request body dto.OperateByIDs true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/del [post]
func (b *BaseApi) DeleteHost(c *gin.Context) {
	var req dto.OperateByIDs
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.Delete(req.IDs); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host
// @Summary Update host
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/update [post]
func (b *BaseApi) UpdateHost(c *gin.Context) {
	var req dto.HostOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	var err error
	if len(req.Password) != 0 && req.AuthMode == "password" {
		req.Password, err = hostService.EncryptHost(req.Password)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		req.PrivateKey = ""
		req.PassPhrase = ""
	}
	if len(req.PrivateKey) != 0 && req.AuthMode == "key" {
		req.PrivateKey, err = hostService.EncryptHost(req.PrivateKey)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		if len(req.PassPhrase) != 0 {
			req.PassPhrase, err = encrypt.StringEncrypt(req.PassPhrase)
			if err != nil {
				helper.BadRequest(c, err)
				return
			}
		}
		req.Password = ""
	}

	upMap := map[string]interface{}{
		"name":              req.Name,
		"group_id":          req.GroupID,
		"addr":              req.Addr,
		"port":              req.Port,
		"user":              req.User,
		"auth_mode":         req.AuthMode,
		"remember_password": req.RememberPassword,
		"description":       req.Description,
	}
	if req.AuthMode == "password" {
		upMap["password"] = req.Password
		upMap["private_key"] = ""
		upMap["pass_phrase"] = ""
	} else {
		upMap["password"] = ""
		upMap["private_key"] = req.PrivateKey
		upMap["pass_phrase"] = req.PassPhrase
	}
	hostItem, err := hostService.Update(req.ID, upMap)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, hostItem)
}

// @Tags Host
// @Summary Update host group
// @Accept json
// @Param request body dto.ChangeGroup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/update/group [post]
func (b *BaseApi) UpdateHostGroup(c *gin.Context) {
	var req dto.ChangeGroup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if _, err := hostService.Update(req.ID, map[string]interface{}{"group_id": req.GroupID}); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host
// @Summary Get host by ID
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/info [post]
func (b *BaseApi) GetHostByID(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	info, err := hostService.GetHostByID(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, info)
}
