package global

import (
	"github.com/1Panel-dev/1Panel/configs"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	DB     *gorm.DB
	Logger *logrus.Logger
	Config configs.ServerConfig
)
