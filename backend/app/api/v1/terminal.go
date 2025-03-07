package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/copier"
	"github.com/1Panel-dev/1Panel/backend/utils/ssh"
	"github.com/1Panel-dev/1Panel/backend/utils/terminal"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

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

	client, err := connInfo.NewClient()
	if wshandleError(wsConn, errors.WithMessage(err, "failed to set up the connection. Please check the host information")) {
		return
	}
	defer client.Close()
	sws, err := terminal.NewLogicSshWsSession(cols, rows, true, connInfo.Client, wsConn)
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

func (b *BaseApi) ContainerWsSSH(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	if global.CONF.System.IsDemo {
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
	source := c.Query("source")
	var containerID string
	var initCmd string
	switch source {
	case "redis":
		containerID, initCmd, err = loadRedisInitCmd(c)
	case "ollama":
		containerID, initCmd, err = loadOllamaInitCmd(c)
	case "container":
		containerID, initCmd, err = loadContainerInitCmd(c)
	default:
		if wshandleError(wsConn, fmt.Errorf("not support such source %s", source)) {
			return
		}
	}
	if wshandleError(wsConn, err) {
		return
	}
	pidMap := loadMapFromDockerTop(containerID)
	slave, err := terminal.NewCommand("clear && " + initCmd)
	if wshandleError(wsConn, err) {
		return
	}
	defer killBash(containerID, strings.ReplaceAll(initCmd, fmt.Sprintf("docker exec -it %s ", containerID), ""), pidMap)
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

func loadRedisInitCmd(c *gin.Context) (string, string, error) {
	name := c.Query("name")
	from := c.Query("from")
	commands := "redis-cli"
	database, err := databaseService.Get(name)
	if err != nil {
		return "", "", fmt.Errorf("no such database in db, err: %v", err)
	}
	if from == "local" {
		redisInfo, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Name: name, Type: "redis"})
		if err != nil {
			return "", "", fmt.Errorf("no such app in db, err: %v", err)
		}
		name = redisInfo.ContainerName
		if len(database.Password) != 0 {
			commands = "redis-cli -a " + database.Password + " --no-auth-warning"
		}
	} else {
		commands = fmt.Sprintf("redis-cli -h %s -p %v", database.Address, database.Port)
		if len(database.Password) != 0 {
			commands = fmt.Sprintf("redis-cli -h %s -p %v -a %s --no-auth-warning", database.Address, database.Port, database.Password)
		}
		name = "1Panel-redis-cli-tools"
	}
	return name, fmt.Sprintf("docker exec -it %s %s", name, commands), nil
}

func loadOllamaInitCmd(c *gin.Context) (string, string, error) {
	name := c.Query("name")
	if cmd.CheckIllegal(name) {
		return "", "", fmt.Errorf("ollama model %s contains illegal characters", name)
	}
	ollamaInfo, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Name: "", Type: "ollama"})
	if err != nil {
		return "", "", fmt.Errorf("no such app in db, err: %v", err)
	}
	containerName := ollamaInfo.ContainerName
	return containerName, fmt.Sprintf("docker exec -it %s ollama run %s", containerName, name), nil
}

func loadContainerInitCmd(c *gin.Context) (string, string, error) {
	containerID := c.Query("containerid")
	command := c.Query("command")
	user := c.Query("user")
	if cmd.CheckIllegal(user, containerID, command) {
		return "", "", fmt.Errorf("the command contains illegal characters. command: %s, user: %s, containerID: %s", command, user, containerID)
	}
	if len(command) == 0 || len(containerID) == 0 {
		return "", "", fmt.Errorf("error param of command: %s or containerID: %s", command, containerID)
	}
	command = fmt.Sprintf("docker exec -it %s %s", containerID, command)
	if len(user) != 0 {
		command = fmt.Sprintf("docker exec -it -u %s %s %s", user, containerID, command)
	}

	return containerID, command, nil
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
