package global

import (
	"1Panel/configs"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	Logger    *logrus.Logger
	Config    configs.ServerConfig
	Validator *validator.Validate
)
