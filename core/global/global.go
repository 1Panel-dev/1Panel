package global

import (
	"github.com/1Panel-dev/1Panel/core/configs"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	DB      *gorm.DB
	TaskDB  *gorm.DB
	LOG     *logrus.Logger
	CONF    configs.ServerConfig
	VALID   *validator.Validate
	SESSION *psession.PSession
	Viper   *viper.Viper

	I18n *i18n.Localizer

	Cron *cron.Cron
)

type DBOption func(*gorm.DB) *gorm.DB
