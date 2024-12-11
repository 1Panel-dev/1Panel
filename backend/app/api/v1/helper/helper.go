package helper

import (
	"context"
	"fmt"
	"github.com/1Panel-dev/1Panel/cmd/server/res"
	"net/http"
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func ErrorWithDetail(ctx *gin.Context, code int, msgKey string, err error) {
	res := dto.Response{
		Code:    code,
		Message: "",
	}
	if msgKey == constant.ErrTypeInternalServer {
		switch {
		case errors.Is(err, constant.ErrRecordExist):
			res.Message = i18n.GetMsgWithMap("ErrRecordExist", nil)
		case errors.Is(constant.ErrRecordNotFound, err):
			res.Message = i18n.GetMsgWithMap("ErrRecordNotFound", nil)
		case errors.Is(constant.ErrInvalidParams, err):
			res.Message = i18n.GetMsgWithMap("ErrInvalidParams", nil)
		case errors.Is(constant.ErrStructTransform, err):
			res.Message = i18n.GetMsgWithMap("ErrStructTransform", map[string]interface{}{"detail": err})
		case errors.Is(constant.ErrCaptchaCode, err):
			res.Code = constant.CodeAuth
			res.Message = "ErrCaptchaCode"
		case errors.Is(constant.ErrAuth, err):
			res.Code = constant.CodeAuth
			res.Message = "ErrAuth"
		case errors.Is(constant.ErrInitialPassword, err):
			res.Message = i18n.GetMsgWithMap("ErrInitialPassword", map[string]interface{}{"detail": err})
		case errors.As(err, &buserr.BusinessError{}):
			res.Message = err.Error()
		default:
			res.Message = i18n.GetMsgWithMap(msgKey, map[string]interface{}{"detail": err})
		}
	} else {
		res.Message = i18n.GetMsgWithMap(msgKey, map[string]interface{}{"detail": err})
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func SuccessWithData(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := dto.Response{
		Code: constant.CodeSuccess,
		Data: data,
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func SuccessWithOutData(ctx *gin.Context) {
	res := dto.Response{
		Code:    constant.CodeSuccess,
		Message: "success",
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func SuccessWithMsg(ctx *gin.Context, msg string) {
	res := dto.Response{
		Code:    constant.CodeSuccess,
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
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return err
	}
	if err := global.VALID.Struct(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return err
	}
	return nil
}

func CheckBind(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return err
	}
	return nil
}

func ErrWithHtml(ctx *gin.Context, code int, scope string) {
	if code == 444 {
		ctx.String(444, "")
		ctx.Abort()
		return
	}
	file := fmt.Sprintf("html/%d.html", code)
	if code == 200 && scope != "" {
		file = fmt.Sprintf("html/200_%s.html", scope)
	}
	data, err := res.ErrorMsg.ReadFile(file)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		ctx.Abort()
		return
	}
	ctx.Data(code, "text/html; charset=utf-8", data)
	ctx.Abort()
}
