package errres

import (
	"errors"

	"github.com/1Panel-dev/1Panel/app/dto"
)

const (
	Success            = 0
	Error              = 500
	InvalidParams      = 400
	InvalidCommon      = 10000
	InvalidJwtExpired  = 10001
	InvalidJwtNotFound = 10002
)

var (
	OK           = dto.NewError(Success, "Ok")
	InvalidParam = dto.NewError(InvalidParams, "InvalidParams")
	JwtExpired   = dto.NewError(InvalidJwtExpired, "JwtExpired")
	JwtNotFound  = dto.NewError(InvalidJwtNotFound, "JwtNotFound")
)

var (
	TokenExpired     = errors.New("token is expired")
	TokenNotValidYet = errors.New("token not active yet")
	TokenMalformed   = errors.New("that's not even a token")
	TokenInvalid     = errors.New("couldn't handle this token")
)
