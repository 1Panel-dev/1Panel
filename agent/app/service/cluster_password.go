package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	clusterSQLPassword    = "PANEL_DB_ROOT_PASSWORD"
	clusterRedisPassword  = "PANEL_REDIS_ROOT_PASSWORD"
	clusterSourcePassword = "MASTER_ROOT_PASSWORD"
)

type ClusterPasswordMemberRequest struct {
	InstallID       uint     `json:"installID" validate:"required"`
	App             string   `json:"app" validate:"required"`
	Template        string   `json:"template" validate:"required"`
	Role            string   `json:"role" validate:"oneof=master slave"`
	Action          string   `json:"action" validate:"oneof=check rotate sync"`
	CurrentPassword string   `json:"currentPassword"`
	NewPassword     string   `json:"newPassword" validate:"required"`
	SourceID        string   `json:"sourceID"`
	RootHosts       []string `json:"rootHosts"`
}

type ClusterPasswordMemberResponse struct {
	Identity  string   `json:"identity"`
	RootHosts []string `json:"rootHosts,omitempty"`
}

func UpdateClusterMemberPassword(ctx context.Context, req ClusterPasswordMemberRequest) (ClusterPasswordMemberResponse, error) {
	var result ClusterPasswordMemberResponse
	if !validClusterPasswordRequest(req) {
		return result, buserr.New("ErrClusterPasswordRequestInvalid")
	}
	unlock, err := lockClusterDataOperation(model.AppInstall{BaseModel: model.BaseModel{ID: req.InstallID}})
	if err != nil {
		return result, err
	}
	defer unlock()
	install, err := appInstallRepo.GetSnapshot(ctx, req.InstallID)
	if err != nil || install.ID != req.InstallID || install.App.Key != req.App || install.Version != req.Template ||
		!strings.HasSuffix(install.Version, "-"+req.Role) {
		return result, buserr.New("ErrClusterPasswordMemberUnavailable")
	}
	if err := guardClusterRestore(install); err != nil {
		return result, buserr.New("TaskIsExecuting")
	}
	if install.Param != "" {
		var values map[string]json.RawMessage
		if err := json.Unmarshal([]byte(install.Param), &values); err != nil || values == nil {
			return result, buserr.New("ErrClusterPasswordConfiguration")
		}
	}
	db, err := newClusterDatabase(install)
	if err != nil {
		return result, buserr.New("ErrClusterPasswordConfiguration")
	}
	unlockContainer, err := containerOperationLock.lockContext(ctx, install.ContainerName)
	if err != nil {
		return result, buserr.New("TaskIsExecuting")
	}
	defer unlockContainer()
	cli, err := docker.NewDockerClientWithContext(ctx)
	if err != nil {
		return result, buserr.New("ErrClusterPasswordMemberUnavailable")
	}
	defer cli.Close()
	db.dockerHost = cli.DaemonHost()
	info, err := cli.ContainerInspect(ctx, install.ContainerName)
	if err != nil || info.ContainerJSONBase == nil || info.Config == nil || info.State == nil ||
		!info.State.Running || info.State.Paused || info.State.Restarting || info.State.Dead {
		return result, buserr.New("ErrClusterPasswordMemberUnavailable")
	}
	mode := map[string]string{
		constant.AppMysqlCluster: "MYSQL", constant.AppPostgresqlCluster: "POSTGRESQL", constant.AppRedisCluster: "REDIS",
	}[req.App]
	if !containsRestoreEnv(info.Config.Env, mode+"_REPLICATION_MODE="+req.Role) {
		return result, buserr.New("ErrClusterPasswordSourceMismatch")
	}
	localKey, changes := clusterPasswordChanges(req)
	password, result, err := loadClusterPasswordMember(ctx, db, req, db.env[localKey])
	if err != nil {
		return result, err
	}
	if len(req.RootHosts) > 0 && !slices.Equal(result.RootHosts, req.RootHosts) {
		return result, buserr.New("ErrClusterPasswordMySQLAccounts")
	}
	forceRecreate := false
	if req.App == constant.AppRedisCluster {
		if req.Role == "master" {
			forceRecreate = password != req.NewPassword
		} else {
			upstream, err := redisClusterPasswordQuery(ctx, db, password, "CONFIG", "GET", "masterauth")
			if err != nil {
				return result, buserr.New("ErrClusterPasswordConfiguration")
			}
			forceRecreate = upstream != "masterauth\n"+req.NewPassword
		}
	}
	config, err := prepareClusterPasswordConfig(ctx, cli, db, info, req, changes, forceRecreate)
	if err != nil {
		return result, err
	}
	if config.recreate && req.App == constant.AppRedisCluster {
		if err := ensureRedisPasswordPersistence(ctx, db, password); err != nil {
			return result, err
		}
	}
	if req.Action == "check" {
		return result, nil
	}
	if config.recreate {
		current, err := cli.ContainerInspect(ctx, install.ContainerName)
		if err != nil || current.ContainerJSONBase == nil || current.ID != info.ID {
			return result, buserr.New("ErrClusterPasswordMemberUnavailable")
		}
		if !config.matchesPorts(current.HostConfig) {
			return result, buserr.New("ErrClusterPasswordPortBindings")
		}
	}
	if req.Action == "rotate" && password != req.NewPassword && req.App != constant.AppRedisCluster {
		if err := rotateClusterSQLPassword(ctx, db, req, password, result.RootHosts); err != nil {
			return result, err
		}
	}
	if req.App != constant.AppRedisCluster && password != req.NewPassword {
		password = req.NewPassword
		verify := req
		verify.SourceID = result.Identity
		result, err = waitClusterPasswordMember(ctx, db, verify, password)
		if err != nil {
			return result, err
		}
	}

	path := install.GetEnvPath()
	if !bytes.Equal(config.previousEnv, config.nextEnv) {
		current, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(current, config.previousEnv) {
			return result, buserr.New("ErrClusterPasswordConfiguration")
		}
		if err := writeUpgradeFile(path, config.nextEnv, 0600); err != nil {
			return result, buserr.New("ErrClusterPasswordSave")
		}
	}
	if config.recreate {
		if req.App == constant.AppRedisCluster {
			stop, cancel := context.WithTimeout(ctx, 5*time.Minute)
			timeout := -1
			err := cli.ContainerStop(stop, info.ID, container.StopOptions{Timeout: &timeout})
			cancel()
			if err != nil {
				return result, buserr.New("ErrClusterPasswordUnconfirmed")
			}
		}
		if err := compose.RecreatePasswordService(ctx, install.GetComposePath(), config.project, config.service, db.dockerHost, config.environment); err != nil {
			if req.App == constant.AppRedisCluster {
				restartStoppedClusterPasswordContainer(ctx, cli, info, err)
			}
			return result, buserr.New("ErrClusterPasswordUnconfirmed")
		}
		current, err := cli.ContainerInspect(ctx, install.ContainerName)
		if err != nil || current.ContainerJSONBase == nil || !config.matchesPorts(current.HostConfig) {
			return result, buserr.New("ErrClusterPasswordUnconfirmed")
		}
		verify := req
		if req.Role == "master" {
			verify.SourceID = result.Identity
			if req.App == constant.AppRedisCluster {
				verify.SourceID = ""
				password = req.NewPassword
			}
		}
		result, err = waitClusterPasswordMember(ctx, db, verify, password)
		if err != nil {
			return result, err
		}
		if err := verifyClusterPasswordAuthentication(ctx, db, req.App); err != nil {
			return result, err
		}
	} else if req.App == constant.AppRedisCluster && req.Role == "slave" {
		result, err = waitClusterPasswordMember(ctx, db, req, password)
		if err != nil {
			return result, err
		}
	}
	if err := repo.SaveClusterDatabaseCredentials(ctx, install, localKey, config.changes); err != nil {
		return result, buserr.New("ErrClusterPasswordSave")
	}
	return result, nil
}

func restartStoppedClusterPasswordContainer(ctx context.Context, cli *client.Client, original container.InspectResponse, composeErr error) {
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(composeErr, &exit) || !exit.Exited() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	current, err := cli.ContainerInspect(ctx, original.ID)
	if err == nil && current.ContainerJSONBase != nil && current.ID == original.ID && current.Name == original.Name &&
		current.State != nil && current.State.Status == "exited" && !current.State.Dead {
		_ = cli.ContainerStart(ctx, original.ID, container.StartOptions{})
	}
}

func validClusterPasswordRequest(req ClusterPasswordMemberRequest) bool {
	return req.InstallID > 0 && constant.IsClusterApp(req.App) && req.Template != "" &&
		(req.Role == "master" || req.Role == "slave" && req.SourceID != "") &&
		(req.Action == "check" || req.Action == "sync" || req.Action == "rotate" && req.Role == "master") &&
		validClusterPassword(req.NewPassword) && (req.CurrentPassword == "" || validClusterPassword(req.CurrentPassword)) &&
		(req.App != constant.AppRedisCluster || !strings.ContainsFunc(req.NewPassword, unicode.IsSpace))
}

func validClusterPassword(password string) bool {
	return password != "" && len(password) <= 512 && utf8.ValidString(password) &&
		!strings.ContainsFunc(password, unicode.IsControl)
}

func clusterPasswordChanges(req ClusterPasswordMemberRequest) (string, map[string]string) {
	if req.App == constant.AppRedisCluster {
		if req.Role == "slave" {
			return clusterRedisPassword, map[string]string{clusterSourcePassword: req.NewPassword}
		}
		return clusterRedisPassword, map[string]string{clusterRedisPassword: req.NewPassword}
	}
	changes := map[string]string{clusterSQLPassword: req.NewPassword}
	if req.App == constant.AppMysqlCluster && req.Role == "slave" {
		changes[clusterSourcePassword] = req.NewPassword
	}
	return clusterSQLPassword, changes
}
