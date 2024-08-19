package service

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type CommandService struct{}

type ICommandService interface {
	List(req dto.OperateByType) ([]dto.CommandInfo, error)
	SearchForTree(req dto.OperateByType) ([]dto.CommandTree, error)
	SearchWithPage(search dto.SearchCommandWithPage) (int64, interface{}, error)
	Create(commandDto dto.CommandOperate) error
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (u *CommandService) List(req dto.OperateByType) ([]dto.CommandInfo, error) {
	commands, err := commandRepo.GetList(commonRepo.WithOrderBy("name"), commonRepo.WithByType(req.Type))
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	var dtoCommands []dto.CommandInfo
	for _, command := range commands {
		var item dto.CommandInfo
		if err := copier.Copy(&item, &command); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoCommands = append(dtoCommands, item)
	}
	return dtoCommands, err
}

func (u *CommandService) SearchForTree(req dto.OperateByType) ([]dto.CommandTree, error) {
	cmdList, err := commandRepo.GetList(commonRepo.WithOrderBy("name"), commonRepo.WithByType(req.Type))
	if err != nil {
		return nil, err
	}
	groups, err := groupRepo.GetList(commonRepo.WithByType(req.Type))
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
	options := []repo.DBOption{
		commonRepo.WithOrderRuleBy(req.OrderBy, req.Order),
		commonRepo.WithByType(req.Type),
	}
	if len(req.Info) != 0 {
		options = append(options, commandRepo.WithLikeName(req.Info))
	}
	if req.GroupID != 0 {
		options = append(options, groupRepo.WithByGroupID(req.GroupID))
	}
	total, commands, err := commandRepo.Page(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	groups, _ := groupRepo.GetList(commonRepo.WithByType(req.Type))
	var dtoCommands []dto.CommandInfo
	for _, command := range commands {
		var item dto.CommandInfo
		if err := copier.Copy(&item, &command); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
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

func (u *CommandService) Create(commandDto dto.CommandOperate) error {
	command, _ := commandRepo.Get(commonRepo.WithByName(commandDto.Name))
	if command.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&command, &commandDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := commandRepo.Create(&command); err != nil {
		return err
	}
	return nil
}

func (u *CommandService) Delete(ids []uint) error {
	if len(ids) == 1 {
		command, _ := commandRepo.Get(commonRepo.WithByID(ids[0]))
		if command.ID == 0 {
			return constant.ErrRecordNotFound
		}
		return commandRepo.Delete(commonRepo.WithByID(ids[0]))
	}
	return commandRepo.Delete(commonRepo.WithByIDs(ids))
}

func (u *CommandService) Update(id uint, upMap map[string]interface{}) error {
	return commandRepo.Update(id, upMap)
}
