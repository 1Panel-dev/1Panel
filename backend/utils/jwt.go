package utils

import (
	"1Panel/app/dto"
	"1Panel/constant/errres"
	"1Panel/global"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWT struct {
	SigningKey []byte
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(global.Config.JWT.SigningKey),
	}
}

func (j *JWT) CreateToken(request dto.JwtRequest) (string, error) {
	request.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    global.Config.JWT.Issuer,
		NotBefore: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, &request)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenStr string) (*dto.JwtRequest, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &dto.JwtRequest{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errres.TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errres.TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errres.TokenNotValidYet
			} else {
				return nil, errres.TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*dto.JwtRequest); ok && token.Valid {
			return claims, nil
		}
		return nil, errres.TokenInvalid

	} else {
		return nil, errres.TokenInvalid
	}
}
