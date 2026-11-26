package service

import (
	"archive/tar"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/clusterarchive"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
)

func RestoreClusterBackup(ctx context.Context, req request.AppInstallCreate, id, passphrase string) (*model.AppInstall, error) {
	if !filepath.IsLocal(req.Name) || filepath.Base(req.Name) != req.Name || req.EditCompose {
		return nil, buserr.New("ErrClusterRestoreCatalogInstance")
	}
	source, err := OpenClusterBackup(id)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	payload, err := os.CreateTemp(clusterBackupRoot(), ".restore-")
	if err != nil {
		return nil, err
	}
	name := payload.Name()
	submitted := false
	defer func() {
		if !submitted {
			_ = os.Remove(name)
		}
	}()
	manifest, unpackErr := clusterarchive.Unpack(source, payload, passphrase)
	if err := errors.Join(unpackErr, payload.Close()); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	detail, err := appDetailRepo.GetFirst(repo.WithByID(req.AppDetailId))
	if err != nil {
		return nil, err
	}
	app, err := appRepo.GetFirst(repo.WithByID(detail.AppId))
	if err != nil {
		return nil, err
	}
	if manifest.App != app.Key || manifest.Template != detail.Version {
		return nil, buserr.New("ErrClusterRestorePrimaryTemplate")
	}

	install, err := (AppService{}).installWithHooks(req, true, &AppInstallHooks{
		AfterCopyData: func(install *model.AppInstall) error {
			return prepareClusterRestore(install, manifest, name)
		},
		AfterStart: func(install *model.AppInstall, t *task.Task) error {
			if err := finishClusterRestore(t.TaskCtx, install, manifest, name); err != nil {
				_, stopErr := compose.Stop(install.GetComposePath())
				return errors.Join(err, stopErr)
			}
			return nil
		},
		AfterFinish: func() { _ = os.Remove(name) },
	}, true)
	submitted = err == nil
	return install, err
}

func prepareClusterRestore(install *model.AppInstall, manifest clusterarchive.Manifest, payload string) error {
	switch manifest.App {
	case constant.AppMysqlCluster:
		return nil
	case constant.AppRedisCluster:
		if err := validateRestoreBind(install, "data", "/bitnami/redis/data"); err != nil {
			return err
		}
		if err := validateRedisRestoreConfig(filepath.Join(install.GetPath(), "conf", "redis.conf")); err != nil {
			return err
		}
		return restoreRedisSnapshot(install.GetPath(), "data", payload)
	case constant.AppPostgresqlCluster:
		if err := validateRestoreBind(install, "postgresql", "/bitnami/postgresql"); err != nil {
			return err
		}
		return restorePostgresSnapshot(install.GetPath(), "postgresql/data", payload, manifest.EngineVersion)
	default:
		return buserr.New("ErrClusterRestoreUnsupportedEngine")
	}
}

func restoreRedisSnapshot(directory, dataPath, payload string) error {
	file, err := os.Open(payload)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	magic, err := reader.Peek(5)
	if err != nil || string(magic) != "REDIS" {
		return buserr.New("ErrClusterRestoreRedisSnapshot")
	}
	root, err := openEmptyRestoreDir(directory, dataPath)
	if err != nil {
		return err
	}
	defer root.Close()
	return writeRestoreFile(root, "dump.rdb", reader)
}

func restorePostgresSnapshot(directory, dataPath, payload, engineVersion string) error {
	root, err := openEmptyRestoreDir(directory, dataPath)
	if err != nil {
		return err
	}
	defer root.Close()
	file, err := os.Open(payload)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := extractPostgresBackup(root, file); err != nil {
		return err
	}
	major, _, _ := strings.Cut(engineVersion, ".")
	version, err := root.ReadFile("PG_VERSION")
	if err != nil || strings.TrimSpace(string(version)) != major {
		return buserr.New("ErrClusterRestorePostgresMajorVersion")
	}
	return nil
}

func validateRestoreBind(install *model.AppInstall, source, destination string) error {
	env, err := os.ReadFile(install.GetEnvPath())
	if err != nil {
		return err
	}
	project, err := docker.GetComposeProject(install.Name, install.GetPath(), []byte(install.DockerCompose), env, false)
	if err != nil {
		return err
	}
	if len(project.Services) != 1 {
		return buserr.New("ErrClusterRestoreSingleContainer")
	}
	for _, service := range project.Services {
		for _, mount := range service.Volumes {
			if mount.Type == "bind" && !mount.ReadOnly && mount.Target == destination &&
				filepath.Clean(mount.Source) == filepath.Join(install.GetPath(), source) {
				return nil
			}
		}
	}
	return buserr.New("ErrClusterRestoreCatalogDataMount")
}

func openEmptyRestoreDir(directory, name string) (*os.Root, error) {
	app, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer app.Close()
	if err := app.MkdirAll(name, 0700); err != nil {
		return nil, err
	}
	dir, err := app.Open(name)
	if err != nil {
		return nil, err
	}
	_, err = dir.Readdirnames(1)
	_ = dir.Close()
	if err != io.EOF {
		return nil, buserr.New("ErrClusterRestoreDataDirectoryNotEmpty")
	}
	return app.OpenRoot(name)
}

func writeRestoreFile(root *os.Root, name string, src io.Reader) error {
	if err := root.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	out, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, src)
	return errors.Join(copyErr, out.Close())
}

func extractPostgresBackup(root *os.Root, source io.Reader) error {
	reader := tar.NewReader(source)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.Clean(header.Name)
		if !filepath.IsLocal(header.Name) || !filepath.IsLocal(name) {
			return buserr.New("ErrClusterRestorePostgresUnsafePath")
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			return buserr.New("ErrClusterRestorePostgresFileType")
		}
		switch name {
		case "postgresql.auto.conf", "postgresql.conf", "pg_hba.conf", "pg_ident.conf",
			"recovery.signal", "standby.signal", "postmaster.opts", "postmaster.pid":
			continue
		}
		if strings.HasPrefix(name, "pg_replslot/") {
			continue
		}
		if header.Typeflag == tar.TypeDir {
			if err := root.MkdirAll(name, 0700); err != nil {
				return err
			}
		} else if err := writeRestoreFile(root, name, reader); err != nil {
			return err
		}
	}
}

func validateRedisRestoreConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	settings := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if len(line) < 2 || strings.HasPrefix(line[0], "#") {
			continue
		}
		if line[0] == "include" {
			return buserr.New("ErrClusterRestoreRedisCatalogPersistence")
		}
		settings[line[0]] = strings.Trim(line[1], "\"'")
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if settings["appendonly"] != "no" || settings["dir"] != "/bitnami/redis/data" || settings["dbfilename"] != "dump.rdb" {
		return buserr.New("ErrClusterRestoreRedisCatalogRDB")
	}
	return nil
}

func finishClusterRestore(ctx context.Context, install *model.AppInstall, manifest clusterarchive.Manifest, payload string) error {
	db, err := newClusterDatabase(*install)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()
	if err := waitForClusterRestore(ctx, db, manifest); err != nil {
		return err
	}
	if manifest.App != constant.AppMysqlCluster {
		return nil
	}
	file, err := os.Open(payload)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := db.exec(ctx, io.Discard, file, "MYSQL_PWD", db.env["PANEL_DB_ROOT_PASSWORD"],
		"mysql", "--host=127.0.0.1", "--user=root", "--binary-mode=1"); err != nil {
		return err
	}
	if _, err := db.mysql(ctx, "SELECT 1"); err != nil {
		return buserr.New("ErrClusterRestoreMySQLCredentials")
	}
	return nil
}

func waitForClusterRestore(ctx context.Context, db clusterDatabase, manifest clusterarchive.Manifest) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	probe, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	for {
		var version string
		var err error
		switch manifest.App {
		case constant.AppMysqlCluster:
			version, err = db.mysql(probe, "SELECT VERSION()")
		case constant.AppPostgresqlCluster:
			version, err = db.postgres(probe, "SELECT current_setting('server_version') WHERE NOT pg_is_in_recovery()")
		case constant.AppRedisCluster:
			var info string
			info, err = db.redis(probe, "INFO", "server")
			version = redisEngineVersion(info)
		}
		if err == nil && version != "" {
			if version != manifest.EngineVersion {
				return buserr.WithMap("ErrClusterRestoreEngineVersion", map[string]interface{}{
					"current":  version,
					"expected": manifest.EngineVersion,
				}, nil)
			}
			return nil
		}
		select {
		case <-probe.Done():
			return fmt.Errorf("%s: %w", buserr.New("ErrClusterRestoreDatabaseNotReady"), probe.Err())
		case <-ticker.C:
		}
	}
}
