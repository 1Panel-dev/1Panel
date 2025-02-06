package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
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
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type Remote struct {
	Type     string
	Client   *sql.DB
	Database string
	User     string
	Password string
	Address  string
	Port     uint

	SSL        bool
	RootCert   string
	ClientKey  string
	ClientCert string
	SkipVerify bool
}

func NewRemote(db Remote) *Remote {
	return &db
}

func (r *Remote) Create(info CreateInfo) error {
	createSql := fmt.Sprintf("create database `%s` default character set %s collate %s", info.Name, info.Format, formatMap[info.Format])
	if err := r.ExecSQL(createSql, info.Timeout); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "error 1007") {
			return buserr.New(constant.ErrDatabaseIsExist)
		}
		return err
	}

	if err := r.CreateUser(info, true); err != nil {
		_ = r.ExecSQL(fmt.Sprintf("drop database if exists `%s`", info.Name), info.Timeout)
		return err
	}

	return nil
}

func (r *Remote) CreateUser(info CreateInfo, withDeleteDB bool) error {
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
			if strings.Contains(strings.ToLower(err.Error()), "error 1396") {
				return buserr.New(constant.ErrUserIsExist)
			}
			if withDeleteDB {
				_ = r.Delete(DeleteInfo{
					Name:        info.Name,
					Version:     info.Version,
					Username:    info.Username,
					Permission:  info.Permission,
					ForceDelete: true,
					Timeout:     300})
			}
			return err
		}
		grantStr := fmt.Sprintf("grant all privileges on `%s`.* to %s", info.Name, user)
		if info.Name == "*" {
			grantStr = fmt.Sprintf("grant all privileges on *.* to %s", user)
		}
		if strings.HasPrefix(info.Version, "5.7") || strings.HasPrefix(info.Version, "5.6") {
			grantStr = fmt.Sprintf("%s identified by '%s' with grant option;", grantStr, info.Password)
		} else {
			grantStr = grantStr + " with grant option;"
		}
		if err := r.ExecSQL(grantStr, info.Timeout); err != nil {
			if withDeleteDB {
				_ = r.Delete(DeleteInfo{
					Name:        info.Name,
					Version:     info.Version,
					Username:    info.Username,
					Permission:  info.Permission,
					ForceDelete: true,
					Timeout:     300})
			}
			return err
		}
	}
	return nil
}

func (r *Remote) Delete(info DeleteInfo) error {
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

func (r *Remote) ChangePassword(info PasswordChangeInfo) error {
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
				passwordChangeSql = fmt.Sprintf("ALTER USER %s IDENTIFIED BY '%s';", user, info.Password)
			}
			if err := r.ExecSQL(passwordChangeSql, info.Timeout); err != nil {
				return err
			}
		}
		return nil
	}

	hosts, err := r.ExecSQLForHosts(info.Timeout)
	if err != nil {
		return err
	}
	for _, host := range hosts {
		if host == "%" || host == "localhost" {
			passwordRootChangeCMD := fmt.Sprintf("set password for 'root'@'%s' = password('%s')", host, info.Password)
			if !strings.HasPrefix(info.Version, "5.7") && !strings.HasPrefix(info.Version, "5.6") {
				passwordRootChangeCMD = fmt.Sprintf("alter user 'root'@'%s' identified by '%s';", host, info.Password)
			}
			if err := r.ExecSQL(passwordRootChangeCMD, info.Timeout); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Remote) ChangeAccess(info AccessChangeInfo) error {
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
	}, false); err != nil {
		return err
	}
	if err := r.ExecSQL("flush privileges", 300); err != nil {
		return err
	}
	return nil
}

func (r *Remote) Backup(info BackupInfo) error {
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
	dumpCmd := "mysqldump"
	if r.Type == constant.AppMariaDB {
		dumpCmd = "mariadb-dump"
	}
	global.LOG.Infof("start to %s | gzip > %s.gzip", dumpCmd, info.TargetDir+"/"+info.FileName)
	image, err := loadImage(info.Type, info.Version)
	if err != nil {
		return err
	}
	backupCmd := fmt.Sprintf("docker run --rm --net=host -i %s /bin/bash -c '%s --routines -h %s -P %d -u%s -p%s %s --default-character-set=%s %s'",
		image, dumpCmd, r.Address, r.Port, r.User, r.Password, sslSkip(info.Version, r.Type), info.Format, info.Name)

	global.LOG.Debug(strings.ReplaceAll(backupCmd, r.Password, "******"))
	cmd := exec.Command("bash", "-c", backupCmd)
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

func (r *Remote) Recover(info RecoverInfo) error {
	fi, _ := os.Open(info.SourceFile)
	defer fi.Close()

	image, err := loadImage(info.Type, info.Version)
	if err != nil {
		return err
	}

	recoverCmd := fmt.Sprintf("docker run --rm --net=host -i %s /bin/bash -c '%s -h %s -P %d -u%s -p%s %s --default-character-set=%s %s'",
		image, r.Type, r.Address, r.Port, r.User, r.Password, sslSkip(info.Version, r.Type), info.Format, info.Name)

	global.LOG.Debug(strings.ReplaceAll(recoverCmd, r.Password, "******"))
	cmd := exec.Command("bash", "-c", recoverCmd)

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

func (r *Remote) SyncDB(version string) ([]SyncDBInfo, error) {
	var datas []SyncDBInfo
	rows, err := r.Client.Query("select schema_name, default_character_set_name from information_schema.SCHEMATA")
	if err != nil {
		return datas, err
	}
	defer rows.Close()

	for rows.Next() {
		var dbName, charsetName string
		if err = rows.Scan(&dbName, &charsetName); err != nil {
			return datas, err
		}
		if dbName == "information_schema" || dbName == "mysql" || dbName == "performance_schema" || dbName == "sys" || dbName == "__recycle_bin__" || dbName == "recycle_bin" {
			continue
		}
		dataItem := SyncDBInfo{
			Name:      dbName,
			From:      "remote",
			MysqlName: r.Database,
			Format:    charsetName,
		}
		userRows, err := r.Client.Query("select user,host from mysql.db where db = ?", dbName)
		if err != nil {
			global.LOG.Debugf("sync user of db %s failed, err: %v", dbName, err)
			dataItem.Permission = "%"
			datas = append(datas, dataItem)
			continue
		}

		var permissionItem []string
		isLocal := true
		i := 0
		for userRows.Next() {
			var user, host string
			if err = userRows.Scan(&user, &host); err != nil {
				return datas, err
			}
			if user == "root" {
				continue
			}
			if i == 0 {
				dataItem.Username = user
			}
			if dataItem.Username == user && host == "%" {
				isLocal = false
				dataItem.Permission = "%"
			} else if dataItem.Username == user && host != "localhost" {
				isLocal = false
				permissionItem = append(permissionItem, host)
			}
			i++
		}
		if len(dataItem.Username) == 0 {
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
	if err = rows.Err(); err != nil {
		return datas, err
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
		return buserr.New(constant.ErrExecTimeOut)
	}

	return nil
}

func (r *Remote) ExecSQLForHosts(timeout uint) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	results, err := r.Client.QueryContext(ctx, "select host from mysql.user where user='root';")
	if err != nil {
		return nil, err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, buserr.New(constant.ErrExecTimeOut)
	}
	var rows []string
	for results.Next() {
		var host string
		if err := results.Scan(&host); err != nil {
			continue
		}
		rows = append(rows, host)
	}
	return rows, nil
}

func loadImage(dbType, version string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	images, err := cli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if !strings.HasPrefix(tag, dbType+":") {
				continue
			}
			if dbType == "mariadb" && strings.HasPrefix(tag, "mariadb:") {
				return tag, nil
			}
			if strings.HasPrefix(version, "5.6") && strings.HasPrefix(tag, "mysql:5.6") {
				return tag, nil
			}
			if strings.HasPrefix(version, "5.7") && strings.HasPrefix(tag, "mysql:5.7") {
				return tag, nil
			}
			if strings.HasPrefix(version, "8.") && strings.HasPrefix(tag, "mysql:8.") {
				return tag, nil
			}
		}
	}
	return loadVersion(dbType, version), nil
}

func loadVersion(dbType string, version string) string {
	if dbType == "mariadb" {
		return "mariadb:11.3.2 "
	}
	if strings.HasPrefix(version, "5.6") {
		return "mysql:5.6.51"
	}
	if strings.HasPrefix(version, "5.7") {
		return "mysql:5.7.44"
	}
	return "mysql:8.2.0"
}

func sslSkip(version, dbType string) string {
	if dbType == constant.AppMariaDB || strings.HasPrefix(version, "5.6") || strings.HasPrefix(version, "5.7") {
		return "--skip-ssl"
	}
	return "--ssl-mode=DISABLED"
}
