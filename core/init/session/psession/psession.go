package psession

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SessionUser struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type PSession struct {
	Store *gormstore.Store
	db    *gorm.DB
}

func NewPSession(dbPath string) *PSession {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newLogger,
	})
	if err != nil {
		panic(err)
	}
	sqlDB, dbError := db.DB()
	if dbError != nil {
		panic(dbError)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	store := gormstore.New(db, securecookie.GenerateRandomKey(32))
	return &PSession{
		Store: store,
		db:    db,
	}
}

func (p *PSession) Get(c *gin.Context) (SessionUser, error) {
	var result SessionUser
	session, err := p.Store.Get(c.Request, constant.SessionName)
	if err != nil {
		return result, err
	}
	data, ok := session.Values["user"]
	if !ok {
		return result, errors.New("ErrSessionDataNotFound")
	}
	bytes, ok := data.([]byte)
	if !ok {
		return result, errors.New("ErrSessionDataFormat")
	}
	err = json.Unmarshal(bytes, &result)
	return result, err
}

func (p *PSession) Set(c *gin.Context, user SessionUser, secure bool, ttlSeconds int) error {
	session, err := p.Store.Get(c.Request, constant.SessionName)
	if err != nil {
		return err
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	session.Values["user"] = data
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   ttlSeconds,
		HttpOnly: true,
		Secure:   secure,
	}
	return p.Store.Save(c.Request, c.Writer, session)
}

func (p *PSession) Delete(c *gin.Context) error {
	session, err := p.Store.Get(c.Request, constant.SessionName)
	if err != nil {
		return err
	}

	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	return p.Store.Save(c.Request, c.Writer, session)
}

func (p *PSession) Clean() error {
	p.db.Table("sessions").Where("1=1").Delete(nil)
	return nil
}
