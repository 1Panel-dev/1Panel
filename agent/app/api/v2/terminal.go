package v2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/ssh"
	"github.com/1Panel-dev/1Panel/agent/utils/terminal"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	gossh "golang.org/x/crypto/ssh"
)

// @Tags Terminal
// @Summary Ws local terminal
// @Param command query string false "command"
// @Param session query string false "session id to reattach"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/local [get]
func (b *BaseApi) WsLocalTerminal(c *gin.Context) {
	b.runSSHSession(c, loadLocalConn, c.DefaultQuery("command", ""))
}

// @Tags Terminal
// @Summary Ws host SSH
// @Param id query integer false "id"
// @Param command query string false "command"
// @Param session query string false "session id to reattach"
// @Param title query string false "session title shown in the session list"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/ssh [get]
func (b *BaseApi) WsHostSSH(c *gin.Context) {
	b.runSSHSession(c, func() (*ssh.SSHClient, error) {
		hostID, _ := strconv.Atoi(c.DefaultQuery("id", "0"))
		if hostID <= 0 {
			return nil, errors.New("missing host id")
		}
		host, err := service.GetHostInfo(uint(hostID))
		return newHostSSHClient(host, err)
	}, c.DefaultQuery("command", ""))
}

// @Tags Terminal
// @Summary Ws container terminal
// @Param cols query integer false "cols"
// @Param rows query integer false "rows"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/container [get]
func (b *BaseApi) WsContainerTerminal(c *gin.Context) {
	wsConn, cols, rows, ok := prepareTerminalSession(c)
	if !ok {
		return
	}
	defer wsConn.Close()

	slave, err := loadContainerTerminalCommand(c)
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
	closeTerminalConn(wsConn)
}

func prepareTerminalSession(c *gin.Context) (*websocket.Conn, int, int, bool) {
	if !websocket.IsWebSocketUpgrade(c.Request) {
		helper.Success(c)
		return nil, 0, 0, false
	}
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return nil, 0, 0, false
	}

	if global.CONF.Base.IsDemo {
		if wshandleError(wsConn, errors.New("   demo server, prohibit this operation!")) {
			return nil, 0, 0, false
		}
	}

	cols, err := strconv.Atoi(c.DefaultQuery("cols", "80"))
	if wshandleError(wsConn, errors.WithMessage(err, "invalid param cols in request")) {
		return nil, 0, 0, false
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "40"))
	if wshandleError(wsConn, errors.WithMessage(err, "invalid param rows in request")) {
		return nil, 0, 0, false
	}
	return wsConn, cols, rows, true
}

func (b *BaseApi) runSSHSession(c *gin.Context, connect func() (*ssh.SSHClient, error), command string) {
	wsConn, cols, rows, ok := prepareTerminalSession(c)
	if !ok {
		return
	}
	defer wsConn.Close()

	hostID, _ := strconv.Atoi(c.DefaultQuery("id", "0"))
	opts := terminal.SessionOptions{
		Owner:   loadAuditUser(c),
		Title:   sanitizeTerminalTitle(c.Query("title")),
		HostID:  uint(max(hostID, 0)),
		Cols:    cols,
		Rows:    rows,
		InitCmd: command,
	}
	err := terminal.Serve(wsConn, strings.TrimSpace(c.Query("session")), opts, func() (*gossh.Client, error) {
		client, err := connect()
		if err != nil {
			return nil, errors.WithMessage(err, "failed to set up the connection. Please check the host information")
		}
		return client.Client, nil
	})
	if err != nil {
		_ = wshandleError(wsConn, err)
	}
}

// @Tags Terminal
// @Summary List the caller's live terminal sessions
// @Success 200 {array} terminal.Info
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/sessions/search [post]
func (b *BaseApi) SearchTerminalSessions(c *gin.Context) {
	helper.SuccessWithData(c, terminal.List(loadAuditUser(c)))
}

// @Tags Terminal
// @Summary Close a terminal session
// @Accept json
// @Param request body dto.TerminalSessionClose true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/sessions/close [post]
func (b *BaseApi) CloseTerminalSession(c *gin.Context) {
	var req dto.TerminalSessionClose
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := terminal.CloseSession(req.ID, loadAuditUser(c)); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Terminal
// @Summary Close every terminal session (panel user logged out)
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/sessions/closeAll [post]
func (b *BaseApi) CloseAllTerminalSessions(c *gin.Context) {
	terminal.CloseAll()
	helper.Success(c)
}

// sanitizeTerminalTitle keeps the title a short single line.
func sanitizeTerminalTitle(title string) string {
	title = strings.Join(strings.Fields(title), " ")
	if r := []rune(title); len(r) > 64 {
		title = string(r[:64])
	}
	return title
}

func closeTerminalConn(wsConn *websocket.Conn) {
	dt := time.Now().Add(time.Second)
	_ = wsConn.WriteControl(websocket.CloseMessage, nil, dt)
}

func newHostSSHClient(host *model.Host, err error) (*ssh.SSHClient, error) {
	if err != nil {
		return nil, errors.WithMessage(err, "load host info by id failed")
	}
	connInfo := ssh.ConnInfo{
		Addr:       host.Addr,
		Port:       int(host.Port),
		User:       host.User,
		AuthMode:   host.AuthMode,
		Password:   host.Password,
		PrivateKey: []byte(host.PrivateKey),
	}
	if len(host.PassPhrase) != 0 {
		connInfo.PassPhrase = []byte(host.PassPhrase)
	}
	return ssh.NewClient(connInfo)
}

func loadContainerTerminalCommand(c *gin.Context) (*terminal.LocalCommand, error) {
	source := c.Query("source")
	var (
		initCmd []string
		err     error
	)
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
		return nil, fmt.Errorf("not support such source %s", source)
	}
	if err != nil {
		return nil, err
	}
	return terminal.NewCommand("docker", initCmd...)
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
	if len(databaseType) == 0 {
		return nil, fmt.Errorf("error param of database: %s or database type: %s", database, databaseType)
	}
	databaseConn, err := appInstallService.LoadConnInfo(dto.OperationWithNameAndType{Type: databaseType, Name: database})
	if err != nil {
		return nil, fmt.Errorf("no such database in db, err: %v", err)
	}
	if len(databaseConn.ContainerName) == 0 {
		return nil, fmt.Errorf("no such database container for database: %s or database type: %s", database, databaseType)
	}
	commands := []string{"exec", "-it", databaseConn.ContainerName}
	switch databaseType {
	case "mysql", "mysql-cluster":
		commands = append(commands, []string{"mysql", "-uroot", "-p" + databaseConn.Password}...)
	case "mariadb":
		commands = append(commands, []string{"mariadb", "-uroot", "-p" + databaseConn.Password}...)
	case "mongodb":
		commands = append(commands, []string{
			"mongosh",
			"--username", databaseConn.Username,
			"--password", databaseConn.Password,
			"--authenticationDatabase", "admin",
		}...)
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
	ReadBufferSize:  4096,
	WriteBufferSize: 16384,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
