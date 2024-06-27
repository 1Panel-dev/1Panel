package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql/client"
)

type MysqlClient interface {
	Create(info client.CreateInfo) error
	CreateUser(info client.CreateInfo, withDeleteDB bool) error
	Delete(info client.DeleteInfo) error

	ChangePassword(info client.PasswordChangeInfo) error
	ChangeAccess(info client.AccessChangeInfo) error

	Backup(info client.BackupInfo) error
	Recover(info client.RecoverInfo) error

	SyncDB(version string) ([]client.SyncDBInfo, error)
	Close()
}

func NewMysqlClient(conn client.DBInfo) (MysqlClient, error) {
	if conn.From == "local" {
		connArgs := []string{"exec", conn.Address, conn.Type, "-u" + conn.Username, "-p" + conn.Password, "-e"}
		return client.NewLocal(connArgs, conn.Type, conn.Address, conn.Password, conn.Database), nil
	}

	if strings.Contains(conn.Address, ":") {
		conn.Address = fmt.Sprintf("[%s]", conn.Address)
	}

	tlsItem, err := client.ConnWithSSL(conn.SSL, conn.SkipVerify, conn.ClientKey, conn.ClientCert, conn.RootCert)
	if err != nil {
		return nil, err
	}
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8%s", conn.Username, conn.Password, conn.Address, conn.Port, tlsItem)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		global.LOG.Errorf("test mysql conn failed, err: %v", err)
		return nil, err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, buserr.New(constant.ErrExecTimeOut)
	}

	return client.NewRemote(client.Remote{
		Type:     conn.Type,
		Client:   db,
		Database: conn.Database,
		User:     conn.Username,
		Password: conn.Password,
		Address:  conn.Address,
		Port:     conn.Port,

		SSL:        conn.SSL,
		RootCert:   conn.RootCert,
		ClientKey:  conn.ClientKey,
		ClientCert: conn.ClientCert,
		SkipVerify: conn.SkipVerify,
	}), nil
}
