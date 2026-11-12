package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (b *BaseApi) ChatAgentAccount(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2*1024*1024)
	var req dto.ModelChat
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	response, err := agentService.ChatAccount(c.Request.Context(), req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.StreamResponse(c, response)
}
