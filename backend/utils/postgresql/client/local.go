package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type Local struct {
	PrefixCommand []string
	Database      string
	Username      string
	Password      string
	ContainerName string
}

func NewLocal(command []string, containerName, username, password, database string) *Local {
	return &Local{PrefixCommand: command, ContainerName: containerName, Username: username, Password: password, Database: database}
}

func (r *Local) Create(info CreateInfo) error {
	createSql := fmt.Sprintf("CREATE DATABASE \"%s\"", info.Name)
	if err := r.ExecSQL(createSql, info.Timeout); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return buserr.New(constant.ErrDatabaseIsExist)
		}
		return err
	}

	if err := r.CreateUser(info, true); err != nil {
		_ = r.ExecSQL(fmt.Sprintf("DROP DATABASE \"%s\"", info.Name), info.Timeout)
		return err
	}

	return nil
}

func (r *Local) ChangePrivileges(info Privileges) error {
	super := "SUPERUSER"
	if !info.SuperUser {
		super = "NOSUPERUSER"
	}
	changeSql := fmt.Sprintf("ALTER USER \"%s\" WITH %s", info.Username, super)
	return r.ExecSQL(changeSql, info.Timeout)
}

func (r *Local) CreateUser(info CreateInfo, withDeleteDB bool) error {
	createSql := fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD '%s'", info.Username, info.Password)
	if err := r.ExecSQL(createSql, info.Timeout); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return buserr.New(constant.ErrUserIsExist)
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
	grantStr := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE \"%s\" TO \"%s\"", info.Name, info.Username)
	if err := r.ExecSQL(grantStr, info.Timeout); err != nil {
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

func (r *Local) Delete(info DeleteInfo) error {
	if len(info.Name) != 0 {
		dropSql := fmt.Sprintf("DROP DATABASE \"%s\"", info.Name)
		if err := r.ExecSQL(dropSql, info.Timeout); err != nil && !info.ForceDelete {
			return err
		}
	}
	dropSql := fmt.Sprintf("DROP USER \"%s\"", info.Username)
	if err := r.ExecSQL(dropSql, info.Timeout); err != nil && !info.ForceDelete {
		if strings.Contains(strings.ToLower(err.Error()), "depend on it") {
			return buserr.WithDetail(constant.ErrInUsed, info.Username, nil)
		}
		return err
	}
	return nil
}

func (r *Local) ChangePassword(info PasswordChangeInfo) error {
	changeSql := fmt.Sprintf("ALTER USER \"%s\" WITH PASSWORD '%s'", info.Username, info.Password)
	if err := r.ExecSQL(changeSql, info.Timeout); err != nil {
		return err
	}

	return nil
}

func (r *Local) Backup(info BackupInfo) error {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(info.TargetDir) {
		if err := os.MkdirAll(info.TargetDir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", info.TargetDir, err)
		}
	}
	outfile, err := os.OpenFile(path.Join(info.TargetDir, info.FileName), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("open file %s failed, err: %v", path.Join(info.TargetDir, info.FileName), err)
	}
	defer outfile.Close()
	global.LOG.Infof("start to pg_dump | gzip > %s.gzip", info.TargetDir+"/"+info.FileName)
	cmd := exec.Command("docker", "exec", r.ContainerName, "pg_dump", "-F", "c", "-U", r.Username, "-d", info.Name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	gzipCmd := exec.Command("gzip", "-cf")
	gzipCmd.Stdin, _ = cmd.StdoutPipe()
	gzipCmd.Stdout = outfile
	_ = gzipCmd.Start()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("handle backup database failed, err: %v", stderr.String())
	}
	_ = gzipCmd.Wait()
	return nil
}

func (r *Local) Recover(info RecoverInfo) error {
	fi, _ := os.Open(info.SourceFile)
	defer fi.Close()
	cmd := exec.Command("docker", "exec", "-i", r.ContainerName, "pg_restore", "-F", "c", "-c", "-U", r.Username, "-d", info.Name)
	if strings.HasSuffix(info.SourceFile, ".gz") {
		gzipFile, err := os.Open(info.SourceFile)
		if err != nil {
			return err
		}
		defer gzipFile.Close()
		gzipReader, err := gzip.NewReader(gzipFile)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		cmd.Stdin = gzipReader
	} else {
		cmd.Stdin = fi
	}
	stdout, err := cmd.CombinedOutput()
	if err != nil || strings.HasPrefix(string(stdout), "ERROR ") {
		return errors.New(string(stdout))
	}

	return nil
}

func (r *Local) SyncDB() ([]SyncDBInfo, error) {
	var datas []SyncDBInfo
	lines, err := r.ExecSQLForRows("SELECT datname FROM pg_database", 300)
	if err != nil {
		return datas, err
	}
	for _, line := range lines {
		itemLine := strings.TrimLeft(line, " ")
		if len(itemLine) == 0 || itemLine == "postgres" || itemLine == "template1" || itemLine == "template0" || itemLine == r.Username {
			continue
		}
		datas = append(datas, SyncDBInfo{Name: itemLine, From: "local", PostgresqlName: r.Database})
	}
	return datas, nil
}

func (r *Local) Close() {}

func (r *Local) ExecSQL(command string, timeout uint) error {
	itemCommand := r.PrefixCommand[:]
	itemCommand = append(itemCommand, command)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", itemCommand...)
	stdout, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return buserr.New(constant.ErrExecTimeOut)
	}
	if err != nil || strings.HasPrefix(string(stdout), "ERROR ") {
		return errors.New(string(stdout))
	}
	return nil
}

func (r *Local) ExecSQLForRows(command string, timeout uint) ([]string, error) {
	itemCommand := r.PrefixCommand[:]
	itemCommand = append(itemCommand, command)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", itemCommand...)
	stdout, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, buserr.New(constant.ErrExecTimeOut)
	}
	if err != nil || strings.HasPrefix(string(stdout), "ERROR ") {
		return nil, errors.New(string(stdout))
	}
	return strings.Split(string(stdout), "\n"), nil
}
