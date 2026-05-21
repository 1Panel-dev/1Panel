package client

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/docker/docker/api/types/image"
	"github.com/pkg/errors"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const maxPgDumpStderrCapture = 64 * 1024

var pgDumpMagic = []byte("PGDMP")

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit > 0 && b.buf.Len() >= b.limit {
		b.truncated += len(p)
		return len(p), nil
	}
	if b.limit > 0 && b.buf.Len()+len(p) > b.limit {
		keep := b.limit - b.buf.Len()
		_, _ = b.buf.Write(p[:keep])
		b.truncated += len(p) - keep
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *limitedBuffer) String() string {
	if b.truncated == 0 {
		return b.buf.String()
	}
	return fmt.Sprintf("%s\n... truncated %d bytes ...", b.buf.String(), b.truncated)
}

type Remote struct {
	Client   *sql.DB
	From     string
	Database string
	User     string
	Password string
	Address  string
	Port     uint
}

func NewRemote(db Remote) *Remote {
	return &db
}
func (r *Remote) Create(info CreateInfo) error {
	createSql := fmt.Sprintf("CREATE DATABASE \"%s\"", info.Name)
	if err := r.ExecSQL(createSql, info.Timeout); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return buserr.New("ErrDatabaseIsExist")
		}
		return err
	}
	if err := r.CreateUser(info, true); err != nil {
		return err
	}
	return nil
}

func (r *Remote) CreateUser(info CreateInfo, withDeleteDB bool) error {
	createSql := fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD '%s'", info.Username, info.Password)
	if err := r.ExecSQL(createSql, info.Timeout); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return buserr.New("ErrUserIsExist")
		}
		if withDeleteDB {
			_ = r.Delete(DeleteInfo{
				Name:        info.Name,
				Username:    info.Username,
				ForceDelete: true,
				Timeout:     300})
		}
		return err
	}
	if info.SuperUser {
		if err := r.ChangePrivileges(Privileges{SuperUser: true, Username: info.Username, Timeout: info.Timeout}); err != nil {
			if withDeleteDB {
				_ = r.Delete(DeleteInfo{
					Name:        info.Name,
					Username:    info.Username,
					ForceDelete: true,
					Timeout:     300})
			}
			return err
		}
	}
	grantSql := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE \"%s\" TO \"%s\"", info.Name, info.Username)
	if err := r.ExecSQL(grantSql, info.Timeout); err != nil {
		if withDeleteDB {
			_ = r.Delete(DeleteInfo{
				Name:        info.Name,
				Username:    info.Username,
				ForceDelete: true,
				Timeout:     300})
		}
		return err
	}

	return nil
}

func (r *Remote) Delete(info DeleteInfo) error {
	if len(info.Name) != 0 {
		inUse, err := r.isDatabaseInUse(info.Name, info.Timeout)
		if err != nil && !info.ForceDelete {
			return fmt.Errorf("check database connections failed, err: %v", err)
		}
		if inUse && !info.ForceDelete {
			return buserr.WithDetail("ErrInUsed", info.Name, nil)
		}
		dropSql := fmt.Sprintf("DROP DATABASE \"%s\"", info.Name)
		if err := r.ExecSQL(dropSql, info.Timeout); err != nil && !info.ForceDelete {
			return fmt.Errorf("drop database failed, err: %v", err)
		}
	}
	dropSql := fmt.Sprintf("DROP USER \"%s\"", info.Username)
	if err := r.ExecSQL(dropSql, info.Timeout); err != nil && !info.ForceDelete {
		return fmt.Errorf("drop user failed, err: %v", err)
	}
	return nil
}

func (r *Remote) isDatabaseInUse(name string, timeout uint) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var count int
	if err := r.Client.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()",
		name,
	).Scan(&count); err != nil {
		return false, err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return false, buserr.New("ErrExecTimeOut")
	}
	return count > 0, nil
}

func (r *Remote) ChangePrivileges(info Privileges) error {
	super := "SUPERUSER"
	if !info.SuperUser {
		super = "NOSUPERUSER"
	}
	return r.ExecSQL(fmt.Sprintf("ALTER USER \"%s\" WITH %s", info.Username, super), info.Timeout)
}

func (r *Remote) ChangePassword(info PasswordChangeInfo) error {
	return r.ExecSQL(fmt.Sprintf("ALTER USER \"%s\" WITH ENCRYPTED PASSWORD '%s'", info.Username, info.Password), info.Timeout)
}

func (r *Remote) Backup(info BackupInfo) error {
	if cmd.CheckIllegal(r.Password, r.Address, r.User, info.Name) {
		return buserr.New("ErrCmdIllegal")
	}
	imageTag, err := loadImageTag(info.Database)
	if err != nil {
		return err
	}
	info.Task.Log(i18n.GetWithName("RemoteBackup", imageTag))
	fileOp := files.NewFileOp()
	if !fileOp.Stat(info.TargetDir) {
		if err := os.MkdirAll(info.TargetDir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", info.TargetDir, err)
		}
	}
	fileNameItem := info.TargetDir + "/" + strings.TrimSuffix(info.FileName, ".gz")
	backupFile, err := os.OpenFile(fileNameItem, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	backupFileClosed := false
	defer func() {
		if !backupFileClosed {
			_ = backupFile.Close()
		}
	}()
	backupCommand := exec.Command(
		"docker",
		"run", "--rm", "--net=host", "-i",
		"-e", "PGPASSWORD="+r.Password,
		imageTag,
		"pg_dump",
		"-h", r.Address,
		"-p", fmt.Sprintf("%d", r.Port),
		"--no-owner",
		"-Fc",
		"-U", r.User,
		info.Name,
	)
	backupCommand.Stdout = backupFile
	stderr := &limitedBuffer{limit: maxPgDumpStderrCapture}
	backupCommand.Stderr = stderr
	if err := backupCommand.Run(); err != nil {
		return fmt.Errorf("backup failed, stderr: %s, err: %v", strings.TrimSpace(stderr.String()), err)
	}
	if err := backupFile.Close(); err != nil {
		return fmt.Errorf("close backup file failed, err: %v", err)
	}
	backupFileClosed = true

	b := make([]byte, len(pgDumpMagic))
	handle, err := os.OpenFile(fileNameItem, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("backup file not found,err:%v", err)
	}
	defer handle.Close()
	if _, err := io.ReadFull(handle, b); err != nil {
		return fmt.Errorf("read backup header failed, stderr: %s, err: %v", strings.TrimSpace(stderr.String()), err)
	}
	if !bytes.Equal(b, pgDumpMagic) {
		return fmt.Errorf("backup failed, invalid pg dump header: %q, stderr: %s", string(b), strings.TrimSpace(stderr.String()))
	}

	gzipCmd := exec.Command("gzip", fileNameItem)
	stdout, err := gzipCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gzip file %s failed, stdout: %v, err: %v", strings.TrimSuffix(info.FileName, ".gz"), string(stdout), err)
	}
	return nil
}

func (r *Remote) Recover(info RecoverInfo) error {
	if cmd.CheckIllegal(r.Password, r.Address, r.User, info.Name, info.Username) {
		return buserr.New("ErrCmdIllegal")
	}
	imageTag, err := loadImageTag(info.Database)
	if err != nil {
		return err
	}
	info.Task.Log(i18n.GetWithName("RemoteRecover", imageTag))
	fileName := info.SourceFile
	if strings.HasSuffix(info.SourceFile, ".sql.gz") {
		fileName = strings.TrimSuffix(info.SourceFile, ".gz")
		gzipCmd := exec.Command("gunzip", info.SourceFile)
		stdout, err := gzipCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("gunzip file %s failed, stdout: %v, err: %v", info.SourceFile, string(stdout), err)
		}
		defer func() {
			gzipCmd := exec.Command("gzip", fileName)
			_, _ = gzipCmd.CombinedOutput()
		}()
	}
	restoreFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer restoreFile.Close()
	recoverCommand := exec.Command(
		"docker",
		"run", "--rm", "--net=host", "-i",
		"-e", "PGPASSWORD="+r.Password,
		imageTag,
		"pg_restore",
		"-h", r.Address,
		"-p", fmt.Sprintf("%d", r.Port),
		"--verbose",
		"--clean",
		"--no-privileges",
		"--no-owner",
		"-Fc",
		"-c",
		"--if-exists",
		"--no-owner",
		"-U", r.User,
		"-d", info.Name,
		"--role="+info.Username,
	)
	recoverCommand.Stdin = restoreFile
	pipe, _ := recoverCommand.StdoutPipe()
	stderrPipe, _ := recoverCommand.StderrPipe()
	defer pipe.Close()
	defer stderrPipe.Close()
	if err := recoverCommand.Start(); err != nil {
		return err
	}
	reader := bufio.NewReader(pipe)
	for {
		readString, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			all, _ := io.ReadAll(stderrPipe)
			global.LOG.Errorf("[PostgreSQL] DB:[%s] Recover Error: %s", info.Name, string(all))
			return err
		}
		global.LOG.Infof("[PostgreSQL] DB:[%s] Restoring: %s", info.Name, readString)
	}
	if err := recoverCommand.Wait(); err != nil {
		all, _ := io.ReadAll(stderrPipe)
		global.LOG.Errorf("[PostgreSQL] DB:[%s] Recover Error: %s", info.Name, string(all))
		return err
	}

	return nil
}

func (r *Remote) SyncDB() ([]SyncDBInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	var datas []SyncDBInfo
	rows, err := r.Client.Query("SELECT datname FROM pg_database;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		if len(dbName) == 0 || dbName == "template1" || dbName == "template0" || dbName == r.User {
			continue
		}
		datas = append(datas, SyncDBInfo{Name: dbName, From: r.From, PostgresqlName: r.Database})
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, buserr.New("ErrExecTimeOut")
	}
	return datas, nil
}

func (r *Remote) Close() {
	_ = r.Client.Close()
}

func (r *Remote) ExecSQL(command string, timeout uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	if _, err := r.Client.ExecContext(ctx, command); err != nil {
		return err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return buserr.New("ErrExecTimeOut")
	}

	return nil
}

func loadImageTag(database string) (string, error) {
	var db model.Database
	if err := global.DB.Model(&model.Database{}).Where("name = ?", database).First(&db).Error; err != nil {
		return "", fmt.Errorf("load database %s info failed, err: %v", database, err)
	}

	client, err := docker.NewDockerClient()
	if err != nil {
		return "", fmt.Errorf("create docker client failed, err: %v", err)
	}
	defer client.Close()
	images, _ := client.ImageList(context.Background(), image.ListOptions{})
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if strings.HasPrefix(tag, "postgres:"+strings.TrimSuffix(db.Version, "x")) {
				return tag, nil
			}
		}
	}
	return "postgres:" + strings.ReplaceAll(db.Version, ".x", "-alpine"), nil
}
