package client

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Remote struct {
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
func (r *Remote) Status() Status {
	status := Status{}
	var i int64
	var s string
	var f float64
	_ = r.Client.QueryRow("select count(*) from pg_stat_activity WHERE client_addr is not NULL;").Scan(&i)
	status.CurrentConnections = fmt.Sprintf("%d",i)
	_ = r.Client.QueryRow("SELECT current_timestamp - pg_postmaster_start_time();").Scan(&s)
	before,_, _ := strings.Cut(s, ".")
	status.Uptime = before
	_ = r.Client.QueryRow("select sum(blks_hit)*100/sum(blks_hit+blks_read) as hit_ratio from pg_stat_database;").Scan(&f)
	status.HitRatio = fmt.Sprintf("%0.2f",f)
	var a1,a2,a3 int64
	_ = r.Client.QueryRow("select buffers_clean, maxwritten_clean, buffers_backend_fsync from pg_stat_bgwriter;").Scan(&a1, &a2, &a3)
	status.BuffersClean = fmt.Sprintf("%d",a1)
	status.MaxwrittenClean = fmt.Sprintf("%d",a2)
	status.BuffersBackendFsync= fmt.Sprintf("%d",a3)
	rows, err := r.Client.Query("SHOW ALL;")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var k,v string
			err := rows.Scan(&k, &v,&s)
			if err != nil {
				continue
			}
			if k == "autovacuum" {
				status.Autovacuum = v
			}
			if k == "max_connections" {
				status.MaxConnections = v
			}
			if k == "server_version" {
				status.Version = v
			}
			if k == "shared_buffers" {
				status.SharedBuffers = v
			}
		}
	}
	return status
}
func (r *Remote) Create(info CreateInfo) error {
	createUser := fmt.Sprintf(`CREATE USER "%s" WITH PASSWORD '%s';`, info.Username, info.Password)
	createDB := fmt.Sprintf(`CREATE DATABASE "%s" OWNER "%s";`, info.Name, info.Username)
	grant := fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE "%s" TO "%s";`, info.Name, info.Username)
	if err := r.ExecSQL(createUser, info.Timeout); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already") {
			return buserr.New(constant.ErrUserIsExist)
		}
		return err
	}
	if err := r.ExecSQL(createDB, info.Timeout); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already") {
			_ = r.ExecSQL(fmt.Sprintf(`DROP DATABASE "%s"`, info.Name), info.Timeout)
			return buserr.New(constant.ErrDatabaseIsExist)
		}
		return err
	}
	_ = r.ExecSQL(grant, info.Timeout)
	return nil
}

func (r *Remote) CreateUser(info CreateInfo, withDeleteDB bool) error {
	sql1 := fmt.Sprintf(`CREATE USER "%s" WITH PASSWORD '%s';
GRANT ALL PRIVILEGES ON DATABASE "%s" TO "%s";`, info.Username, info.Password, info.Name, info.Username)
	err := r.ExecSQL(sql1, info.Timeout)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already") {
			return buserr.New(constant.ErrUserIsExist)
		}
	}

	return nil
}

func (r *Remote) Delete(info DeleteInfo) error {
	//暂时不支持强制删除,就算附加了 WITH(FORCE) 也会删除失败
	err := r.ExecSQL(fmt.Sprintf(`DROP DATABASE "%s"`, info.Name), info.Timeout)
	if err != nil {
		return err
	}
	return r.ExecSQL(fmt.Sprintf(`DROP USER "%s"`, info.Username), info.Timeout)
}

func (r *Remote) ChangePassword(info PasswordChangeInfo) error {
	return r.ExecSQL(fmt.Sprintf(`ALTER USER "%s" WITH ENCRYPTED PASSWORD '%s';`, info.Username, info.Password), info.Timeout)
}
func (r *Remote) ReloadConf()error {
	return r.ExecSQL("SELECT pg_reload_conf();",5)
}
func (r *Remote) ChangeAccess(info AccessChangeInfo) error {
	return nil
}

func (r *Remote) Backup(info BackupInfo) error {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(info.TargetDir) {
		if err := os.MkdirAll(info.TargetDir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", info.TargetDir, err)
		}
	}
	fileNameItem := info.TargetDir + "/" + strings.TrimSuffix(info.FileName, ".gz")

	backupCommand := exec.Command("bash", "-c",
		fmt.Sprintf("docker run --rm --net=host -i postgres:alpine /bin/bash -c 'PGPASSWORD=%s pg_dump  -h %s -p %d --no-owner -Fc -U %s %s' > %s",
			r.Password, r.Address, r.Port, r.User, info.Name, fileNameItem))
	_ = backupCommand.Run()
	b := make([]byte, 5)
	n := []byte{80, 71, 68, 77, 80}
	handle, err := os.OpenFile(fileNameItem, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("backup file not found,err:%v", err)
	}
	defer handle.Close()
	_, _ = handle.Read(b)
	if string(b) != string(n) {
		errBytes, _ := os.ReadFile(fileNameItem)
		return fmt.Errorf("backup failed,err:%s", string(errBytes))
	}

	gzipCmd := exec.Command("gzip", fileNameItem)
	stdout, err := gzipCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gzip file %s failed, stdout: %v, err: %v", strings.TrimSuffix(info.FileName, ".gz"), string(stdout), err)
	}
	return nil
}

func (r *Remote) Recover(info RecoverInfo) error {
	fileName := info.SourceFile
	if strings.HasSuffix(info.SourceFile, ".sql.gz") {
		fileName = strings.TrimSuffix(info.SourceFile, ".gz")
		gzipCmd := exec.Command("gunzip", info.SourceFile)
		stdout, err := gzipCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("gunzip file %s failed,	 stdout: %v, err: %v", info.SourceFile, string(stdout), err)
		}
		defer func() {
			gzipCmd := exec.Command("gzip", fileName)
			_, _ = gzipCmd.CombinedOutput()
		}()
	}
	recoverCommand := exec.Command("bash", "-c",
		fmt.Sprintf("docker run --rm --net=host -i postgres:alpine /bin/bash -c 'PGPASSWORD=%s pg_restore -h %s -p %d --verbose --clean --no-privileges --no-owner -Fc -U %s -d %s --role=%s' < %s",
			r.Password, r.Address, r.Port, r.User, info.Name, info.Username, fileName))
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
			global.LOG.Errorf("[Postgresql] DB:[%s] Recover Error: %s", info.Name, string(all))
			return err
		}
		global.LOG.Infof("[Postgresql] DB:[%s] Restoring: %s", info.Name, readString)
	}

	return nil
}

func (r *Remote) SyncDB(version string) ([]SyncDBInfo, error) {
	//如果需要同步数据库,则需要强制修改用户密码,否则无法获取真实密码,后面可考虑改为添加服务器账号,手动将账号/数据库添加到管理列表
	var datas []SyncDBInfo
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
	if ctx.Err() == context.DeadlineExceeded {
		return buserr.New(constant.ErrExecTimeOut)
	}

	return nil
}
