package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

const mcpMysqlQueryTimeout = 30 * time.Second
const mcpMysqlQueryOutputLimit = 1 << 20

func mcpLocalMysqlQuery(ctx context.Context) *gorm.DB {
	return global.DB.WithContext(ctx).Table("databases AS d").
		Joins("JOIN app_installs AS ai ON ai.id = d.app_install_id AND ai.name = d.name").
		Joins("JOIN apps AS a ON a.id = ai.app_id").
		Where("d.type = ? AND d.\"from\" = ? AND a.key = ?", constant.AppMysql, "local", constant.AppMysql).
		Where("ai.status <> ? AND ai.container_name <> ?", constant.StatusDeleted, "")
}

func (u *DatabaseService) ListForMCP(ctx context.Context, dbType string) ([]dto.DatabaseOption, error) {
	if dbType != constant.AppMysql {
		return nil, errors.New("only installed local MySQL instances are supported")
	}
	result := make([]dto.DatabaseOption, 0)
	err := mcpLocalMysqlQuery(ctx).Select("d.id, d.type, d.\"from\", d.name AS database, ai.version").Order("d.id").Scan(&result).Error
	if err != nil {
		return nil, errors.New("cannot read local MySQL instances")
	}
	return result, nil
}

type mcpMysqlReadRunner func(context.Context, string, string, string) (string, error)

func (u *MysqlService) LoadStatusForMCP(ctx context.Context, req dto.OperationWithNameAndType) (*dto.MysqlStatus, error) {
	return u.loadStatusForMCP(ctx, req, runMCPMysqlRead)
}
func (u *MysqlService) loadStatusForMCP(ctx context.Context, req dto.OperationWithNameAndType, run mcpMysqlReadRunner) (*dto.MysqlStatus, error) {
	if req.Type != constant.AppMysql || req.Name == "" || len(req.Name) > 128 || strings.TrimSpace(req.Name) != req.Name || strings.ContainsAny(req.Name, "\x00\r\n/\\") || ctx.Err() != nil {
		return nil, errors.New("an explicitly named installed local MySQL instance is required")
	}
	ctx, cancel := context.WithTimeout(ctx, mcpMysqlQueryTimeout)
	defer cancel()
	var target struct {
		ID                          uint
		ContainerName, Version, Env string
	}
	if err := mcpLocalMysqlQuery(ctx).Where("d.name = ?", req.Name).Select("d.id, ai.container_name, ai.version, ai.env").Take(&target).Error; err != nil || target.ID == 0 {
		return nil, errors.New("the installed local MySQL instance is unavailable")
	}
	var env map[string]any
	if json.Unmarshal([]byte(target.Env), &env) != nil {
		return nil, errors.New("local MySQL configuration is unavailable")
	}
	password, ok := env["PANEL_DB_ROOT_PASSWORD"].(string)
	if !ok || password == "" {
		return nil, errors.New("local MySQL configuration is unavailable")
	}
	app := &repo.RootInfo{ContainerName: target.ContainerName, Version: target.Version, Key: constant.AppMysql, Password: password}
	rows := func(container, _, password, sql string) ([]string, error) {
		if ctx.Err() != nil {
			return nil, errors.New("local MySQL status query failed")
		}
		output, err := run(ctx, container, password, sql)
		if err != nil {
			if err == errMCPMysqlStopped || err == errMCPMysqlConnectionFailed {
				return nil, err
			}
			return nil, errors.New("local MySQL status query failed")
		}
		return strings.Split(output, "\n"), nil
	}
	maps := func(container, dbType, password, sql string) (map[string]string, error) {
		lines, err := rows(container, dbType, password, sql)
		if err != nil {
			return nil, err
		}
		values := map[string]string{}
		for _, line := range lines {
			if pair := strings.Split(line, "\t"); len(pair) == 2 {
				values[pair[0]] = pair[1]
			}
		}
		return values, nil
	}
	return loadMysqlStatus(req, app, maps, rows)
}

type mcpQueryOutput struct {
	buffer   bytes.Buffer
	mu       sync.Mutex
	exceeded bool
}

func (b *mcpQueryOutput) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(data) > mcpMysqlQueryOutputLimit-b.buffer.Len() {
		b.exceeded = true
		return 0, errors.New("query output limit exceeded")
	}
	return b.buffer.Write(data)
}
func (b *mcpQueryOutput) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Len()
}
func (b *mcpQueryOutput) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}
func runMCPMysqlRead(ctx context.Context, container, password, sql string) (string, error) {
	switch sql {
	case "show global status;", "show master status;", "show binary log status;":
	default:
		return "", errors.New("unsupported MySQL monitoring query")
	}
	command := exec.CommandContext(ctx, "docker", "exec", container, "mysql", "-uroot", "-p"+password, "-e", sql)
	output := &mcpQueryOutput{}
	var stderr bytes.Buffer
	command.Stdout, command.Stderr = output, io.MultiWriter(output, &stderr)
	runErr := command.Run()
	if ctx.Err() != nil || output.exceeded {
		return "", errors.New("local MySQL status query failed")
	}
	if runErr != nil {
		return "", classifyMCPMysqlReadFailure(stderr.String())
	}
	value := strings.ReplaceAll(output.String(), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if strings.HasPrefix(value, "ERROR ") {
		return "", errors.New("local MySQL status query failed")
	}
	return value, nil
}

func (a *AppInstallService) GetAppInstallInfoForMCP(ctx context.Context, id uint) (*response.AppInstallInfo, error) {
	var installed model.AppInstall
	if id == 0 {
		return nil, errors.New("invalid installed application")
	}
	err := global.DB.WithContext(ctx).Preload("App").Where("id = ?", id).First(&installed).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.AppInstallInfo{Status: constant.StatusDeleted}, nil
	}
	if err != nil {
		return nil, errors.New("installed application metadata is unavailable")
	}
	var env map[string]any
	if json.Unmarshal([]byte(installed.Env), &env) != nil {
		return nil, errors.New("installed application metadata is unavailable")
	}
	result := &response.AppInstallInfo{ID: installed.ID, Name: installed.Name, Version: installed.Version, Status: installed.Status, HttpPort: installed.HttpPort, AppKey: installed.App.Key, AppPorts: []int{}}
	for key, value := range env {
		if !strings.Contains(strings.ToUpper(key), "PANEL_APP_PORT") {
			continue
		}
		var port int
		switch item := value.(type) {
		case string:
			port, _ = strconv.Atoi(item)
		case float64:
			if item == float64(int(item)) {
				port = int(item)
			}
		}
		if port > 0 && port <= 65535 {
			result.AppPorts = append(result.AppPorts, port)
		}
	}
	sort.Ints(result.AppPorts)
	return result, nil
}
