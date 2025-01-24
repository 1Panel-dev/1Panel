package v2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/terminal"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func (b *BaseApi) RedisWsSsh(c *gin.Context) {
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
	name := c.Query("name")
	from := c.Query("from")
	commands := []string{"redis-cli"}
	database, err := databaseService.Get(name)
	if wshandleError(wsConn, errors.WithMessage(err, "no such database in db")) {
		return
	}
	if from == "local" {
		redisInfo, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Name: name, Type: "redis"})
		if wshandleError(wsConn, errors.WithMessage(err, "no such database in db")) {
			return
		}
		name = redisInfo.ContainerName
		if len(database.Password) != 0 {
			commands = []string{"redis-cli", "-a", database.Password, "--no-auth-warning"}
		}
	} else {
		itemPort := fmt.Sprintf("%v", database.Port)
		commands = []string{"redis-cli", "-h", database.Address, "-p", itemPort}
		if len(database.Password) != 0 {
			commands = []string{"redis-cli", "-h", database.Address, "-p", itemPort, "-a", database.Password, "--no-auth-warning"}
		}
		name = "1Panel-redis-cli-tools"
	}

	pidMap := loadMapFromDockerTop(name)
	itemCmds := append([]string{"exec", "-it", name}, commands...)
	slave, err := terminal.NewCommand(itemCmds)
	if wshandleError(wsConn, err) {
		return
	}
	defer killBash(name, strings.Join(commands, " "), pidMap)
	defer slave.Close()

	tty, err := terminal.NewLocalWsSession(cols, rows, wsConn, slave, false)
	if wshandleError(wsConn, err) {
		return
	}

	quitChan := make(chan bool, 3)
	tty.Start(quitChan)
	go slave.Wait(quitChan)

	<-quitChan

	global.LOG.Info("websocket finished")
	if wshandleError(wsConn, err) {
		return
	}
}

func (b *BaseApi) ContainerWsSsh(c *gin.Context) {
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

	containerID := c.Query("containerid")
	command := c.Query("command")
	user := c.Query("user")
	if len(command) == 0 || len(containerID) == 0 {
		if wshandleError(wsConn, errors.New("error param of command or containerID")) {
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

	cmds := []string{"exec", containerID, command}
	if len(user) != 0 {
		cmds = []string{"exec", "-u", user, containerID, command}
	}
	if cmd.CheckIllegal(user, containerID, command) {
		if wshandleError(wsConn, errors.New("  The command contains illegal characters.")) {
			return
		}
	}
	stdout, err := cmd.ExecWithCheck("docker", cmds...)
	if wshandleError(wsConn, errors.WithMessage(err, stdout)) {
		return
	}

	commands := []string{"exec", "-it", containerID, command}
	if len(user) != 0 {
		commands = []string{"exec", "-it", "-u", user, containerID, command}
	}
	pidMap := loadMapFromDockerTop(containerID)
	slave, err := terminal.NewCommand(commands)
	if wshandleError(wsConn, err) {
		return
	}
	defer killBash(containerID, command, pidMap)
	defer slave.Close()

	tty, err := terminal.NewLocalWsSession(cols, rows, wsConn, slave, true)
	if wshandleError(wsConn, err) {
		return
	}

	quitChan := make(chan bool, 3)
	tty.Start(quitChan)
	go slave.Wait(quitChan)

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

func loadMapFromDockerTop(containerID string) map[string]string {
	pidMap := make(map[string]string)
	sudo := cmd.SudoHandleCmd()

	stdout, err := cmd.Execf("%s docker top %s -eo pid,command ", sudo, containerID)
	if err != nil {
		return pidMap
	}
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		pidMap[parts[0]] = strings.Join(parts[1:], " ")
	}
	return pidMap
}

func killBash(containerID, comm string, pidMap map[string]string) {
	sudo := cmd.SudoHandleCmd()
	newPidMap := loadMapFromDockerTop(containerID)
	for pid, command := range newPidMap {
		isOld := false
		for pid2 := range pidMap {
			if pid == pid2 {
				isOld = true
				break
			}
		}
		if !isOld && command == comm {
			_, _ = cmd.Execf("%s kill -9 %s", sudo, pid)
		}
	}
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
