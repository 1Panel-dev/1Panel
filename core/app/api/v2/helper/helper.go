package helper

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
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
		case errors.Is(constant.ErrTransform, err):
			res.Message = i18n.GetMsgWithMap("ErrTransform", map[string]interface{}{"detail": err})
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

func InternalServer(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
}

func BadRequest(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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

func ErrResponse(ctx *gin.Context, code int) {
	ctx.JSON(code, nil)
	ctx.Abort()
}
