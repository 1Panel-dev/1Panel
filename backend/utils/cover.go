package utils

import (
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

func ModelToDTO(dto, model interface{}) error {
	if err := copier.Copy(dto, model); err != nil {
		return errors.Wrap(err, "cover to dto err")
	}
	return nil
}

func DTOToModel(model, dto interface{}) error {
	if err := copier.Copy(dto, model); err != nil {
		return errors.Wrap(err, "cover to model err")
	}
	return nil
}
