package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	websocket2 "github.com/1Panel-dev/1Panel/agent/utils/websocket"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) ProcessWs(c *gin.Context) {
	ws, err := wsUpgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	wsClient := websocket2.NewWsClient("processClient", ws)
	go wsClient.Read()
	go wsClient.Write()
}

// @Tags Process
// @Summary Stop Process
// @Param request body request.ProcessReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /process/stop [post]
// @x-panel-log {"bodyKeys":["PID"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"结束进程 [PID]","formatEN":"结束进程 [PID]"}
func (b *BaseApi) StopProcess(c *gin.Context) {
	var req request.ProcessReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := processService.StopProcess(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Process
// @Summary Get Process Info By PID
// @Param pid path int true "PID"
// @Success 200 {object} websocket.PsProcessData
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /process/{pid} [get]
func (b *BaseApi) GetProcessInfoByPID(c *gin.Context) {
	pid, err := helper.GetParamInt32("pid", c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	data, err := processService.GetProcessInfoByPID(pid)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}
