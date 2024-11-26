package global

import (
	"github.com/1Panel-dev/1Panel/agent/configs"
	badger_db "github.com/1Panel-dev/1Panel/agent/init/cache/db"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	MonitorDB *gorm.DB
	TaskDB    *gorm.DB
	CoreDB    *gorm.DB
	LOG       *logrus.Logger
	CONF      configs.ServerConfig
	VALID     *validator.Validate
	CACHE     *badger_db.Cache
	Viper     *viper.Viper

	Cron           *cron.Cron
	MonitorCronID  cron.EntryID
	OneDriveCronID cron.EntryID

	IsMaster bool

	I18n *i18n.Localizer
)
