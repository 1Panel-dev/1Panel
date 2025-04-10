package v2

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
	"github.com/1Panel-dev/1Panel/core/utils/terminal"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// @Tags ScriptLibrary
// @Summary Add script
// @Accept json
// @Param request body dto.ScriptOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /script [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加脚本库脚本 [name]","formatEN":"add script [name]"}
func (b *BaseApi) CreateScript(c *gin.Context) {
	var req dto.ScriptOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := scriptService.Create(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags ScriptLibrary
// @Summary Page script
// @Accept json
// @Param request body dto.SearchPageWithGroup true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /script/search [post]
func (b *BaseApi) SearchScript(c *gin.Context) {
	var req dto.SearchPageWithGroup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := scriptService.Search(c, req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags ScriptLibrary
// @Summary Delete script
// @Accept json
// @Param request body dto.OperateByIDs true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /script/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"script_librarys","output_column":"name","output_value":"names"}],"formatZH":"删除脚本库脚本 [names]","formatEN":"delete script [names]"}
func (b *BaseApi) DeleteScript(c *gin.Context) {
	var req dto.OperateByIDs
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := scriptService.Delete(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags ScriptLibrary
// @Summary Sync script from remote
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /script/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"同步脚本库脚本","formatEN":"sync scripts"}
func (b *BaseApi) SyncScript(c *gin.Context) {
	if err := scriptService.Sync(); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags ScriptLibrary
// @Summary Update script
// @Accept json
// @Param request body dto.ScriptOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /script/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"cronjobs","output_column":"name","output_value":"name"}],"formatZH":"更新脚本库脚本 [name]","formatEN":"update script [name]"}
func (b *BaseApi) UpdateScript(c *gin.Context) {
	var req dto.ScriptOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := scriptService.Update(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

func (b *BaseApi) RunScript(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	if global.CONF.Base.IsDemo {
		if wshandleError(wsConn, errors.New("   demo server, prohibit this operation!")) {
			return
		}
	}

	cols, err := strconv.Atoi(c.DefaultQuery("cols", "80"))
	if wshandleError(wsConn, errors.WithMessage(err, "invalid param cols in request")) {
		return
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "40"))
	if wshandleError(wsConn, errors.WithMessage(err, "invalid param rows in request")) {
		return
	}
	scriptID := c.Query("script_id")
	currentNode := c.Query("current_node")
	intNum, _ := strconv.Atoi(scriptID)
	if intNum == 0 {
		if wshandleError(wsConn, fmt.Errorf("   no such script %v in library, please check and try again!", scriptID)) {
			return
		}
	}
	scriptItem, err := service.LoadScriptInfo(uint(intNum))
	if wshandleError(wsConn, err) {
		return
	}

	fileName := strings.ReplaceAll(scriptItem.Name, " ", "_")
	quitChan := make(chan bool, 3)
	if currentNode == "local" {
		slave, err := terminal.NewCommand(scriptItem.Script)
		if wshandleError(wsConn, err) {
			return
		}
		defer slave.Close()

		tty, err := terminal.NewLocalWsSession(cols, rows, wsConn, slave, true)
		if wshandleError(wsConn, err) {
			return
		}

		quitChan := make(chan bool, 3)
		tty.Start(quitChan)
		go slave.Wait(quitChan)
	} else {
		connInfo, installDir, err := xpack.LoadNodeInfo(currentNode)
		if wshandleError(wsConn, errors.WithMessage(err, "invalid param rows in request")) {
			return
		}
		tmpFile := path.Join(installDir, "1panel/tmp/script")
		initCmd := fmt.Sprintf("d=%s && mkdir -p $d && echo %s > $d/%s && clear && bash $d/%s", tmpFile, scriptItem.Script, fileName, fileName)
		client, err := ssh.NewClient(*connInfo)
		if wshandleError(wsConn, errors.WithMessage(err, "failed to set up the connection. Please check the host information")) {
			return
		}
		defer client.Close()

		sws, err := terminal.NewLogicSshWsSession(cols, rows, client.Client, wsConn, initCmd)
		if wshandleError(wsConn, err) {
			return
		}
		defer sws.Close()
		sws.Start(quitChan)
		go sws.Wait(quitChan)
	}

	<-quitChan

	global.LOG.Info("websocket finished")
	if wshandleError(wsConn, err) {
		return
	}
}
