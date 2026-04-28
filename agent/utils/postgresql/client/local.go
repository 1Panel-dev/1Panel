package client

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
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
			return buserr.New("ErrDatabaseIsExist")
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
	global.LOG.Infof("start to pg_dump | gzip > %s.gzip", info.TargetDir+"/"+info.FileName)

	cmdMgr := cmd.NewCommandMgr(cmd.WithOutputFile(path.Join(info.TargetDir, info.FileName)))
	if _, err := cmdMgr.RunPipe(
		cmd.PipeCommand{Name: "docker", Args: []string{"exec", "-i", "-e", "PGPASSWORD=" + r.Password, r.ContainerName, "pg_dump", "-F", "c", "-U", r.Username, "-d", info.Name}},
		cmd.PipeCommand{Name: "gzip", Args: []string{"-cf"}},
	); err != nil {
		return fmt.Errorf("handle backup database failed, err: %v", err)
	}
	return nil
}

func (r *Local) Recover(info RecoverInfo) error {
	fi, _ := os.Open(info.SourceFile)
	defer fi.Close()

	cmd := exec.Command("docker", "exec", "-i", "-e", "PGPASSWORD="+r.Password, r.ContainerName,
		"pg_restore", "-F", "c", "-c", "--if-exists", "--no-owner", "-U", r.Username, "-d", info.Name,
	)
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
	if err != nil {
		if strings.HasPrefix(string(stdout), "ERROR ") {
			return errors.New(string(stdout))
		}
		return err
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
		if len(itemLine) == 0 || itemLine == "template1" || itemLine == "template0" || itemLine == r.Username {
			continue
		}
		datas = append(datas, SyncDBInfo{Name: itemLine, From: "local", PostgresqlName: r.Database})
	}
	return datas, nil
}

func (r *Local) Close() {}

func (r *Local) isDatabaseInUse(name string, timeout uint) (bool, error) {
	escapedName := strings.ReplaceAll(name, "'", "''")
	checkSQL := fmt.Sprintf(
		"SELECT COUNT(*) FROM pg_stat_activity WHERE datname='%s' AND pid <> pg_backend_pid()",
		escapedName,
	)
	lines, err := r.ExecSQLForRows(checkSQL, timeout)
	if err != nil {
		return false, err
	}
	for _, line := range lines {
		countStr := strings.TrimSpace(line)
		if len(countStr) == 0 {
			continue
		}
		count, parseErr := strconv.Atoi(countStr)
		if parseErr != nil {
			return false, parseErr
		}
		return count > 0, nil
	}
	return false, nil
}

func (r *Local) ExecSQL(command string, timeout uint) error {
	itemCommand := r.PrefixCommand[:]
	itemCommand = append(itemCommand, command)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", itemCommand...)
	stdout, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return buserr.New("ErrExecTimeOut")
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
		return nil, buserr.New("ErrExecTimeOut")
	}
	if err != nil || strings.HasPrefix(string(stdout), "ERROR ") {
		return nil, errors.New(string(stdout))
	}
	return strings.Split(string(stdout), "\n"), nil
}
