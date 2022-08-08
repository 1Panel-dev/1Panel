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

var (
	ErrTypeToken        = "ErrToken"
	ErrTypeTokenExpired = "ErrTokenExpired"

	ErrTypeParamInReqBody  = "ErrParamInReqBody"
	ErrTypeParamInReqQuery = "ErrParamInReqQuery"
	ErrTypeInternalServer  = "ErrInternalServer"
	ErrTypeParamValid      = "ErrParamValid"
)

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")

	ErrCaptchaCode   = errors.New("captcha code error")
	ErrPageParam     = errors.New("paging parameter error")
	ErrRecordExist   = errors.New("record already exists")
	ErrCopyTransform = errors.New("type conversion failure")
)
