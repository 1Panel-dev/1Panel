package mysql

import (
	"database/sql"
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql/client"
)

type MysqlClient interface {
	Create(info client.CreateInfo) error
	Delete(info client.DeleteInfo) error

	ChangePassword(info client.PasswordChangeInfo) error
	ChangeAccess(info client.AccessChangeInfo) error

	Backup(info client.BackupInfo) error
	Recover(info client.RecoverInfo) error

	Close()
}

func NewMysqlClient(conn client.DBInfo) (MysqlClient, error) {
	if conn.From == "local" {
		if cmd.CheckIllegal(conn.Address, conn.Username, conn.Password) {
			return nil, buserr.New(constant.ErrCmdIllegal)
		}
		connArgs := []string{"exec", conn.Address, "mysql", "-u" + conn.Username, "-p" + conn.Password, "-e"}
		return client.NewLocal(connArgs, conn.Address, conn.Password), nil
	}

	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", conn.Username, conn.Password, conn.Address, conn.Port)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return client.NewRemote(client.Remote{
		Client:   db,
		User:     conn.Username,
		Password: conn.Password,
		Address:  conn.Address,
		Port:     conn.Port,
	}), nil
}
