package service

import (
	"context"
	"errors"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpapi"
)

func (u *MysqlService) CreateForMCP(ctx context.Context, req dto.MysqlDBCreate) (*model.DatabaseMysql, error) {
	if req.From != "local" || ctx.Err() != nil {
		return nil, errors.New("invalid MCP database target or request")
	}
	writeContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 310*time.Second)
	defer cancel()
	created, err := u.create(writeContext, req, true)
	if err != nil {
		return nil, mcpapi.SanitizeBusinessError(err, "database creation failed; inspect the invocation before retrying")
	}
	return created, nil
}
