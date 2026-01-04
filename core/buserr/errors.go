package buserr

import (
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/pkg/errors"
)

type BusinessError struct {
	Msg    string
	Detail interface{}
	Map    map[string]interface{}
	Err    error
}

func (e BusinessError) Error() string {
	content := ""
	if e.Detail != nil {
		content = i18n.GetErrMsg(e.Msg, map[string]interface{}{"detail": e.Detail})
	} else if e.Map != nil {
		content = i18n.GetErrMsg(e.Msg, e.Map)
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

func New(key string, opts ...Option) BusinessError {
	be := BusinessError{
		Msg: key,
	}

	for _, opt := range opts {
		opt(&be)
	}

	return be
}

func WithErr(Key string, err error) BusinessError {
	paramMap := map[string]interface{}{}
	if err != nil {
		paramMap["err"] = err
	}
	return BusinessError{
		Msg: Key,
		Map: paramMap,
		Err: err,
	}
}

func WithDetail(Key string, detail interface{}, err error) BusinessError {
	return BusinessError{
		Msg:    Key,
		Detail: detail,
		Err:    err,
	}
}

func WithMap(Key string, maps map[string]interface{}, err error) BusinessError {
	return BusinessError{
		Msg: Key,
		Map: maps,
		Err: err,
	}
}

func WithName(Key string, name string) BusinessError {
	paramMap := map[string]interface{}{}
	if name != "" {
		paramMap["name"] = name
	}
	return BusinessError{
		Msg: Key,
		Map: paramMap,
	}
}

type Option func(*BusinessError)

func WithNameOption(name string) Option {
	return func(be *BusinessError) {
		if name != "" {
			if be.Map == nil {
				be.Map = make(map[string]interface{})
			}
			be.Map["name"] = name
		}
	}
}

func WithErrOption(err error) Option {
	return func(be *BusinessError) {
		be.Err = err
		if err != nil {
			if be.Map == nil {
				be.Map = make(map[string]interface{})
			}
			be.Map["err"] = err
		}
	}
}
