package helper

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/i18n"
	"github.com/gin-gonic/gin"
)

func GeneratePaginationFromReq(c *gin.Context) (dto.PageInfo, bool) {
	p, ok1 := c.GetQuery("page")
	ps, ok2 := c.GetQuery("pageSize")
	if !(ok1 && ok2) {
		return dto.PageInfo{Page: 1, PageSize: 10}, false
	}

	page, err := strconv.Atoi(p)
	if err != nil {
		return dto.PageInfo{Page: 1, PageSize: 10}, false
	}
	pageSize, err := strconv.Atoi(ps)
	if err != nil {
		return dto.PageInfo{Page: 1, PageSize: 10}, false
	}

	return dto.PageInfo{Page: page, PageSize: pageSize}, false
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
