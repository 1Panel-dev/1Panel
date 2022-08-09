package jwt

import (
	"time"

	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"

	"github.com/golang-jwt/jwt/v4"
)

type JWT struct {
	SigningKey []byte
}

type JwtRequest struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}

type CustomClaims struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}

type BaseClaims struct {
	ID   uint
	Name string
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(global.CONF.JWT.SigningKey),
	}
}

func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	claims := CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: global.CONF.JWT.BufferTime,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(global.CONF.JWT.ExpiresTime))),
			Issuer:    global.CONF.JWT.Issuer,
		},
	}
	return claims
}

func (j *JWT) CreateToken(request CustomClaims) (string, error) {
	request.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    global.CONF.JWT.Issuer,
		NotBefore: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(global.CONF.JWT.ExpiresTime))),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, &request)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenStr string) (*JwtRequest, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JwtRequest{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil || token == nil {
		return nil, constant.ErrTokenParse
	}
	if claims, ok := token.Claims.(*JwtRequest); ok && token.Valid {
		return claims, nil
	}
	return nil, constant.ErrTokenParse
}
