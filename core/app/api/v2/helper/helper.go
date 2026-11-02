package helper

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/1Panel-dev/1Panel/core/cmd/server/res"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/gin-gonic/gin"
)

func ErrorWithDetail(ctx *gin.Context, code int, msgKey string, err error) {
	applyErrorMetadata(ctx, code, msgKey, err)
	res := dto.Response{
		Code:    code,
		Message: "",
	}
	if msgKey == "ErrCaptchaCode" || msgKey == "ErrAuth" || msgKey == "ErrLoginLocked" {
		res.Code = 401
		res.Message = msgKey
	}
	if msgKey == "ErrRBAC" {
		res.Code = 412
		if err != nil {
			res.Message = err.Error()
			ctx.JSON(http.StatusOK, res)
			ctx.Abort()
			return
		}
	}
	res.Message = i18n.GetMsgWithMap(msgKey, map[string]interface{}{"detail": err})
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func InternalServer(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, http.StatusInternalServerError, "ErrInternalServer", err)
}

func BadRequest(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, http.StatusBadRequest, "ErrInvalidParams", err)
}

func BadAuth(ctx *gin.Context, msgKey string, err error) {
	ErrorWithDetail(ctx, http.StatusUnauthorized, msgKey, err)
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

func GetParamID(c *gin.Context) (uint, error) {
	idParam, ok := c.Params.Get("id")
	if !ok {
		return 0, errors.New("error id in path")
	}
	intNum, _ := strconv.Atoi(idParam)
	return uint(intNum), nil
}

const errorMetadataKey = "error_response_metadata_policies"

type ErrorMetadataPolicy func(*gin.Context, int, string, error)

func WithErrorMetadata(policies ...ErrorMetadataPolicy) gin.HandlerFunc {
	configured := append([]ErrorMetadataPolicy(nil), policies...)
	return func(c *gin.Context) {
		c.Set(errorMetadataKey, configured)
		c.Next()
	}
}

func applyErrorMetadata(c *gin.Context, code int, messageKey string, err error) {
	value, _ := c.Get(errorMetadataKey)
	policies, _ := value.([]ErrorMetadataPolicy)
	for _, policy := range policies {
		if policy != nil {
			policy(c, code, messageKey, err)
		}
	}
}
