package v2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

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
)

// closeCodeSessionNotFound is sent when the session is gone or not owned by the caller.
const closeCodeSessionNotFound = 4404

// maxTerminalTitleRunes caps the client supplied tab name.
const maxTerminalTitleRunes = 64

// @Tags Terminal
// @Summary Ws local terminal
// @Param command query string false "command"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/local [get]
func (b *BaseApi) WsLocalTerminal(c *gin.Context) {
	b.runSSHSession(c, sshSessionOption{
		kind:    terminal.SessionKindLocal,
		connect: loadLocalConn,
		command: c.DefaultQuery("command", ""),
	})
}

// @Tags Terminal
// @Summary Ws host SSH
// @Param id query integer false "id"
// @Param command query string false "command"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/ssh [get]
func (b *BaseApi) WsHostSSH(c *gin.Context) {
	hostID, _ := strconv.Atoi(c.DefaultQuery("id", "0"))
	b.runSSHSession(c, sshSessionOption{
		kind:   terminal.SessionKindSSH,
		hostID: uint(max(hostID, 0)),
		connect: func() (*ssh.SSHClient, error) {
			if hostID <= 0 {
				return nil, errors.New("missing host id")
			}
			host, err := service.GetHostInfo(uint(hostID))
			return newHostSSHClient(host, err)
		},
		command: c.DefaultQuery("command", ""),
	})
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

type sshSessionOption struct {
	kind    string
	hostID  uint
	connect func() (*ssh.SSHClient, error)
	command string
}

func (b *BaseApi) runSSHSession(c *gin.Context, opt sshSessionOption) {
	wsConn, cols, rows, ok := prepareTerminalSession(c)
	if !ok {
		return
	}
	defer wsConn.Close()

	owner := panelUser(c)
	if sessionID := strings.TrimSpace(c.Query("session")); len(sessionID) != 0 {
		attachTerminalSession(wsConn, sessionID, owner, cols, rows)
		return
	}

	client, clientErr := opt.connect()
	if wshandleError(wsConn, errors.WithMessage(clientErr, "failed to set up the connection. Please check the host information")) {
		return
	}

	sess, err := terminal.DefaultManager.OpenSession(client.Client, terminal.SessionOptions{
		Kind:    opt.kind,
		HostID:  opt.hostID,
		Title:   sanitizeTerminalTitle(c.Query("title")),
		Owner:   owner,
		Cols:    cols,
		Rows:    rows,
		InitCmd: opt.command,
	})
	if err != nil {
		// session does not exist yet, so we still own the ssh client
		client.Close()
		_ = wshandleError(wsConn, err)
		return
	}
	// no defer sess.Close(): pinned sessions outlive this websocket
	att, err := sess.Attach(wsConn, cols, rows)
	if err != nil {
		sess.Close()
		_ = wshandleError(wsConn, err)
		return
	}
	att.Run()

	closeTerminalConn(wsConn)
}

// attachTerminalSession binds the websocket to an existing session.
func attachTerminalSession(wsConn *websocket.Conn, sessionID, owner string, cols, rows int) {
	sess, err := terminal.DefaultManager.Lookup(sessionID, owner)
	if err != nil {
		closeTerminalConnWithCode(wsConn, closeCodeSessionNotFound, "session not found")
		return
	}
	att, err := sess.Attach(wsConn, cols, rows)
	if err != nil {
		global.LOG.Errorf("attach terminal session %s failed, err: %v", sessionID, err)
		closeTerminalConnWithCode(wsConn, closeCodeSessionNotFound, "session not found")
		return
	}
	att.Run()

	closeTerminalConn(wsConn)
}

// sanitizeTerminalTitle keeps a client supplied tab name printable and short.
func sanitizeTerminalTitle(title string) string {
	var builder strings.Builder
	count := 0
	for _, r := range strings.TrimSpace(title) {
		if unicode.IsControl(r) {
			continue
		}
		if count == maxTerminalTitleRunes {
			break
		}
		builder.WriteRune(r)
		count++
	}
	return strings.TrimSpace(builder.String())
}

// @Tags Terminal
// @Summary Search terminal sessions
// @Accept json
// @Success 200 {array} dto.TerminalSessionInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/sessions/search [post]
func (b *BaseApi) SearchTerminalSessions(c *gin.Context) {
	list, err := terminalSessionService.List(panelUser(c))
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Terminal
// @Summary Pin or unpin a terminal session
// @Accept json
// @Param request body dto.TerminalSessionPin true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/terminal/sessions/pin [post]
func (b *BaseApi) PinTerminalSession(c *gin.Context) {
	var req dto.TerminalSessionPin
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := terminalSessionService.Pin(panelUser(c), req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
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
	if err := terminalSessionService.Close(panelUser(c), req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func closeTerminalConn(wsConn *websocket.Conn) {
	dt := time.Now().Add(time.Second)
	_ = wsConn.WriteControl(websocket.CloseMessage, nil, dt)
}

// closeTerminalConnWithCode reports a terminal specific failure to the client.
func closeTerminalConnWithCode(wsConn *websocket.Conn, code int, reason string) {
	dt := time.Now().Add(time.Second)
	_ = wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), dt)
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
