package helper

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/gin-gonic/gin"
)

func GeneratePaginationFromReq(c *gin.Context) (*dto.PageInfo, bool) {
	p, ok1 := c.GetQuery("page")
	ps, ok2 := c.GetQuery("pageSize")
	if !(ok1 && ok2) {
		return nil, false
	}

	page, err := strconv.Atoi(p)
	if err != nil {
		return nil, false
	}
	pageSize, err := strconv.Atoi(ps)
	if err != nil {
		return nil, false
	}

	return &dto.PageInfo{Page: page, PageSize: pageSize}, true
}

func ErrorWithDetail(ctx *gin.Context, code int, msgKey string, err error) {
	res := dto.Response{
		Code: code,
		Msg:  "",
	}
	if msgKey == constant.ErrTypeInternalServer {
		switch {
		case errors.Is(err, constant.ErrRecordExist):
			res.Msg = i18n.GetMsgWithMap("ErrRecordExist", map[string]interface{}{"detail": err})
		case errors.Is(constant.ErrRecordNotFound, err):
			res.Msg = i18n.GetMsgWithMap("ErrRecordNotFound", map[string]interface{}{"detail": err})
		case errors.Is(constant.ErrStructTransform, err):
			res.Msg = i18n.GetMsgWithMap("ErrStructTransform", map[string]interface{}{"detail": err})
		case errors.Is(constant.ErrCaptchaCode, err):
			res.Msg = i18n.GetMsgWithMap("ErrCaptchaCode", map[string]interface{}{"detail": err})
		case errors.Is(constant.ErrAuth, err):
			res.Msg = i18n.GetMsgWithMap("ErrAuth", map[string]interface{}{"detail": err})
		case errors.Is(constant.ErrInitialPassword, err):
			res.Msg = i18n.GetMsgWithMap("ErrInitialPassword", map[string]interface{}{"detail": err})
		case errors.As(err, &buserr.BusinessError{}):
			res.Msg = err.Error()
		default:
			res.Msg = i18n.GetMsgWithMap(msgKey, map[string]interface{}{"detail": err})
		}
	} else {
		res.Msg = i18n.GetMsgWithMap(msgKey, map[string]interface{}{"detail": err})
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

func SuccessWithMsg(ctx *gin.Context, msg string) {
	res := dto.Response{
		Code: constant.CodeSuccess,
		Msg:  msg,
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
