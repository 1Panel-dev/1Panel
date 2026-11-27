package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/clusterarchive"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

type ClusterBackupArtifact struct {
	ID       string                  `json:"id"`
	Manifest clusterarchive.Manifest `json:"manifest"`
}

func clusterBackupRoot() string {
	return filepath.Join(global.Dir.LocalBackupDir, "application-ha")
}

func clusterBackupPath(id string) (string, error) {
	parsed, err := uuid.Parse(id)
	if err != nil || parsed.String() != id || parsed == uuid.Nil {
		return "", buserr.New("ErrClusterBackupID")
	}
	return filepath.Join(clusterBackupRoot(), id+".1pha"), nil
}

func OpenClusterBackup(id string) (*os.File, error) {
	path, err := clusterBackupPath(id)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func saveClusterBackup(ctx context.Context, write func(io.Writer) (clusterarchive.Manifest, error)) (ClusterBackupArtifact, error) {
	if err := os.MkdirAll(clusterBackupRoot(), 0700); err != nil {
		return ClusterBackupArtifact{}, err
	}
	file, err := os.CreateTemp(clusterBackupRoot(), ".backup-")
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	manifest, err := write(file)
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	if err := ctx.Err(); err != nil {
		return ClusterBackupArtifact{}, err
	}
	if err := file.Sync(); err != nil {
		return ClusterBackupArtifact{}, err
	}
	if err := file.Close(); err != nil {
		return ClusterBackupArtifact{}, err
	}
	id := uuid.NewString()
	dest, err := clusterBackupPath(id)
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	if err := os.Rename(file.Name(), dest); err != nil {
		return ClusterBackupArtifact{}, err
	}
	return ClusterBackupArtifact{ID: id, Manifest: manifest}, nil
}

func ImportClusterBackup(ctx context.Context, archive io.Reader, password, app string) (ClusterBackupArtifact, error) {
	return saveClusterBackup(ctx, func(out io.Writer) (clusterarchive.Manifest, error) {
		manifest, err := clusterarchive.Unpack(io.TeeReader(archive, out), io.Discard, password)
		if err != nil {
			return clusterarchive.Manifest{}, err
		}
		if manifest.App != app {
			return clusterarchive.Manifest{}, buserr.New("ErrClusterBackupEngineMismatch")
		}
		return manifest, nil
	})
}

func ImportClusterBackupFile(ctx context.Context, path, password, app string) (ClusterBackupArtifact, error) {
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".1pha") {
		return ClusterBackupArtifact{}, buserr.New("ErrClusterBackupSelectFile")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	if files.ShouldFilterSensitivePath(path) || files.ShouldFilterSensitivePath(resolved) || files.ShouldDenySensitiveFileRead(path) {
		return ClusterBackupArtifact{}, buserr.New("ErrClusterBackupFileInaccessible")
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	if !info.Mode().IsRegular() {
		return ClusterBackupArtifact{}, buserr.New("ErrClusterBackupRegularFile")
	}
	file, err := os.OpenFile(resolved, os.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return ClusterBackupArtifact{}, buserr.New("ErrClusterBackupFileChanged")
	}
	return ImportClusterBackup(ctx, file, password, app)
}

func CreateClusterBackup(ctx context.Context, installID uint, password string) (ClusterBackupArtifact, error) {
	if password == "" {
		return ClusterBackupArtifact{}, clusterarchive.ErrPassphraseRequired
	}
	install, err := appInstallRepo.GetFirst(repo.WithByID(installID))
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	if install.ID != installID || !constant.IsClusterApp(install.App.Key) || !strings.HasSuffix(install.Version, "-master") {
		return ClusterBackupArtifact{}, buserr.New("ErrClusterBackupNotPrimary")
	}
	source, err := newClusterDatabase(install)
	if err != nil {
		return ClusterBackupArtifact{}, err
	}
	return saveClusterBackup(ctx, func(out io.Writer) (clusterarchive.Manifest, error) {
		payload, err := os.CreateTemp(clusterBackupRoot(), ".payload-")
		if err != nil {
			return clusterarchive.Manifest{}, err
		}
		defer os.Remove(payload.Name())
		defer payload.Close()
		engine, format, err := source.capture(ctx, payload)
		if err != nil {
			return clusterarchive.Manifest{}, err
		}
		if err := payload.Close(); err != nil {
			return clusterarchive.Manifest{}, err
		}
		return clusterarchive.Pack(out, payload.Name(), clusterarchive.Manifest{
			App: install.App.Key, Template: install.Version, EngineVersion: engine, Format: format,
			CreatedAt: time.Now().UTC(),
		}, password)
	})
}

type clusterDatabase struct {
	install    model.AppInstall
	env        map[string]string
	dockerHost string
}

func newClusterDatabase(install model.AppInstall) (clusterDatabase, error) {
	name := install.ContainerName
	if name == "" || strings.ContainsAny(name, ", \r\n\t") || strings.HasPrefix(name, "-") {
		return clusterDatabase{}, buserr.New("ErrClusterBackupContainerUnavailable")
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(install.Env), &raw); err != nil {
		return clusterDatabase{}, err
	}
	values := make(map[string]string)
	for key, value := range raw {
		if s, ok := value.(string); ok {
			values[key] = s
		}
	}
	return clusterDatabase{install: install, env: values}, nil
}

func (d clusterDatabase) exec(ctx context.Context, out io.Writer, input io.Reader, credential, secret string, args ...string) error {
	dockerArgs := []string{"exec", "-i"}
	if credential != "" {
		dockerArgs = append(dockerArgs, "-e", credential)
	}
	dockerArgs = append(dockerArgs, d.install.ContainerName)
	dockerArgs = append(dockerArgs, args...)
	command := exec.CommandContext(ctx, "docker", dockerArgs...)
	if credential != "" {
		command.Env = append(command.Environ(), credential+"="+secret)
	}
	if d.dockerHost != "" {
		command.Env = append(command.Environ(), "DOCKER_HOST="+d.dockerHost, "DOCKER_CONTEXT=")
	}
	command.Stdout, command.Stdin = out, input
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s: %w", buserr.WithName("ErrClusterCommandFailed", args[0]), err)
	}
	return nil
}

func (d clusterDatabase) result(ctx context.Context, credential, secret string, args ...string) (string, error) {
	var output bytes.Buffer
	err := d.exec(ctx, &output, nil, credential, secret, args...)
	return strings.TrimSpace(output.String()), err
}

func (d clusterDatabase) mysql(ctx context.Context, query string) (string, error) {
	return d.result(ctx, "MYSQL_PWD", d.env["PANEL_DB_ROOT_PASSWORD"],
		"mysql", "--host=127.0.0.1", "--user=root", "--batch", "--skip-column-names", "--execute", query)
}

func (d clusterDatabase) postgres(ctx context.Context, query string) (string, error) {
	return d.result(ctx, "PGPASSWORD", d.env["PANEL_DB_ROOT_PASSWORD"],
		"psql", "--no-password", "--host=127.0.0.1", "--username=postgres", "--dbname=postgres", "--no-psqlrc", "--tuples-only", "--no-align", "--command", query)
}

func (d clusterDatabase) redis(ctx context.Context, args ...string) (string, error) {
	return d.result(ctx, "REDISCLI_AUTH", d.env["PANEL_REDIS_ROOT_PASSWORD"], append([]string{"redis-cli", "--raw"}, args...)...)
}

func (d clusterDatabase) capture(ctx context.Context, out io.Writer) (string, string, error) {
	switch d.install.App.Key {
	case constant.AppMysqlCluster:
		version, err := d.mysql(ctx, "SELECT VERSION()")
		if err != nil {
			return "", "", err
		}
		nonInnoDB, err := d.mysql(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_SCHEMA NOT IN ('mysql','information_schema','performance_schema','sys') AND ENGINE <> 'InnoDB'")
		if err != nil {
			return "", "", err
		}
		if nonInnoDB != "0" {
			return "", "", buserr.New("ErrClusterBackupMySQLInnoDB")
		}
		err = d.exec(ctx, out, nil, "MYSQL_PWD", d.env["PANEL_DB_ROOT_PASSWORD"],
			"mysqldump", "--host=127.0.0.1", "--user=root", "--single-transaction", "--all-databases",
			"--no-tablespaces", "--set-gtid-purged=OFF", "--routines", "--events", "--triggers", "--flush-privileges")
		return version, "sql", err
	case constant.AppPostgresqlCluster:
		version, err := d.postgres(ctx, "SHOW server_version")
		if err != nil {
			return "", "", err
		}
		role, err := d.postgres(ctx, "SELECT pg_is_in_recovery()")
		if err != nil {
			return "", "", err
		}
		if role != "f" {
			return "", "", buserr.New("ErrClusterBackupPostgresRecovery")
		}
		spaces, err := d.postgres(ctx, "SELECT COUNT(*) FROM pg_tablespace WHERE spcname NOT IN ('pg_default','pg_global')")
		if err != nil {
			return "", "", err
		}
		if spaces != "0" {
			return "", "", buserr.New("ErrClusterBackupPostgresTablespaces")
		}
		user := d.env["REPLICATION_USER"]
		if user == "" {
			return "", "", buserr.New("ErrClusterBackupPostgresReplicationCredentials")
		}
		err = d.exec(ctx, out, nil, "PGPASSWORD", d.env["REPLICATION_PASSWORD"],
			"pg_basebackup", "--host=127.0.0.1", "--username="+user, "--no-password", "--format=tar",
			"--pgdata=-", "--wal-method=fetch", "--no-manifest")
		return version, "basebackup", err
	case constant.AppRedisCluster:
		info, err := d.redis(ctx, "INFO", "server")
		if err != nil {
			return "", "", err
		}
		role, err := d.redis(ctx, "ROLE")
		if err != nil {
			return "", "", err
		}
		if line, _, _ := strings.Cut(role, "\n"); strings.TrimSpace(line) != "master" {
			return "", "", buserr.New("ErrClusterBackupRedisNotPrimary")
		}
		version := redisEngineVersion(info)
		if version == "" {
			return "", "", buserr.New("ErrClusterBackupRedisVersion")
		}
		temp := "/tmp/1panel-ha-" + uuid.NewString() + ".rdb"
		defer func() {
			clean, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_ = d.exec(clean, io.Discard, nil, "", "", "rm", "-f", temp)
		}()
		if err := d.exec(ctx, io.Discard, nil, "REDISCLI_AUTH", d.env["PANEL_REDIS_ROOT_PASSWORD"], "redis-cli", "--rdb", temp); err != nil {
			return "", "", err
		}
		err = d.exec(ctx, out, nil, "", "", "cat", temp)
		return version, "rdb", err
	default:
		return "", "", buserr.New("ErrClusterBackupUnsupportedEngine")
	}
}

func redisEngineVersion(info string) string {
	for _, line := range strings.Split(info, "\n") {
		if version, ok := strings.CutPrefix(line, "redis_version:"); ok {
			return strings.TrimSpace(version)
		}
	}
	return ""
}

func VerifyClusterBackup(ctx context.Context, artifact ClusterBackupArtifact, passphrase string) error {
	f, err := OpenClusterBackup(artifact.ID)
	if err != nil {
		return err
	}
	defer f.Close()
	manifest, err := clusterarchive.Unpack(f, io.Discard, passphrase)
	if err != nil {
		return err
	}
	if manifest.Digest != artifact.Manifest.Digest {
		return clusterarchive.ErrChecksum
	}
	return ctx.Err()
}
