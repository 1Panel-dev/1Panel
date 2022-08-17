package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/ssh"
	"github.com/1Panel-dev/1Panel/utils/terminal"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (b *BaseApi) WsSsh(c *gin.Context) {
	host := ssh.ConnInfo{
		Addr:     "172.16.10.111",
		Port:     22,
		User:     "root",
		AuthMode: "password",
		Password: "Calong@2015",
	}

	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	cols, err := strconv.Atoi(c.DefaultQuery("cols", "80"))
	if wshandleError(wsConn, err) {
		return
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "40"))
	if wshandleError(wsConn, err) {
		return
	}

	client, err := host.NewClient()
	if wshandleError(wsConn, err) {
		return
	}
	defer client.Close()
	ssConn, err := host.NewSshConn(cols, rows)
	if wshandleError(wsConn, err) {
		return
	}
	defer ssConn.Close()

	sws, err := terminal.NewLogicSshWsSession(cols, rows, true, host.Client, wsConn)
	if wshandleError(wsConn, err) {
		return
	}
	defer sws.Close()

	quitChan := make(chan bool, 3)
	sws.Start(quitChan)
	go sws.Wait(quitChan)

	<-quitChan

	global.LOG.Info("websocket finished")
	if wshandleError(wsConn, err) {
		return
	}
}

func wshandleError(ws *websocket.Conn, err error) bool {
	if err != nil {
		global.LOG.Errorf("handler ws faled:, err: %v", err)
		dt := time.Now().Add(time.Second)
		if err := ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt); err != nil {
			global.LOG.Errorf("websocket writes control message failed, err: %v", err)
		}
		return true
	}
	return false
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
