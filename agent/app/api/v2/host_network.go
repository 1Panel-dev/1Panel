package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/gin-gonic/gin"
)

var hostNetworkService = &service.HostNetworkService{}

// @Tags Host Network
// @Summary Add interface address
// @Accept json
// @Param request body dto.HostNetworkAddressChange true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/network/address [post]
// @x-panel-log {"bodyKeys":["interface","address"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加网卡 [interface] 地址 [address]","formatEN":"add address [address] to interface [interface]"}
func (b *BaseApi) AddHostNetworkAddress(c *gin.Context) {
	var req dto.HostNetworkAddressChange
	if helper.CheckBindAndValidate(&req, c) != nil {
		return
	}
	if err := hostNetworkService.UpdateAddress(req, false); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host Network
// @Summary Remove interface address
// @Accept json
// @Param request body dto.HostNetworkAddressChange true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/network/address/del [post]
// @x-panel-log {"bodyKeys":["interface","address"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"移除网卡 [interface] 地址 [address]","formatEN":"remove address [address] from interface [interface]"}
func (b *BaseApi) DeleteHostNetworkAddress(c *gin.Context) {
	var req dto.HostNetworkAddressChange
	if helper.CheckBindAndValidate(&req, c) != nil {
		return
	}
	if err := hostNetworkService.UpdateAddress(req, true); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host Network
// @Summary Add route
// @Accept json
// @Param request body dto.HostNetworkRouteChange true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/network/routes [post]
// @x-panel-log {"bodyKeys":["destination","gateway","interface"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加路由 [destination] 网关 [gateway] 设备 [interface]","formatEN":"add route [destination] gateway [gateway] device [interface]"}
func (b *BaseApi) AddHostNetworkRoute(c *gin.Context) {
	var req dto.HostNetworkRouteChange
	if helper.CheckBindAndValidate(&req, c) != nil {
		return
	}
	if err := hostNetworkService.UpdateRoute(req, false); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host Network
// @Summary Remove route
// @Accept json
// @Param request body dto.HostNetworkRouteChange true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/network/routes/del [post]
// @x-panel-log {"bodyKeys":["destination","gateway","interface"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"移除路由 [destination] 网关 [gateway] 设备 [interface]","formatEN":"remove route [destination] gateway [gateway] device [interface]"}
func (b *BaseApi) DeleteHostNetworkRoute(c *gin.Context) {
	var req dto.HostNetworkRouteChange
	if helper.CheckBindAndValidate(&req, c) != nil {
		return
	}
	if err := hostNetworkService.UpdateRoute(req, true); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Host Network
// @Summary Check route reachability
// @Accept json
// @Param request body dto.HostNetworkRoute true "request"
// @Success 200 {object} dto.HostNetworkRouteCheck
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/network/routes/check [post]
func (b *BaseApi) CheckHostNetworkRoute(c *gin.Context) {
	var req dto.HostNetworkRoute
	if helper.CheckBindAndValidate(&req, c) != nil {
		return
	}
	data, err := hostNetworkService.CheckRoute(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Host Network
// @Summary Check reachability for a batch of routes
// @Accept json
// @Param request body dto.HostNetworkRouteBatchCheck true "request"
// @Success 200 {array} dto.HostNetworkRouteCheck
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/network/routes/check/batch [post]
func (b *BaseApi) CheckHostNetworkRoutes(c *gin.Context) {
	var req dto.HostNetworkRouteBatchCheck
	if helper.CheckBindAndValidate(&req, c) != nil {
		return
	}
	data, err := hostNetworkService.CheckRoutes(c.Request.Context(), req.Routes)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Host Network
// @Summary Get network overview
// @Success 200 {object} dto.HostNetworkOverview
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/network [get]
func (b *BaseApi) GetHostNetworkOverview(c *gin.Context) {
	data, err := hostNetworkService.Overview()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}
