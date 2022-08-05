package dto

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/i18n"

	"github.com/gin-gonic/gin"
)

type PageResult struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
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

func NewError(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  i18n.GetMsg(msg),
		Data: gin.H{},
	}
}

func NewSuccess(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  i18n.GetMsg(msg),
		Data: gin.H{},
	}
}

func (r *Result) Success() {
	r.Ctx.JSON(http.StatusOK, map[string]interface{}{})
	r.Ctx.Abort()
}

func (r *Result) ErrorCode(code int, msg string) {
	res := Response{}
	res.Code = code
	res.Msg = i18n.GetMsg(msg)
	res.Data = gin.H{}
	r.Ctx.JSON(http.StatusOK, res)
	r.Ctx.Abort()
}

func (r *Result) Error(res Response) {
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
