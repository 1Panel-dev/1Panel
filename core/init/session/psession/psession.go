package psession

import (
	"encoding/json"
	"errors"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
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
		return result, errors.New("session data not found")
	}
	bytes, ok := data.([]byte)
	if !ok {
		return result, errors.New("invalid session data format")
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
	p.db.Debug().Table("sessions").Where("1=1").Delete(nil)
	return nil
}
