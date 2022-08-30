package service

import (
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type CommandService struct{}

type ICommandService interface {
	Search() ([]model.Command, error)
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Create(commandDto dto.CommandCreate) error
	Update(id uint, upMap map[string]interface{}) error
	Delete(name string) error
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (u *CommandService) Search() ([]model.Command, error) {
	commands, err := commandRepo.GetList()
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	return commands, err
}

func (u *CommandService) SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, commands, err := commandRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Name))
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

func (u *CommandService) Create(commandDto dto.CommandCreate) error {
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

func (u *CommandService) Delete(name string) error {
	command, _ := commandRepo.Get(commonRepo.WithByName(name))
	if command.ID == 0 {
		return constant.ErrRecordNotFound
	}
	return commandRepo.Delete(commonRepo.WithByID(command.ID))
}

func (u *CommandService) Update(id uint, upMap map[string]interface{}) error {
	return commandRepo.Update(id, upMap)
}
