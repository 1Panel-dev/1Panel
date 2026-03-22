package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Toolbox
// @Summary Run network diagnostic tool
// @Accept json
// @Param request body dto.NetToolReq true "request"
// @Success 200 {object} dto.NetToolRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/net [post]
func (b *BaseApi) RunNetTool(c *gin.Context) {
	var req dto.NetToolReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	var output string
	var err error

	switch req.Tool {
	case "ping":
		output, err = toolboxNetService.Ping(req)
	case "traceroute":
		output, err = toolboxNetService.Traceroute(req)
	case "dns":
		output, err = toolboxNetService.DNSLookup(req)
	case "http":
		output, err = toolboxNetService.HTTPCheck(req)
	}

	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.NetToolRes{Output: output})
}
