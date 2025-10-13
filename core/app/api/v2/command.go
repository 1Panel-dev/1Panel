package v2

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/gin-gonic/gin"
)

// @Tags Command
// @Summary Upload command csv for list
// @Success 200 {string} path
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/import [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"上传快速命令文件","formatEN":"upload quick commands with csv"}
func (b *BaseApi) UploadCommandCsv(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		helper.BadRequest(c, errors.New("no such files"))
		return
	}
	uploadFile, _ := files[0].Open()
	reader := csv.NewReader(uploadFile)
	if _, err := reader.Read(); err != nil {
		helper.BadRequest(c, fmt.Errorf("read title failed, err: %v", err))
		return
	}
	groupRepo := repo.NewIGroupRepo()
	group, _ := groupRepo.Get(repo.WithByType("command"), groupRepo.WithByDefault(true))
	var commands []dto.CommandInfo
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			helper.BadRequest(c, fmt.Errorf("read content failed, err: %v", err))
			return
		}
		if len(record) >= 2 {
			commands = append(commands, dto.CommandInfo{
				Name:        record[0],
				Type:        "command",
				GroupID:     group.ID,
				Command:     record[1],
				GroupBelong: group.Name,
			})
		}
	}

	helper.SuccessWithData(c, commands)
}

// @Tags Command
// @Summary Export command
// @Success 200 {string} path
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/export [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"导出快速命令","formatEN":"export quick commands"}
func (b *BaseApi) ExportCommands(c *gin.Context) {
	file, err := commandService.Export()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, file)
}

// @Tags Command
// @Summary Import command
// @Success 200 {string} path
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/import [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"导入快速命令","formatEN":"import quick commands"}
func (b *BaseApi) ImportCommands(c *gin.Context) {
	var req dto.CommandImport
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	groupRepo := repo.NewIGroupRepo()
	group, _ := groupRepo.Get(repo.WithByType("command"), groupRepo.WithByDefault(true))
	for _, item := range req.Items {
		if item.GroupID == 0 {
			item.GroupID = group.ID
		}
		_ = commandService.Create(item)
	}
	helper.Success(c)
}

// @Tags Command
// @Summary Create command
// @Accept json
// @Param request body dto.CommandOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands [post]
// @x-panel-log {"bodyKeys":["name","command"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建快捷命令 [name][command]","formatEN":"create quick command [name][command]"}
func (b *BaseApi) CreateCommand(c *gin.Context) {
	var req dto.CommandOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := commandService.Create(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Command
// @Summary Page commands
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/search [post]
func (b *BaseApi) SearchCommand(c *gin.Context) {
	var req dto.SearchCommandWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := commandService.SearchWithPage(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Command
// @Summary Tree commands
// @Accept json
// @Param request body dto.OperateByType true "request"
// @Success 200 {Array} dto.CommandTree
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/tree [get]
func (b *BaseApi) SearchCommandTree(c *gin.Context) {
	var req dto.OperateByType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	list, err := commandService.SearchForTree(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Command
// @Summary List commands
// @Accept json
// @Param request body dto.OperateByType true "request"
// @Success 200 {object} dto.CommandInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/command [get]
func (b *BaseApi) ListCommand(c *gin.Context) {
	var req dto.OperateByType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	list, err := commandService.List(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Command
// @Summary Delete command
// @Accept json
// @Param request body dto.OperateByIDs true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"commands","output_column":"name","output_value":"names"}],"formatZH":"删除快捷命令 [names]","formatEN":"delete quick command [names]"}
func (b *BaseApi) DeleteCommand(c *gin.Context) {
	var req dto.OperateByIDs
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := commandService.Delete(req.IDs); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Command
// @Summary Update command
// @Accept json
// @Param request body dto.CommandOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/commands/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新快捷命令 [name]","formatEN":"update quick command [name]"}
func (b *BaseApi) UpdateCommand(c *gin.Context) {
	var req dto.CommandOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := commandService.Update(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
