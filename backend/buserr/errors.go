package buserr

import (
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/pkg/errors"
)

type BusinessError struct {
	Msg    string
	Detail interface{}
	Err    error
}

func (e BusinessError) Error() string {
	content := ""
	if e.Detail != nil {
		content = i18n.GetErrMsg(e.Msg, map[string]interface{}{"detail": e.Detail})
	} else {
		content = i18n.GetErrMsg(e.Msg, nil)
	}
	if content == "" {
		if e.Err != nil {
			return e.Err.Error()
		}
		return errors.New(e.Msg).Error()
	}
	return content
}

func New(Key string) BusinessError {
	return BusinessError{
		Msg:    Key,
		Detail: nil,
		Err:    nil,
	}
}

func WithMessage(Key string, detail interface{}, err error) BusinessError {
	return BusinessError{
		Msg:    Key,
		Detail: detail,
		Err:    err,
	}
}
