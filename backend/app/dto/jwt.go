package dto

import "github.com/golang-jwt/jwt/v4"

type JwtRequest struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}

type BaseClaims struct {
	Username string
}
