package jwt

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"

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
	settingRepo := repo.NewISettingRepo()
	jwtSign, _ := settingRepo.Get(settingRepo.WithByKey("JWTSigningKey"))
	return &JWT{
		[]byte(jwtSign.Value),
	}
}

func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	claims := CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: constant.JWTBufferTime,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(constant.JWTBufferTime))),
			Issuer:    constant.JWTIssuer,
		},
	}
	return claims
}

func (j *JWT) CreateToken(request CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &request)
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
