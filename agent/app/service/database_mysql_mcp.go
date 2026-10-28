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
	created, err := createMCPMysqlDatabase(writeContext, req)
	if err != nil {
		return nil, mcpapi.SanitizeBusinessError(err, "database creation failed; inspect the invocation before retrying")
	}
	return created, nil
}

func createMCPMysqlDatabase(ctx context.Context, req dto.MysqlDBCreate) (*model.DatabaseMysql, error) {
	creation, err := prepareMysqlCreation(req)
	if err != nil {
		return nil, err
	}
	local, version, err := loadInstalledLocalMysqlClient(creation.request.Database)
	if err != nil {
		return nil, err
	}
	defer local.Close()
	deadline := time.Now().Add(300 * time.Second)
	if writeDeadline, ok := ctx.Deadline(); ok && writeDeadline.Add(-10*time.Second).Before(deadline) {
		deadline = writeDeadline.Add(-10 * time.Second)
	}
	sqlContext, cancel := context.WithDeadline(ctx, deadline)
	err = local.CreateWithContext(sqlContext, creation.createInfo(version))
	cancel()
	if err != nil {
		return nil, err
	}
	if err := saveMysqlCreationMetadata(ctx, creation.request, creation.permissionHosts, &creation.item); err != nil {
		return nil, err
	}
	return &creation.item, nil
}
