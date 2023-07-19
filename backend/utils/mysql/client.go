package mysql

import (
	"database/sql"
	"errors"
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

	Close()
}

func NewMysqlClient(conn client.DBInfo) (MysqlClient, error) {
	if conn.Type == "remote" {
		connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", conn.UserName, conn.Password, conn.Address, conn.Port)
		db, err := sql.Open("mysql", connArgs)
		if err != nil {
			return nil, err
		}
		return client.NewRemote(db), nil
	}
	if conn.Type == "local" {
		if cmd.CheckIllegal(conn.Address, conn.UserName, conn.Password) {
			return nil, buserr.New(constant.ErrCmdIllegal)
		}
		connArgs := []string{"exec", conn.Address, "mysql", "-u" + conn.UserName, "-p" + conn.Password + "-e"}
		return client.NewLocal(connArgs, conn.Address), nil
	}
	return nil, errors.New("no such type")
}
