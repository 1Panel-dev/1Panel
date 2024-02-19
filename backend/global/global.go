package global

import (
	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/init/cache/badger_db"
	"github.com/1Panel-dev/1Panel/backend/init/session/psession"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	DB      *gorm.DB
	LOG     *logrus.Logger
	CONF    configs.ServerConfig
	VALID   *validator.Validate
	SESSION *psession.PSession
	CACHE   *badger_db.Cache
	Viper   *viper.Viper

	Cron           *cron.Cron
	MonitorCronID  cron.EntryID
	OneDriveCronID cron.EntryID

	I18n *i18n.Localizer
)
