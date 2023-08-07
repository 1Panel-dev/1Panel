package client

import (
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
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type Local struct {
	PrefixCommand []string
	From          string
	Password      string
	ContainerName string
}

func NewLocal(command []string, containerName, password, from string) *Local {
	return &Local{PrefixCommand: command, ContainerName: containerName, Password: password, From: from}
}

func (r *Local) Create(info CreateInfo) error {
	createSql := fmt.Sprintf("create database `%s` default character set %s collate %s", info.Name, info.Format, formatMap[info.Format])
	if err := r.ExecSQL(createSql, info.Timeout); err != nil {
		if strings.Contains(err.Error(), "ERROR 1007") {
			return buserr.New(constant.ErrDatabaseIsExist)
		}
		return err
	}

	if err := r.CreateUser(info); err != nil {
		_ = r.ExecSQL(fmt.Sprintf("drop database if exists `%s`", info.Name), info.Timeout)
		return err
	}

	return nil
}

func (r *Local) CreateUser(info CreateInfo) error {
	var userlist []string
	if strings.Contains(info.Permission, ",") {
		ips := strings.Split(info.Permission, ",")
		for _, ip := range ips {
			if len(ip) != 0 {
				userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", info.Username, ip))
			}
		}
	} else {
		userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", info.Username, info.Permission))
	}

	for _, user := range userlist {
		if err := r.ExecSQL(fmt.Sprintf("create user %s identified by '%s';", user, info.Password), info.Timeout); err != nil {
			if strings.Contains(err.Error(), "ERROR 1396") {
				return buserr.New(constant.ErrUserIsExist)
			}
			_ = r.Delete(DeleteInfo{
				Name:        info.Name,
				Version:     info.Version,
				Username:    info.Username,
				Permission:  info.Permission,
				ForceDelete: true,
				Timeout:     300})
			return err
		}
		grantStr := fmt.Sprintf("grant all privileges on `%s`.* to %s", info.Name, user)
		if info.Name == "*" {
			grantStr = fmt.Sprintf("grant all privileges on *.* to %s", user)
		}
		if strings.HasPrefix(info.Version, "5.7") || strings.HasPrefix(info.Version, "5.6") {
			grantStr = fmt.Sprintf("%s identified by '%s' with grant option;", grantStr, info.Password)
		}
		if err := r.ExecSQL(grantStr, info.Timeout); err != nil {
			_ = r.Delete(DeleteInfo{
				Name:        info.Name,
				Version:     info.Version,
				Username:    info.Username,
				Permission:  info.Permission,
				ForceDelete: true,
				Timeout:     300})
			return err
		}
	}
	return nil
}

func (r *Local) Delete(info DeleteInfo) error {
	var userlist []string
	if strings.Contains(info.Permission, ",") {
		ips := strings.Split(info.Permission, ",")
		for _, ip := range ips {
			if len(ip) != 0 {
				userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", info.Username, ip))
			}
		}
	} else {
		userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", info.Username, info.Permission))
	}

	for _, user := range userlist {
		if strings.HasPrefix(info.Version, "5.6") {
			if err := r.ExecSQL(fmt.Sprintf("drop user %s", user), info.Timeout); err != nil && !info.ForceDelete {
				return err
			}
		} else {
			if err := r.ExecSQL(fmt.Sprintf("drop user if exists %s", user), info.Timeout); err != nil && !info.ForceDelete {
				return err
			}
		}
	}
	if len(info.Name) != 0 {
		if err := r.ExecSQL(fmt.Sprintf("drop database if exists `%s`", info.Name), info.Timeout); err != nil && !info.ForceDelete {
			return err
		}
	}
	if !info.ForceDelete {
		global.LOG.Info("execute delete database sql successful, now start to drop uploads and records")
	}

	return nil
}

func (r *Local) ChangePassword(info PasswordChangeInfo) error {
	if info.Username != "root" {
		var userlist []string
		if strings.Contains(info.Permission, ",") {
			ips := strings.Split(info.Permission, ",")
			for _, ip := range ips {
				if len(ip) != 0 {
					userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", info.Username, ip))
				}
			}
		} else {
			userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", info.Username, info.Permission))
		}

		for _, user := range userlist {
			passwordChangeSql := fmt.Sprintf("set password for %s = password('%s')", user, info.Password)
			if !strings.HasPrefix(info.Version, "5.7") && !strings.HasPrefix(info.Version, "5.6") {
				passwordChangeSql = fmt.Sprintf("ALTER USER %s IDENTIFIED WITH mysql_native_password BY '%s';", user, info.Password)
			}
			if err := r.ExecSQL(passwordChangeSql, info.Timeout); err != nil {
				return err
			}
		}
		return nil
	}

	hosts, err := r.ExecSQLForRows("select host from mysql.user where user='root';", info.Timeout)
	if err != nil {
		return err
	}
	for _, host := range hosts {
		if host == "%" || host == "localhost" {
			passwordRootChangeCMD := fmt.Sprintf("set password for 'root'@'%s' = password('%s')", host, info.Password)
			if !strings.HasPrefix(info.Version, "5.7") && !strings.HasPrefix(info.Version, "5.6") {
				passwordRootChangeCMD = fmt.Sprintf("alter user 'root'@'%s' identified with mysql_native_password BY '%s';", host, info.Password)
			}
			if err := r.ExecSQL(passwordRootChangeCMD, info.Timeout); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Local) ChangeAccess(info AccessChangeInfo) error {
	if info.Username == "root" {
		info.OldPermission = "%"
		info.Name = "*"
		info.Password = r.Password
	}
	if info.Permission != info.OldPermission {
		if err := r.Delete(DeleteInfo{
			Version:     info.Version,
			Username:    info.Username,
			Permission:  info.OldPermission,
			ForceDelete: true,
			Timeout:     300}); err != nil {
			return err
		}
		if info.Username == "root" {
			return nil
		}
	}
	if err := r.CreateUser(CreateInfo{
		Name:       info.Name,
		Version:    info.Version,
		Username:   info.Username,
		Password:   info.Password,
		Permission: info.Permission,
		Timeout:    info.Timeout,
	}); err != nil {
		return err
	}
	if err := r.ExecSQL("flush privileges", 300); err != nil {
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
	outfile, _ := os.OpenFile(path.Join(info.TargetDir, info.FileName), os.O_RDWR|os.O_CREATE, 0755)
	global.LOG.Infof("start to mysqldump | gzip > %s.gzip", info.TargetDir+"/"+info.FileName)
	cmd := exec.Command("docker", "exec", r.ContainerName, "mysqldump", "-uroot", "-p"+r.Password, info.Name)
	gzipCmd := exec.Command("gzip", "-cf")
	gzipCmd.Stdin, _ = cmd.StdoutPipe()
	gzipCmd.Stdout = outfile
	_ = gzipCmd.Start()
	_ = cmd.Run()
	_ = gzipCmd.Wait()
	return nil
}

func (r *Local) Recover(info RecoverInfo) error {
	fi, _ := os.Open(info.SourceFile)
	defer fi.Close()
	cmd := exec.Command("docker", "exec", "-i", r.ContainerName, "mysql", "-uroot", "-p"+r.Password, info.Name)
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
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return errors.New(stdStr)
	}

	return nil
}

func (r *Local) SyncDB(version string) ([]SyncDBInfo, error) {
	var datas []SyncDBInfo
	lines, err := r.ExecSQLForRows("SELECT SCHEMA_NAME, DEFAULT_CHARACTER_SET_NAME FROM information_schema.SCHEMATA", 300)
	if err != nil {
		return datas, err
	}
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == "SCHEMA_NAME" || parts[0] == "information_schema" || parts[0] == "mysql" || parts[0] == "performance_schema" || parts[0] == "sys" {
			continue
		}
		dataItem := SyncDBInfo{
			Name:   parts[0],
			From:   r.From,
			Format: parts[1],
		}
		userLines, err := r.ExecSQLForRows(fmt.Sprintf("SELECT USER,HOST FROM mysql.DB WHERE DB = '%s'", parts[0]), 300)
		if err != nil {
			return datas, err
		}

		var permissionItem []string
		isLocal := true
		i := 0
		for _, userline := range userLines {
			userparts := strings.Fields(userline)
			if len(userparts) != 2 {
				continue
			}
			if userparts[0] == "root" {
				continue
			}
			if i == 0 {
				dataItem.Username = userparts[0]
			}
			dataItem.Username = userparts[0]
			if dataItem.Username == userparts[0] && userparts[1] == "%" {
				isLocal = false
				dataItem.Permission = "%"
			} else if dataItem.Username == userparts[0] && userparts[1] != "localhost" {
				isLocal = false
				permissionItem = append(permissionItem, userparts[1])
			}
		}
		if len(dataItem.Username) == 0 {
			if err := r.CreateUser(CreateInfo{
				Name:       parts[0],
				Format:     parts[1],
				Version:    version,
				Username:   parts[0],
				Password:   common.RandStr(16),
				Permission: "%",
				Timeout:    300,
			}); err != nil {
				global.LOG.Errorf("sync from remote server failed, err: create user failed %v", err)
			}
			dataItem.Username = parts[0]
			dataItem.Permission = "%"
		} else {
			if isLocal {
				dataItem.Permission = "localhost"
			}
			if len(dataItem.Permission) == 0 {
				dataItem.Permission = strings.Join(permissionItem, ",")
			}
		}
		datas = append(datas, dataItem)
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
	if ctx.Err() == context.DeadlineExceeded {
		return buserr.New(constant.ErrExecTimeOut)
	}
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return errors.New(stdStr)
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
	if ctx.Err() == context.DeadlineExceeded {
		return nil, buserr.New(constant.ErrExecTimeOut)
	}
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return nil, errors.New(stdStr)
	}
	return strings.Split(stdStr, "\n"), nil
}
