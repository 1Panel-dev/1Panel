package constant

import (
	"errors"
)

const (
	CodeSuccess           = 200
	CodeErrBadRequest     = 400
	CodeErrUnauthorized   = 401
	CodeErrForbidden      = 403
	CodeErrNotFound       = 404
	CodeErrInternalServer = 500
	CodeErrHeader         = 406
)

// internal
var (
	ErrCaptchaCode     = errors.New("ErrCaptchaCode")
	ErrRecordExist     = errors.New("ErrRecordExist")
	ErrRecordNotFound  = errors.New("ErrRecordNotFound")
	ErrStructTransform = errors.New("ErrStructTransform")

	ErrTokenParse = errors.New("ErrTokenParse")

	ErrPageGenerate = errors.New("generate page info failed")
)

// api
var (
	ErrTypeInternalServer = "ErrInternalServer"
	ErrTypeInvalidParams  = "ErrInvalidParams"
	ErrTypeToken          = "ErrToken"
	ErrTypeTokenTimeOut   = "ErrTokenTimeOut"
)
