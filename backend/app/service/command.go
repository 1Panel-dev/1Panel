package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type CommandService struct{}

type ICommandService interface {
	List() ([]dto.CommandInfo, error)
	SearchForTree() ([]dto.CommandTree, error)
	SearchWithPage(search dto.SearchCommandWithPage) (int64, interface{}, error)
	Create(commandDto dto.CommandOperate) error
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error

	SearchRedisCommandWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	ListRedisCommand() ([]dto.RedisCommand, error)
	SaveRedisCommand(commandDto dto.RedisCommand) error
	DeleteRedisCommand(ids []uint) error
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (u *CommandService) List() ([]dto.CommandInfo, error) {
	commands, err := commandRepo.GetList(commonRepo.WithOrderBy("name"))
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

func (u *CommandService) SearchForTree() ([]dto.CommandTree, error) {
	cmdList, err := commandRepo.GetList(commonRepo.WithOrderBy("name"))
	if err != nil {
		return nil, err
	}
	groups, err := groupRepo.GetList(commonRepo.WithByType("command"))
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

func (u *CommandService) SearchWithPage(search dto.SearchCommandWithPage) (int64, interface{}, error) {
	total, commands, err := commandRepo.Page(search.Page, search.PageSize, commandRepo.WithLikeName(search.Name), commonRepo.WithLikeName(search.Info), commonRepo.WithByGroupID(search.GroupID), commonRepo.WithOrderRuleBy(search.OrderBy, search.Order))
	if err != nil {
		return 0, nil, err
	}
	groups, _ := groupRepo.GetList(commonRepo.WithByType("command"), commonRepo.WithOrderBy("name"))
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
	return commandRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *CommandService) Update(id uint, upMap map[string]interface{}) error {
	return commandRepo.Update(id, upMap)
}

func (u *CommandService) SearchRedisCommandWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, commands, err := commandRepo.PageRedis(search.Page, search.PageSize, commandRepo.WithLikeName(search.Info))
	if err != nil {
		return 0, nil, err
	}
	var dtoCommands []dto.RedisCommand
	for _, command := range commands {
		var item dto.RedisCommand
		if err := copier.Copy(&item, &command); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoCommands = append(dtoCommands, item)
	}
	return total, dtoCommands, err
}

func (u *CommandService) ListRedisCommand() ([]dto.RedisCommand, error) {
	commands, err := commandRepo.GetRedisList()
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	var dtoCommands []dto.RedisCommand
	for _, command := range commands {
		var item dto.RedisCommand
		if err := copier.Copy(&item, &command); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoCommands = append(dtoCommands, item)
	}
	return dtoCommands, err
}

func (u *CommandService) SaveRedisCommand(req dto.RedisCommand) error {
	if req.ID == 0 {
		command, _ := commandRepo.GetRedis(commonRepo.WithByName(req.Name))
		if command.ID != 0 {
			return constant.ErrRecordExist
		}
	}
	var command model.RedisCommand
	if err := copier.Copy(&command, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := commandRepo.SaveRedis(&command); err != nil {
		return err
	}
	return nil
}

func (u *CommandService) DeleteRedisCommand(ids []uint) error {
	if len(ids) == 1 {
		command, _ := commandRepo.GetRedis(commonRepo.WithByID(ids[0]))
		if command.ID == 0 {
			return constant.ErrRecordNotFound
		}
		return commandRepo.DeleteRedis(commonRepo.WithByID(ids[0]))
	}
	return commandRepo.DeleteRedis(commonRepo.WithIdsIn(ids))
}
