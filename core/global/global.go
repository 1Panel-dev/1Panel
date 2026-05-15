package global

import (
	"github.com/1Panel-dev/1Panel/core/init/auth"
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
	AlertDB *gorm.DB
	TaskDB  *gorm.DB
	AgentDB *gorm.DB
	LOG     *logrus.Logger
	CONF    ServerConfig
	VALID   *validator.Validate
	SESSION *psession.PSession
	Viper   *viper.Viper

	I18n       *i18n.Localizer
	I18nForCmd *i18n.Localizer

	Cron *cron.Cron

	ScriptSyncJobID cron.EntryID

	IPTracker *auth.IPTracker
)

type DBOption func(*gorm.DB) *gorm.DB

func RepoURL() string {
	if CONF.Base.IsEnterprise {
		return "https://resource.fit2cloud.com/1panel/package/enterprise"
	}
	if CONF.Base.IsFxplay {
		return "https://resource.fit2cloud.com/1panel/package/fusionxplay"
	}
	if CONF.Base.Edition != "intl" {
		return "https://resource.fit2cloud.com/1panel/package/v2"
	}
	return "https://resource.1panel.pro/v2"
}
func ResourceURL() string {
	if CONF.Base.IsEnterprise {
		return "https://resource.fit2cloud.com/1panel/resource/v2"
	}
	if CONF.Base.IsFxplay {
		return "https://resource.fit2cloud.com/1panel/resource/fusionxplay"
	}
	if CONF.Base.Edition != "intl" {
		return "https://resource.fit2cloud.com/1panel/resource/v2"
	}
	return "https://resource.1panel.pro/v2/resource"
}
func AppRepoURL() string {
	if CONF.Base.IsEnterprise {
		return "https://apps-assets.fit2cloud.com"
	}
	if CONF.Base.IsFxplay {
		return "https://apps-assets.fit2cloud.com"
	}
	if CONF.Base.Edition != "intl" {
		return "https://apps-assets.fit2cloud.com"
	}
	return "https://apps.1panel.pro"
}
