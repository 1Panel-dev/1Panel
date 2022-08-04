package result

import (
	"github.com/1Panel-dev/1Panel/app/i18n"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ResultCont struct {
	Code int         `json:"code"` //提示代码
	Msg  string      `json:"msg"`  //提示信息
	Data interface{} `json:"data"` //出错
}

type Result struct {
	Ctx *gin.Context
}

func NewResult(ctx *gin.Context) *Result {
	return &Result{Ctx: ctx}
}

func NewError(code int, msg string) ResultCont {
	return ResultCont{
		Code: code,
		Msg:  i18n.GetMsg(msg),
		Data: gin.H{},
	}
}

func (r *Result) Success(data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := ResultCont{}
	res.Code = 0
	res.Msg = ""
	res.Data = data
	r.Ctx.JSON(http.StatusOK, res)
}

func (r *Result) ErrorCode(code int, msg string) {
	res := ResultCont{}
	res.Code = code
	res.Msg = i18n.GetMsg(msg)
	res.Data = gin.H{}
	r.Ctx.JSON(http.StatusOK, res)
	r.Ctx.Abort()
}

func (r *Result) Error(res ResultCont) {
	r.Ctx.JSON(http.StatusOK, res)
	r.Ctx.Abort()
}
