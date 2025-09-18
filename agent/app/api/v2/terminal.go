package v2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

	dt := time.Now().Add(time.Second)
	_ = wsConn.WriteControl(websocket.CloseMessage, nil, dt)
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
	var initCmd []string
	switch source {
	case "redis", "redis-cluster":
		initCmd, err = loadRedisInitCmd(c, source)
	case "ollama":
		initCmd, err = loadOllamaInitCmd(c)
	case "container":
		initCmd, err = loadContainerInitCmd(c)
	case "database":
		initCmd, err = loadDatabaseInitCmd(c)
	default:
		if wshandleError(wsConn, fmt.Errorf("not support such source %s", source)) {
			return
		}
	}
	if wshandleError(wsConn, err) {
		return
	}
	slave, err := terminal.NewCommand("docker", initCmd...)
	if wshandleError(wsConn, err) {
		return
	}
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
	dt := time.Now().Add(time.Second)
	_ = wsConn.WriteControl(websocket.CloseMessage, nil, dt)
}

func loadRedisInitCmd(c *gin.Context, redisType string) ([]string, error) {
	name := c.Query("name")
	from := c.Query("from")
	commands := []string{"exec", "-it"}
	database, err := databaseService.Get(name)
	if err != nil {
		return nil, fmt.Errorf("no such database in db, err: %v", err)
	}
	if from == "local" {
		redisInfo, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Name: name, Type: redisType})
		if err != nil {
			return nil, fmt.Errorf("no such app in db, err: %v", err)
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
	return commands, nil
}

func loadOllamaInitCmd(c *gin.Context) ([]string, error) {
	name := c.Query("name")
	if cmd.CheckIllegal(name) {
		return nil, fmt.Errorf("ollama model %s contains illegal characters", name)
	}
	ollamaInfo, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Name: "", Type: "ollama"})
	if err != nil {
		return nil, fmt.Errorf("no such app in db, err: %v", err)
	}
	containerName := ollamaInfo.ContainerName
	return []string{"exec", "-it", containerName, "ollama", "run", name}, nil
}

func loadContainerInitCmd(c *gin.Context) ([]string, error) {
	containerID := c.Query("containerid")
	command := c.Query("command")
	user := c.Query("user")
	if cmd.CheckIllegal(user, containerID, command) {
		return nil, fmt.Errorf("the command contains illegal characters. command: %s, user: %s, containerID: %s", command, user, containerID)
	}
	if len(command) == 0 || len(containerID) == 0 {
		return nil, fmt.Errorf("error param of command: %s or containerID: %s", command, containerID)
	}
	commands := []string{"exec", "-it", containerID, command}
	if len(user) != 0 {
		commands = []string{"exec", "-it", "-u", user, containerID, command}
	}

	return commands, nil
}

func loadDatabaseInitCmd(c *gin.Context) ([]string, error) {
	database := c.Query("database")
	databaseType := c.Query("databaseType")
	if len(database) == 0 || len(databaseType) == 0 {
		return nil, fmt.Errorf("error param of database: %s or database type: %s", database, databaseType)
	}
	databaseConn, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Type: databaseType, Name: database})
	if err != nil {
		return nil, fmt.Errorf("no such database in db, err: %v", err)
	}
	commands := []string{"exec", "-it", databaseConn.ContainerName}
	switch databaseType {
	case "mysql", "mysql-cluster":
		commands = append(commands, []string{"mysql", "-uroot", "-p" + databaseConn.Password}...)
	case "mariadb":
		commands = append(commands, []string{"mariadb", "-uroot", "-p" + databaseConn.Password}...)
	case "postgresql", "postgresql-cluster":
		commands = []string{"exec", "-e", fmt.Sprintf("PGPASSWORD=%s", databaseConn.Password), "-it", databaseConn.ContainerName, "psql", "-t", "-U", databaseConn.Username}
	}

	return commands, nil
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

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
