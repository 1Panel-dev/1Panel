package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/clusterarchive"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/google/uuid"
)

const originalRestoreWork = ".1panel-restore-active"

type OriginalClusterRestoreRequest struct {
	InstallID     uint   `json:"installID" validate:"required"`
	TaskID        string `json:"taskID" validate:"required"`
	Action        string `json:"action" validate:"required"`
	App           string `json:"app" validate:"required"`
	Role          string `json:"role" validate:"required"`
	Template      string `json:"template" validate:"required"`
	EngineVersion string `json:"engineVersion" validate:"required"`
	Artifact      string `json:"artifact"`
	Passphrase    string `json:"passphrase"`
}

type OriginalClusterRestoreResult struct {
	WasRunning   bool   `json:"wasRunning"`
	PreviousData string `json:"previousData,omitempty"`
}

type originalRestoreState struct {
	TaskID           string
	ContainerID      string
	WasRunning       bool
	RestartPolicy    container.RestartPolicy
	Prepared         bool
	Replacing        bool
	PublishAttempted bool
}

func lockClusterDataOperation(install model.AppInstall) (func(), error) {
	if _, busy := appUpgradeLocks.LoadOrStore(install.ID, struct{}{}); busy {
		return nil, buserr.New("TaskIsExecuting")
	}
	return func() { appUpgradeLocks.Delete(install.ID) }, nil
}

func guardClusterRestore(install model.AppInstall) error {
	_, err := os.Lstat(filepath.Join(install.GetPath(), originalRestoreWork))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return buserr.New("TaskIsExecuting")
}

func RestoreOriginalClusterMember(ctx context.Context, req OriginalClusterRestoreRequest) (OriginalClusterRestoreResult, error) {
	var result OriginalClusterRestoreResult
	id, err := uuid.Parse(req.TaskID)
	if err != nil || id == uuid.Nil || id.String() != req.TaskID ||
		(req.Role != "master" && req.Role != "slave") || !strings.HasSuffix(req.Template, "-"+req.Role) {
		return result, buserr.New("ErrClusterRestoreOriginalRequest")
	}
	install, err := appInstallRepo.GetFirst(repo.WithByID(req.InstallID))
	if err != nil {
		return result, err
	}
	if install.ID != req.InstallID || install.App.Key != req.App || install.Version != req.Template || !constant.IsClusterApp(req.App) {
		return result, buserr.New("ErrClusterRestoreMemberVersionChanged")
	}
	unlock, err := lockClusterDataOperation(install)
	if err != nil {
		return result, err
	}
	defer unlock()
	install, err = appInstallRepo.GetFirst(repo.WithByID(req.InstallID))
	if err != nil {
		return result, err
	}
	if install.ID != req.InstallID || install.App.Key != req.App || install.Version != req.Template {
		return result, buserr.New("ErrClusterRestoreMemberChanged")
	}
	root, err := os.OpenRoot(install.GetPath())
	if err != nil {
		return result, err
	}
	defer root.Close()
	if req.Action == "prepare" {
		return prepareOriginalClusterMember(ctx, root, install, req)
	}

	state, err := readOriginalRestoreState(root, originalRestoreWork)
	if errors.Is(err, os.ErrNotExist) {
		switch req.Action {
		case "rollback":
			return result, nil
		case "finish":
			dir := originalRestoreHistory(req.TaskID)
			saved, savedErr := readOriginalRestoreState(root, dir)
			if savedErr == nil && saved.TaskID == req.TaskID {
				result.PreviousData = filepath.Join(root.Name(), dir, "previous")
				return result, nil
			}
		}
	}
	if err != nil {
		return result, err
	}
	if state.TaskID != req.TaskID {
		return result, buserr.New("ErrClusterRestoreMemberBusy")
	}
	data, _ := originalRestoreMount(req.App)
	if req.Action == "finish" {
		if err := checkReplacedClusterData(root, data); err != nil {
			return result, err
		}
		dir := originalRestoreHistory(req.TaskID)
		if err := root.Rename(originalRestoreWork, dir); err != nil {
			return result, err
		}
		result.PreviousData = filepath.Join(root.Name(), dir, "previous")
		return result, nil
	}
	cli, err := docker.NewDockerClientWithContext(ctx)
	if err != nil {
		return result, err
	}
	defer cli.Close()
	switch req.Action {
	case "replace":
		if !state.Prepared || state.PublishAttempted {
			return result, buserr.New("ErrClusterRestoreMemberState")
		}
		if !state.Replacing {
			state.Replacing = true
			if err := writeOriginalRestoreState(root, originalRestoreWork, state); err != nil {
				return result, err
			}
		}
		if err := pauseOriginalRestoreContainer(ctx, cli, state.ContainerID); err != nil {
			return result, err
		}
		return result, replaceOriginalClusterData(root, data)
	case "start":
		if err := checkReplacedClusterData(root, data); err != nil {
			return result, err
		}
		state.PublishAttempted = true
		if err := writeOriginalRestoreState(root, originalRestoreWork, state); err != nil {
			return result, err
		}
		if err := startOriginalRestoreContainer(ctx, cli, state.ContainerID); err != nil {
			return result, err
		}
		if err := waitForOriginalRestoreMember(ctx, install, req); err != nil {
			return result, err
		}
		if err := setRestoreRestartPolicy(ctx, cli, state.ContainerID, state.RestartPolicy); err != nil {
			return result, err
		}
		return result, syncAppInstallStatus(&install, true)
	case "pause":
		if err := pauseOriginalRestoreContainer(ctx, cli, state.ContainerID); err != nil {
			return result, err
		}
		return result, syncAppInstallStatus(&install, true)
	case "rollback":
		if state.PublishAttempted {
			return result, buserr.New("ErrClusterRestoreRollbackRequiresReview")
		}
		if !state.Replacing {
			if !state.Prepared && req.Role == "master" {
				if err := removeRestoreValidationContainer(ctx, cli, req); err != nil {
					return result, err
				}
			}
			return result, root.RemoveAll(originalRestoreWork)
		}
		if err := rollbackOriginalClusterData(root, data); err != nil {
			return result, err
		}
		if state.WasRunning {
			if err := startOriginalRestoreContainer(ctx, cli, state.ContainerID); err != nil {
				return result, err
			}
		}
		if err := setRestoreRestartPolicy(ctx, cli, state.ContainerID, state.RestartPolicy); err != nil {
			return result, err
		}
		return result, root.RemoveAll(originalRestoreWork)
	default:
		return result, buserr.New("ErrClusterRestoreOriginalAction")
	}
}

func prepareOriginalClusterMember(ctx context.Context, root *os.Root, install model.AppInstall, req OriginalClusterRestoreRequest) (OriginalClusterRestoreResult, error) {
	var result OriginalClusterRestoreResult
	if existing, readErr := readOriginalRestoreState(root, originalRestoreWork); readErr == nil {
		if existing.TaskID == req.TaskID && existing.Prepared {
			return OriginalClusterRestoreResult{WasRunning: existing.WasRunning}, nil
		}
		return result, buserr.New("ErrClusterRestoreMemberBusy")
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return result, readErr
	}
	operation, err := taskRepo.GetExecutingTask(task.TaskScopeApp, install.ID)
	if err != nil {
		return result, err
	}
	if operation.ID != "" {
		return result, buserr.New("TaskIsExecuting")
	}
	db, err := newClusterDatabase(install)
	if err != nil {
		return result, err
	}
	cli, err := docker.NewDockerClientWithContext(ctx)
	if err != nil {
		return result, err
	}
	defer cli.Close()
	info, err := cli.ContainerInspect(ctx, install.ContainerName)
	if err != nil {
		return result, err
	}
	data, destination := originalRestoreMount(req.App)
	mounts, err := originalRestoreMounts(root, info, data, destination, req)
	if err != nil {
		return result, err
	}
	if info.State.Running {
		probe, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := checkOriginalRestoreVersion(probe, db, req)
		if err == nil {
			err = checkOriginalRestoreDataLocation(probe, db, req)
		}
		cancel()
		if err != nil {
			return result, err
		}
	}
	stat, err := root.Lstat(data)
	if err != nil {
		return result, err
	}
	if !stat.IsDir() {
		return result, buserr.New("ErrClusterRestoreDataMountDirectory")
	}
	state := originalRestoreState{
		TaskID: req.TaskID, ContainerID: info.ID, WasRunning: info.State.Running, RestartPolicy: info.HostConfig.RestartPolicy,
	}
	scratch, err := os.MkdirTemp(root.Name(), ".1panel-restore-")
	if err != nil {
		return result, err
	}
	scratch = filepath.Base(scratch)
	defer root.RemoveAll(scratch)
	if err := writeOriginalRestoreState(root, scratch, state); err != nil {
		return result, err
	}
	if err := root.Rename(scratch, originalRestoreWork); err != nil {
		return result, err
	}
	staged := filepath.Join(root.Name(), originalRestoreWork)
	if err := os.Mkdir(filepath.Join(staged, "restored"), stat.Mode().Perm()); err != nil {
		return result, err
	}
	if req.Role == "master" {
		if err := stageOriginalRestorePrimary(ctx, cli, db, info, mounts, staged, stat, req); err != nil {
			return result, err
		}
	} else if err := copyRestoreOwnership(filepath.Join(staged, "restored"), stat); err != nil {
		return result, err
	}
	state.Prepared = true
	return OriginalClusterRestoreResult{WasRunning: state.WasRunning}, writeOriginalRestoreState(root, originalRestoreWork, state)
}

func originalRestoreMount(app string) (data string, destination string) {
	switch app {
	case constant.AppMysqlCluster:
		return "data", "/bitnami/mysql/data"
	case constant.AppPostgresqlCluster:
		return "postgresql", "/bitnami/postgresql"
	default:
		return "data", "/bitnami/redis/data"
	}
}

func originalRestoreMounts(root *os.Root, info container.InspectResponse, data, destination string, req OriginalClusterRestoreRequest) ([]container.MountPoint, error) {
	if info.ContainerJSONBase == nil || info.State == nil || info.Config == nil || info.HostConfig == nil ||
		info.State.Paused || info.State.Restarting || info.State.Dead || info.Image == "" {
		return nil, buserr.New("ErrClusterBackupContainerUnavailable")
	}
	names := map[string]string{
		constant.AppMysqlCluster: "MYSQL", constant.AppPostgresqlCluster: "POSTGRESQL", constant.AppRedisCluster: "REDIS",
	}
	if !containsRestoreEnv(info.Config.Env, names[req.App]+"_REPLICATION_MODE="+req.Role) {
		return nil, buserr.New("ErrClusterRestoreMemberRoleMismatch")
	}
	base, err := filepath.EvalSymlinks(root.Name())
	if err != nil {
		return nil, err
	}
	var mounts []container.MountPoint
	hasData := false
	for _, item := range info.Mounts {
		if item.Destination == "/etc/localtime" && !item.RW {
			continue
		}
		want := ""
		switch item.Destination {
		case destination:
			want = data
		case "/opt/bitnami/mysql/conf/my.cnf":
			want = "conf/my.cnf"
		case "/bitnami/postgresql/conf":
			want = "data"
		case "/opt/bitnami/redis/mounted-etc/redis.conf":
			want = "conf/redis.conf"
		case "/opt/bitnami/redis/logs":
			want = "logs"
		}
		source, err := filepath.EvalSymlinks(item.Source)
		if err != nil || want == "" || item.Type != mount.TypeBind || source != filepath.Join(base, want) {
			return nil, buserr.New("ErrClusterRestoreOriginalCatalogMounts")
		}
		if item.Destination == destination {
			hasData = item.RW
		}
		mounts = append(mounts, item)
	}
	if !hasData {
		return nil, buserr.New("ErrClusterRestoreDataMountUnavailable")
	}
	return mounts, nil
}

func containsRestoreEnv(env []string, value string) bool {
	for _, item := range env {
		if item == value {
			return true
		}
	}
	return false
}

func stageOriginalRestorePrimary(ctx context.Context, cli *client.Client, db clusterDatabase, info container.InspectResponse, mounts []container.MountPoint, staged string, stat fs.FileInfo, req OriginalClusterRestoreRequest) error {
	backup, err := OpenClusterBackup(req.Artifact)
	if err != nil {
		return err
	}
	defer backup.Close()
	payload, err := os.CreateTemp(staged, ".payload-")
	if err != nil {
		return err
	}
	defer os.Remove(payload.Name())
	manifest, err := clusterarchive.Unpack(backup, payload, req.Passphrase)
	if err = errors.Join(err, payload.Close()); err != nil {
		return err
	}
	if manifest.App != req.App || manifest.Template != req.Template || manifest.EngineVersion != req.EngineVersion {
		return buserr.New("ErrClusterRestoreOriginalTemplateVersion")
	}
	switch req.App {
	case constant.AppRedisCluster:
		if err := validateRedisRestoreConfig(filepath.Join(db.install.GetPath(), "conf", "redis.conf")); err != nil {
			return err
		}
		if err := restoreRedisSnapshot(staged, "restored", payload.Name()); err != nil {
			return err
		}
	case constant.AppPostgresqlCluster:
		if err := restorePostgresSnapshot(staged, "restored/data", payload.Name(), manifest.EngineVersion); err != nil {
			return err
		}
	}
	if err := filepath.WalkDir(filepath.Join(staged, "restored"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return copyRestoreOwnership(path, stat)
	}); err != nil {
		return err
	}

	_, destination := originalRestoreMount(req.App)
	isolated := make([]mount.Mount, 0, len(mounts))
	var configCopies []string
	for i, item := range mounts {
		source := filepath.Join(staged, "restored")
		if item.Destination != destination {
			source = filepath.Join(staged, "mount-"+strconv.Itoa(i))
			var err error
			if item.Destination == "/opt/bitnami/redis/logs" {
				err = os.Mkdir(source, 0755)
				if err == nil {
					err = copyRestoreOwnership(source, stat)
				}
			} else {
				err = copyRestoreConfig(item.Source, source)
			}
			if err != nil {
				return err
			}
			configCopies = append(configCopies, source)
		}
		isolated = append(isolated, mount.Mount{Type: mount.TypeBind, Source: source, Target: item.Destination, ReadOnly: !item.RW})
	}
	cfg := &container.Config{
		Image: info.Image, User: info.Config.User, Env: info.Config.Env,
		Entrypoint: info.Config.Entrypoint, Cmd: info.Config.Cmd, WorkingDir: info.Config.WorkingDir,
		Labels: map[string]string{"1panel.cluster.restore": req.TaskID},
	}
	created, err := cli.ContainerCreate(ctx, cfg, &container.HostConfig{
		NetworkMode: "none", Mounts: isolated, ShmSize: info.HostConfig.ShmSize,
	}, nil, nil, restoreValidationContainerName(req))
	if err != nil {
		return err
	}
	validateErr := func() error {
		if err := cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
			return err
		}
		install := db.install
		install.ContainerName = created.ID
		if err := finishClusterRestore(ctx, &install, manifest, payload.Name()); err != nil {
			return err
		}
		db.install = install
		probe, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return checkOriginalRestoreDataLocation(probe, db, req)
	}()
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if cleanupErr := removeRestoreValidationContainer(cleanupCtx, cli, req); cleanupErr != nil {
		return errors.Join(validateErr, cleanupErr)
	}
	for _, path := range configCopies {
		validateErr = errors.Join(validateErr, os.RemoveAll(path))
	}
	return validateErr
}

func restoreValidationContainerName(req OriginalClusterRestoreRequest) string {
	return fmt.Sprintf("1panel-ha-restore-%s-%d", req.TaskID, req.InstallID)
}

func removeRestoreValidationContainer(ctx context.Context, cli *client.Client, req OriginalClusterRestoreRequest) error {
	info, err := cli.ContainerInspect(ctx, restoreValidationContainerName(req))
	if errdefs.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Config == nil || info.Config.Labels["1panel.cluster.restore"] != req.TaskID {
		return buserr.New("ErrClusterRestoreValidationContainer")
	}
	stopErr := stopOriginalRestoreContainer(ctx, cli, info.ID)
	removeErr := cli.ContainerRemove(ctx, info.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
	if errdefs.IsNotFound(removeErr) {
		removeErr = nil
	}
	return errors.Join(stopErr, removeErr)
}

func copyRestoreOwnership(path string, info fs.FileInfo) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return buserr.New("ErrClusterRestoreDataOwnership")
	}
	return os.Lchown(path, int(stat.Uid), int(stat.Gid))
}

func copyRestoreConfig(source, dest string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, relative)
		switch {
		case info.IsDir():
			err = os.Mkdir(target, info.Mode().Perm())
		case info.Mode().IsRegular():
			err = copyRestoreConfigFile(path, target, info.Mode().Perm())
		default:
			return buserr.New("ErrClusterRestoreConfigFileType")
		}
		if err != nil {
			return err
		}
		return copyRestoreOwnership(target, info)
	})
}

func copyRestoreConfigFile(source, dest string, mode fs.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	return errors.Join(err, out.Close())
}

func checkOriginalRestoreVersion(ctx context.Context, db clusterDatabase, req OriginalClusterRestoreRequest) error {
	var version string
	var err error
	switch req.App {
	case constant.AppMysqlCluster:
		version, err = db.mysql(ctx, "SELECT VERSION()")
	case constant.AppPostgresqlCluster:
		version, err = db.postgres(ctx, "SHOW server_version")
	case constant.AppRedisCluster:
		version, err = db.redis(ctx, "INFO", "server")
		version = redisEngineVersion(version)
	}
	if err != nil {
		return err
	}
	if version != req.EngineVersion {
		return buserr.WithMap("ErrClusterRestoreEngineVersion", map[string]interface{}{
			"current": version, "expected": req.EngineVersion,
		}, nil)
	}
	return nil
}

func checkOriginalRestoreDataLocation(ctx context.Context, db clusterDatabase, req OriginalClusterRestoreRequest) error {
	_, expected := originalRestoreMount(req.App)
	var directory string
	var err error
	switch req.App {
	case constant.AppMysqlCluster:
		directory, err = db.mysql(ctx, "SELECT @@datadir")
	case constant.AppPostgresqlCluster:
		directory, err = db.postgres(ctx, "SHOW data_directory")
		expected += "/data"
	case constant.AppRedisCluster:
		var config string
		config, err = db.redis(ctx, "CONFIG", "GET", "dir")
		directory = strings.TrimPrefix(config, "dir\n")
		if err == nil && req.Role == "master" {
			for _, setting := range []struct{ key, value string }{{"appendonly", "no"}, {"dbfilename", "dump.rdb"}} {
				output, err := db.redis(ctx, "CONFIG", "GET", setting.key)
				if err != nil {
					return err
				}
				if output != setting.key+"\n"+setting.value {
					return buserr.New("ErrClusterRestoreRedisActiveRDB")
				}
			}
		}
	}
	if err != nil {
		return err
	}
	if filepath.Clean(directory) != expected {
		return buserr.New("ErrClusterRestoreDataDirectoryMismatch")
	}
	return nil
}

func waitForOriginalRestoreMember(ctx context.Context, install model.AppInstall, req OriginalClusterRestoreRequest) error {
	db, err := newClusterDatabase(install)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if req.Role == "master" {
		if err := waitForClusterRestore(ctx, db, clusterarchive.Manifest{App: req.App, EngineVersion: req.EngineVersion}); err != nil {
			return err
		}
		return checkOriginalRestoreDataLocation(ctx, db, req)
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		err := checkOriginalRestoreVersion(ctx, db, req)
		if err == nil {
			return checkOriginalRestoreDataLocation(ctx, db, req)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", buserr.New("ErrClusterRestoreReplicaNotReady"), err)
		case <-ticker.C:
		}
	}
}

func stopOriginalRestoreContainer(ctx context.Context, cli *client.Client, containerID string) error {
	info, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	if !info.State.Running && !info.State.Restarting {
		return nil
	}
	timeout := 120
	return cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

func setRestoreRestartPolicy(ctx context.Context, cli *client.Client, containerID string, policy container.RestartPolicy) error {
	if policy.Name == "" {
		policy.Name = "no"
	}
	_, err := cli.ContainerUpdate(ctx, containerID, container.UpdateConfig{RestartPolicy: policy})
	return err
}

func pauseOriginalRestoreContainer(ctx context.Context, cli *client.Client, containerID string) error {
	if err := setRestoreRestartPolicy(ctx, cli, containerID, container.RestartPolicy{Name: "no"}); err != nil {
		return err
	}
	return stopOriginalRestoreContainer(ctx, cli, containerID)
}

func startOriginalRestoreContainer(ctx context.Context, cli *client.Client, containerID string) error {
	info, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	if info.State.Running {
		return nil
	}
	return cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

func originalRestoreHistory(taskID string) string {
	return ".1panel-pre-restore-" + taskID
}

func readOriginalRestoreState(root *os.Root, dir string) (originalRestoreState, error) {
	var state originalRestoreState
	raw, err := root.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return state, err
	}
	return state, nil
}

func writeOriginalRestoreState(root *os.Root, dir string, state originalRestoreState) error {
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, "state.tmp")
	if err := root.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return root.Rename(tmp, filepath.Join(dir, "state.json"))
}

func restoreDirectoryExists(root *os.Root, name string) (bool, error) {
	stat, err := root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !stat.IsDir() {
		return false, buserr.New("ErrClusterRestoreDataPathDirectory")
	}
	return true, nil
}

func replaceOriginalClusterData(root *os.Root, data string) error {
	previous := filepath.Join(originalRestoreWork, "previous")
	restored := filepath.Join(originalRestoreWork, "restored")
	hadPrevious, err := restoreDirectoryExists(root, previous)
	if err != nil {
		return err
	}
	hasRestored, err := restoreDirectoryExists(root, restored)
	if err != nil {
		return err
	}
	if !hasRestored {
		return checkReplacedClusterData(root, data)
	}
	hasLive, err := restoreDirectoryExists(root, data)
	if err != nil {
		return err
	}
	if !hadPrevious {
		if !hasLive {
			return buserr.New("ErrClusterRestoreOriginalDataUnavailable")
		}
		if err := root.Rename(data, previous); err != nil {
			return err
		}
	} else if hasLive {
		return buserr.New("ErrClusterRestoreDataDirectoryChanged")
	}
	return root.Rename(restored, data)
}

func checkReplacedClusterData(root *os.Root, data string) error {
	for _, dir := range []string{data, filepath.Join(originalRestoreWork, "previous")} {
		exists, err := restoreDirectoryExists(root, dir)
		if err != nil {
			return err
		}
		if !exists {
			return buserr.New("ErrClusterRestoreDataReplacementIncomplete")
		}
	}
	staged, err := restoreDirectoryExists(root, filepath.Join(originalRestoreWork, "restored"))
	if err != nil {
		return err
	}
	if staged {
		return buserr.New("ErrClusterRestoreDataReplacementIncomplete")
	}
	return nil
}

func rollbackOriginalClusterData(root *os.Root, data string) error {
	previous := filepath.Join(originalRestoreWork, "previous")
	exists, err := restoreDirectoryExists(root, previous)
	if err != nil || !exists {
		return err
	}
	restored := filepath.Join(originalRestoreWork, "restored")
	live, err := restoreDirectoryExists(root, data)
	if err != nil {
		return err
	}
	if live {
		if staged, err := restoreDirectoryExists(root, restored); err != nil {
			return err
		} else if staged {
			return buserr.New("ErrClusterRestoreRollbackDirectoryChanged")
		}
		if err := root.Rename(data, restored); err != nil {
			return err
		}
	}
	return root.Rename(previous, data)
}
