package errres

import (
	"1Panel/app/dto"
	"errors"
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
	OK           = dto.NewSuccess(Success, "Ok")
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
