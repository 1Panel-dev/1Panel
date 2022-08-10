package session

import (
	"github.com/1Panel-dev/1Panel/global"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func Init() {
	cs := &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs([]byte(global.CONF.Session.SessionKey)),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: global.CONF.Session.ExpiresTime,
		},
	}
	global.SESSION = cs
}

type SessionUser struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
