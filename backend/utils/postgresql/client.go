package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/postgresql/client"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type PostgresqlClient interface {
	Create(info client.CreateInfo) error
	Delete(info client.DeleteInfo) error
	ReloadConf()error
	ChangePassword(info client.PasswordChangeInfo) error
	ChangeAccess(info client.AccessChangeInfo) error

	Backup(info client.BackupInfo) error
	Recover(info client.RecoverInfo) error
	Status() client.Status
	SyncDB(version string) ([]client.SyncDBInfo, error)
	Close()
}

func NewPostgresqlClient(conn client.DBInfo) (PostgresqlClient, error) {
	if conn.Port==0 {
		conn.Port=5432
	}
	if conn.From == "local" {
		conn.Address = "127.0.0.1"
	}
	connArgs := fmt.Sprintf("postgres://%s:%s@%s:%d/?sslmode=disable", conn.Username, conn.Password, conn.Address, conn.Port)
	db, err := sql.Open("pgx", connArgs)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, buserr.New(constant.ErrExecTimeOut)
	}

	return client.NewRemote(client.Remote{
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
