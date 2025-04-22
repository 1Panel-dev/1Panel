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

func (b *BaseApi) WsSSH(c *gin.Context) {
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

	client, err := loadLocalConn()
	if wshandleError(wsConn, errors.WithMessage(err, "failed to set up the connection. Please check the host information")) {
		return
	}
	defer client.Close()
	command := c.DefaultQuery("command", "")
	sws, err := terminal.NewLogicSshWsSession(cols, rows, client.Client, wsConn, command)
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

func (b *BaseApi) ContainerWsSSH(c *gin.Context) {
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
	source := c.Query("source")
	var containerID string
	var initCmd []string
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
	slave, err := terminal.NewCommand("docker", initCmd...)
	if wshandleError(wsConn, err) {
		return
	}
	defer killBash(containerID, strings.ReplaceAll(strings.Join(initCmd, " "), fmt.Sprintf("exec -it %s ", containerID), ""), pidMap)
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

func loadRedisInitCmd(c *gin.Context) (string, []string, error) {
	name := c.Query("name")
	from := c.Query("from")
	commands := []string{"exec", "-it"}
	database, err := databaseService.Get(name)
	if err != nil {
		return "", nil, fmt.Errorf("no such database in db, err: %v", err)
	}
	if from == "local" {
		redisInfo, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Name: name, Type: "redis"})
		if err != nil {
			return "", nil, fmt.Errorf("no such app in db, err: %v", err)
		}
		name = redisInfo.ContainerName
		commands = append(commands, []string{name, "redis-cli"}...)
		if len(database.Password) != 0 {
			commands = append(commands, []string{"-a", database.Password, "--no-auth-warning"}...)
		}
	} else {
		name = "1Panel-redis-cli-tools"
		commands = append(commands, []string{name, "redis-cli", "-h", database.Address, "-p", fmt.Sprintf("%v", database.Port)}...)
		if len(database.Password) != 0 {
			commands = append(commands, []string{"-a", database.Password, "--no-auth-warning"}...)
		}
	}
	return name, commands, nil
}

func loadOllamaInitCmd(c *gin.Context) (string, []string, error) {
	name := c.Query("name")
	if cmd.CheckIllegal(name) {
		return "", nil, fmt.Errorf("ollama model %s contains illegal characters", name)
	}
	ollamaInfo, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Name: "", Type: "ollama"})
	if err != nil {
		return "", nil, fmt.Errorf("no such app in db, err: %v", err)
	}
	containerName := ollamaInfo.ContainerName
	return containerName, []string{"exec", "-it", containerName, "ollama", "run", name}, nil
}

func loadContainerInitCmd(c *gin.Context) (string, []string, error) {
	containerID := c.Query("containerid")
	command := c.Query("command")
	user := c.Query("user")
	if cmd.CheckIllegal(user, containerID, command) {
		return "", nil, fmt.Errorf("the command contains illegal characters. command: %s, user: %s, containerID: %s", command, user, containerID)
	}
	if len(command) == 0 || len(containerID) == 0 {
		return "", nil, fmt.Errorf("error param of command: %s or containerID: %s", command, containerID)
	}
	commands := []string{"exec", "-it", containerID, command}
	if len(user) != 0 {
		commands = []string{"exec", "-it", "-u", user, containerID, command}
	}

	return containerID, commands, nil
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

	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s docker top %s -eo pid,command ", sudo, containerID)
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
			_, _ = cmd.RunDefaultWithStdoutBashCf("%s kill -9 %s", sudo, pid)
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
