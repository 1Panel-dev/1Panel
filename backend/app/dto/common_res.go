package dto

import (
	"github.com/1Panel-dev/1Panel/i18n"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PageResult struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

type ResDef struct {
	Code  int
	MsgID string
}

type Response struct {
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

func NewError(code int, msg string) ResDef {
	return ResDef{
		Code:  code,
		MsgID: msg,
	}
}

func (r *Result) Success() {
	r.Ctx.JSON(http.StatusOK, map[string]interface{}{})
	r.Ctx.Abort()
}

func (r *Result) Error(re ResDef) {
	res := Response{
		Code: re.Code,
		Msg:  i18n.GetMsg(re.MsgID),
	}
	r.Ctx.JSON(http.StatusOK, res)
	r.Ctx.Abort()
}

func (r *Result) ErrorWithDetail(re ResDef, err string) {
	res := Response{
		Code: re.Code,
		Msg:  i18n.GetMsgWithMap(re.MsgID, map[string]interface{}{"detail": err}),
	}
	r.Ctx.JSON(http.StatusOK, res)
	r.Ctx.Abort()
}

func (r *Result) SuccessWithData(data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := Response{}
	res.Code = 0
	res.Msg = ""
	res.Data = data
	r.Ctx.JSON(http.StatusOK, res)
}

func (r *Result) ErrorCode(code int, msg string) {
	res := Response{}
	res.Code = code
	res.Msg = msg
	res.Data = gin.H{}
	r.Ctx.JSON(http.StatusOK, res)
	r.Ctx.Abort()
}
