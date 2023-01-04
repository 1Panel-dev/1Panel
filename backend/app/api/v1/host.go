package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/copier"
	"github.com/1Panel-dev/1Panel/backend/utils/ssh"
	"github.com/gin-gonic/gin"
)

// Create host
// @Tags Host
// @Summary Create host
// @Description 创建主机
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host [post]
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
	host, err := hostService.Create(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, host)
}

// Test host conn by info
// @Tags Host
// @Summary Test host conn by info
// @Description 测试主机连接
// @Accept json
// @Param request body dto.HostConnTest true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/test/byinfo [post]
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

	var connInfo ssh.ConnInfo
	if err := copier.Copy(&connInfo, &req); err != nil {
		helper.SuccessWithData(c, false)
	}
	client, err := connInfo.NewClient()
	if err != nil {
		helper.SuccessWithData(c, false)
	}
	defer client.Close()

	helper.SuccessWithData(c, true)
}

// Test host conn by host id
// @Tags Host
// @Summary Test host conn by host id
// @Description 测试主机连接
// @Accept json
// @Param id path integer true "request"
// @Success 200 {boolean}
// @Security ApiKeyAuth
// @Router /host/test/byid/:id [post]
func (b *BaseApi) TestByID(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	connStatus := hostService.TestLocalConn(id)
	helper.SuccessWithData(c, connStatus)
}

// Load host tree
// @Tags Host
// @Summary Load host tree
// @Description 加载主机树
// @Accept json
// @Param request body dto.SearchForTree true "request"
// @Success 200 {anrry} dto.HostTree
// @Security ApiKeyAuth
// @Router /host/search [post]
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

// Load host info
// @Tags Host
// @Summary Load host info
// @Description 加载主机信息
// @Accept json
// @Param id path integer true "request"
// @Success 200 {object} dto.HostInfo
// @Security ApiKeyAuth
// @Router /host/:id [get]
func (b *BaseApi) GetHostInfo(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	host, err := hostService.GetHostInfo(id)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	var hostDto dto.HostInfo
	if err := copier.Copy(&hostDto, host); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, hostDto)
}

// Delete host
// @Tags Host
// @Summary Delete host
// @Description 删除主机
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_colume":"id","input_value":"id","isList":false,"db":"hosts","output_colume":"addr","output_value":"addr"}],"formatZH":"删除主机 [addr]","formatEN":"delete host [addr]"}
func (b *BaseApi) DeleteHost(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := hostService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// Update host
// @Tags Host
// @Summary Update host
// @Description 更新主机
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/update [post]
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

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["group_belong"] = req.GroupBelong
	upMap["addr"] = req.Addr
	upMap["port"] = req.Port
	upMap["user"] = req.User
	upMap["auth_mode"] = req.AuthMode
	upMap["password"] = req.Password
	upMap["private_key"] = req.PrivateKey
	upMap["description"] = req.Description
	if err := hostService.Update(req.ID, upMap); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
