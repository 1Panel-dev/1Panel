package global

import (
	"github.com/1Panel-dev/1Panel/configs"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	DB      *gorm.DB
	LOG     *logrus.Logger
	CONF    configs.ServerConfig
	VALID   *validator.Validate
	SESSION *sessions.CookieStore
)
