package helper

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func ErrorWithDetail(ctx *gin.Context, code int, msgKey string, err error) {
	res := dto.Response{
		Code:    code,
		Message: "",
	}
	res.Message = i18n.GetMsgWithDetail(msgKey, err.Error())
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func InternalServer(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, http.StatusInternalServerError, "ErrInternalServer", err)
}

func BadRequest(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, http.StatusBadRequest, "ErrInvalidParams", err)
}

func SuccessWithData(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := dto.Response{
		Code: http.StatusOK,
		Data: data,
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func Success(ctx *gin.Context) {
	res := dto.Response{
		Code:    http.StatusOK,
		Message: "success",
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func SuccessWithMsg(ctx *gin.Context, msg string) {
	res := dto.Response{
		Code:    http.StatusOK,
		Message: msg,
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func GetParamID(c *gin.Context) (uint, error) {
	idParam, ok := c.Params.Get("id")
	if !ok {
		return 0, errors.New("error id in path")
	}
	intNum, _ := strconv.Atoi(idParam)
	return uint(intNum), nil
}

func GetIntParamByKey(c *gin.Context, key string) (uint, error) {
	idParam, ok := c.Params.Get(key)
	if !ok {
		return 0, fmt.Errorf("error %s in path", key)
	}
	intNum, _ := strconv.Atoi(idParam)
	return uint(intNum), nil
}

func GetStrParamByKey(c *gin.Context, key string) (string, error) {
	idParam, ok := c.Params.Get(key)
	if !ok {
		return "", fmt.Errorf("error %s in path", key)
	}
	return idParam, nil
}

func GetTxAndContext() (tx *gorm.DB, ctx context.Context) {
	tx = global.DB.Begin()
	ctx = context.WithValue(context.Background(), constant.DB, tx)
	return
}

func CheckBindAndValidate(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindJSON(req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", err)
		return err
	}
	if err := global.VALID.Struct(req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", err)
		return err
	}
	return nil
}

func CheckBind(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", err)
		return err
	}
	return nil
}

func GetParamInt32(paramName string, c *gin.Context) (int32, error) {
	idParam, ok := c.Params.Get(paramName)
	if !ok {
		return 0, errors.New("error id in path")
	}
	intNum, _ := strconv.Atoi(idParam)
	return int32(intNum), nil
}
