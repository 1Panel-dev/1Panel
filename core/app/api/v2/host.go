package v2

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/copier"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
	"github.com/1Panel-dev/1Panel/core/utils/terminal"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// @Tags Host
// @Summary Create host
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200 {object} dto.HostInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts [post]
// @x-panel-log {"bodyKeys":["name","addr"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建主机 [name][addr]","formatEN":"create host [name][addr]"}
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
// @Summary Test host conn by info
// @Accept json
// @Param request body dto.HostConnTest true "request"
// @Success 200 {boolean} status
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts/test/byinfo [post]
func (b *BaseApi) TestByInfo(c *gin.Context) {
	var req dto.HostConnTest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	connStatus := hostService.TestByInfo(req)
	helper.SuccessWithData(c, connStatus)
}

// @Tags Host
// @Summary Test host conn by host id
// @Accept json
// @Param id path integer true "request"
// @Success 200 {boolean} connStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts/test/byid/:id [post]
func (b *BaseApi) TestByID(c *gin.Context) {
	idParam, ok := c.Params.Get("id")
	if !ok {
		helper.BadRequest(c, errors.New("no such params find in request"))
		return
	}
	intNum, err := strconv.Atoi(idParam)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	connStatus := hostService.TestLocalConn(uint(intNum))
	helper.SuccessWithData(c, connStatus)
}

// @Tags Host
// @Summary Load host tree
// @Accept json
// @Param request body dto.SearchForTree true "request"
// @Success 200 {array} dto.HostTree
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts/tree [post]
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
// @Summary Page host
// @Accept json
// @Param request body dto.SearchHostWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts/search [post]
func (b *BaseApi) SearchHost(c *gin.Context) {
	var req dto.SearchHostWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := hostService.SearchWithPage(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Host
// @Summary Delete host
// @Accept json
// @Param request body dto.OperateByIDs true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"hosts","output_column":"addr","output_value":"addrs"}],"formatZH":"删除主机 [addrs]","formatEN":"delete host [addrs]"}
func (b *BaseApi) DeleteHost(c *gin.Context) {
	var req dto.OperateByIDs
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.Delete(req.IDs); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Host
// @Summary Update host
// @Accept json
// @Param request body dto.HostOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts/update [post]
// @x-panel-log {"bodyKeys":["name","addr"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新主机信息 [name][addr]","formatEN":"update host [name][addr]"}
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
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags Host
// @Summary Update host group
// @Accept json
// @Param request body dto.ChangeHostGroup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/hosts/update/group [post]
// @x-panel-log {"bodyKeys":["id","group"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"hosts","output_column":"addr","output_value":"addr"}],"formatZH":"切换主机[addr]分组 => [group]","formatEN":"change host [addr] group => [group]"}
func (b *BaseApi) UpdateHostGroup(c *gin.Context) {
	var req dto.ChangeHostGroup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["group_id"] = req.GroupID
	if err := hostService.Update(req.ID, upMap); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

func (b *BaseApi) WsSsh(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	id, err := strconv.Atoi(c.Query("id"))
	if wshandleError(wsConn, errors.WithMessage(err, "invalid param id in request")) {
		return
	}
	cols, err := strconv.Atoi(c.DefaultQuery("cols", "80"))
	if wshandleError(wsConn, errors.WithMessage(err, "invalid param cols in request")) {
		return
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "40"))
	if wshandleError(wsConn, errors.WithMessage(err, "invalid param rows in request")) {
		return
	}
	host, err := hostService.GetHostInfo(uint(id))
	if wshandleError(wsConn, errors.WithMessage(err, "load host info by id failed")) {
		return
	}
	var connInfo ssh.ConnInfo
	_ = copier.Copy(&connInfo, &host)
	connInfo.PrivateKey = []byte(host.PrivateKey)
	if len(host.PassPhrase) != 0 {
		connInfo.PassPhrase = []byte(host.PassPhrase)
	}

	client, err := ssh.NewClient(connInfo)
	if wshandleError(wsConn, errors.WithMessage(err, "failed to set up the connection. Please check the host information")) {
		return
	}
	defer client.Close()
	sws, err := terminal.NewLogicSshWsSession(cols, rows, true, client.Client, wsConn)
	if wshandleError(wsConn, err) {
		return
	}
	defer sws.Close()

	quitChan := make(chan bool, 3)
	sws.Start(quitChan)
	go sws.Wait(quitChan)

	<-quitChan

	if wshandleError(wsConn, err) {
		return
	}
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wshandleError(ws *websocket.Conn, err error) bool {
	if err != nil {
		global.LOG.Errorf("handler ws faled:, err: %v", err)
		dt := time.Now().Add(time.Second)
		if ctlerr := ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt); ctlerr != nil {
			wsData, err := json.Marshal(terminal.WsMsg{
				Type: terminal.WsMsgCmd,
				Data: base64.StdEncoding.EncodeToString([]byte(err.Error())),
			})
			if err != nil {
				_ = ws.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"cmd\",\"data\":\"failed to encoding to json\"}"))
			} else {
				_ = ws.WriteMessage(websocket.TextMessage, wsData)
			}
		}
		return true
	}
	return false
}
