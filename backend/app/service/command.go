package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type CommandService struct{}

type ICommandService interface {
	List() ([]dto.CommandInfo, error)
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Create(commandDto dto.CommandOperate) error
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (u *CommandService) List() ([]dto.CommandInfo, error) {
	commands, err := commandRepo.GetList()
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

func (u *CommandService) SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, commands, err := commandRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Info))
	var dtoCommands []dto.CommandInfo
	for _, command := range commands {
		var item dto.CommandInfo
		if err := copier.Copy(&item, &command); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
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
