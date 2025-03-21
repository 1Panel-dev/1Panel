package service

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/jinzhu/copier"
)

type CommandService struct{}

type ICommandService interface {
	List(req dto.OperateByType) ([]dto.CommandInfo, error)
	SearchForTree(req dto.OperateByType) ([]dto.CommandTree, error)
	SearchWithPage(search dto.SearchCommandWithPage) (int64, interface{}, error)
	Create(req dto.CommandOperate) error
	Update(req dto.CommandOperate) error
	Delete(ids []uint) error
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (u *CommandService) List(req dto.OperateByType) ([]dto.CommandInfo, error) {
	commands, err := commandRepo.List(repo.WithOrderBy("name"), repo.WithByType(req.Type))
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}
	var dtoCommands []dto.CommandInfo
	for _, command := range commands {
		var item dto.CommandInfo
		if err := copier.Copy(&item, &command); err != nil {
			return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		dtoCommands = append(dtoCommands, item)
	}
	return dtoCommands, err
}

func (u *CommandService) SearchForTree(req dto.OperateByType) ([]dto.CommandTree, error) {
	cmdList, err := commandRepo.List(repo.WithOrderBy("name"), repo.WithByType(req.Type))
	if err != nil {
		return nil, err
	}
	groups, err := groupRepo.GetList(repo.WithByType(req.Type))
	if err != nil {
		return nil, err
	}
	var lists []dto.CommandTree
	for _, group := range groups {
		var data dto.CommandTree
		data.ID = group.ID + 10000
		data.Label = group.Name
		for _, cmd := range cmdList {
			if cmd.GroupID == group.ID {
				data.Children = append(data.Children, dto.CommandInfo{ID: cmd.ID, Name: cmd.Name, Command: cmd.Command})
			}
		}
		if len(data.Children) != 0 {
			lists = append(lists, data)
		}
	}
	return lists, err
}

func (u *CommandService) SearchWithPage(req dto.SearchCommandWithPage) (int64, interface{}, error) {
	options := []global.DBOption{
		repo.WithOrderRuleBy(req.OrderBy, req.Order),
		repo.WithByType(req.Type),
	}
	if len(req.Info) != 0 {
		options = append(options, commandRepo.WithByInfo(req.Info))
	}
	if req.GroupID != 0 {
		options = append(options, repo.WithByGroupID(req.GroupID))
	}
	total, commands, err := commandRepo.Page(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	groups, _ := groupRepo.GetList(repo.WithByType(req.Type))
	var dtoCommands []dto.CommandInfo
	for _, command := range commands {
		var item dto.CommandInfo
		if err := copier.Copy(&item, &command); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		for _, group := range groups {
			if command.GroupID == group.ID {
				item.GroupBelong = group.Name
				item.GroupID = group.ID
			}
		}
		dtoCommands = append(dtoCommands, item)
	}
	return total, dtoCommands, err
}

func (u *CommandService) Create(req dto.CommandOperate) error {
	command, _ := commandRepo.Get(repo.WithByName(req.Name), repo.WithByType(req.Type))
	if command.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	if err := copier.Copy(&command, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	if err := commandRepo.Create(&command); err != nil {
		return err
	}
	return nil
}

func (u *CommandService) Delete(ids []uint) error {
	if len(ids) == 1 {
		command, _ := commandRepo.Get(repo.WithByID(ids[0]))
		if command.ID == 0 {
			return buserr.New("ErrRecordNotFound")
		}
		return commandRepo.Delete(repo.WithByID(ids[0]))
	}
	return commandRepo.Delete(repo.WithByIDs(ids))
}

func (u *CommandService) Update(req dto.CommandOperate) error {
	command, _ := commandRepo.Get(repo.WithByName(req.Name), repo.WithByType(req.Type))
	if command.ID != req.ID {
		return buserr.New("ErrRecordExist")
	}
	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["group_id"] = req.GroupID
	upMap["command"] = req.Command
	return commandRepo.Update(req.ID, upMap)
}
