package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/postgresql/client"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresqlClient interface {
	Create(info client.CreateInfo) error
	Delete(info client.DeleteInfo) error
	ChangePassword(info client.PasswordChangeInfo) error

	Backup(info client.BackupInfo) error
	Recover(info client.RecoverInfo) error
	SyncDB() ([]client.SyncDBInfo, error)
	Close()
}

func NewPostgresqlClient(conn client.DBInfo) (PostgresqlClient, error) {
	if conn.From == "local" {
		connArgs := []string{"exec", conn.Address, "psql", "-t", "-U", conn.Username, "-c"}
		return client.NewLocal(connArgs, conn.Address, conn.Username, conn.Password, conn.Database), nil
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
	}), nil
}
