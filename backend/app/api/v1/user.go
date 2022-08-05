package v1

import (
	"github.com/1Panel-dev/1Panel/constant/errres"
	"strconv"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/global"

	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

func (b *BaseApi) Login(c *gin.Context) {

}

func (b *BaseApi) Register(c *gin.Context) {
	var req dto.UserCreate
	_ = c.ShouldBindJSON(&req)
	res := dto.NewResult(c)

	if err := global.Validator.Struct(req); err != nil {
		res.ErrorWithDetail(errres.InvalidParam, err.Error())
		return
	}
	if err := userService.Register(req); err != nil {
		dto.NewResult(c).ErrorCode(500, err.Error())
		return
	}
	res.Success()
}

func (b *BaseApi) GetUserList(c *gin.Context) {
	// 这里到时候一起拦截一下
	p, ok1 := c.GetQuery("page")
	ps, ok2 := c.GetQuery("pageSize")
	res := dto.NewResult(c)

	if !(ok1 && ok2) {
		res.Error(errres.InvalidParam)
		return
	}
	page, err := strconv.Atoi(p)
	if err != nil {
		global.Logger.Error("获取失败!", err)
		dto.NewResult(c).ErrorCode(500, err.Error())
		return
	}
	pageSize, err := strconv.Atoi(ps)
	if err != nil {
		global.Logger.Error("获取失败!", err)
		dto.NewResult(c).ErrorCode(500, err.Error())
		return
	}

	total, list, err := userService.Page(page, pageSize)
	if err != nil {
		global.Logger.Error("获取失败!", err)
		dto.NewResult(c).ErrorCode(500, err.Error())
		return
	}
	res.SuccessWithData(dto.PageResult{
		Items: list,
		Total: total,
	})
}

func (b *BaseApi) DeleteUser(c *gin.Context) {
	var req dto.OperationWithName
	_ = c.ShouldBindJSON(&req)
	res := dto.NewResult(c)

	if err := global.Validator.Struct(req); err != nil {
		res.Error(errres.InvalidParam)
		return
	}
	if err := userService.Delete(req.Name); err != nil {
		global.Logger.Error("删除失败!", err)
		dto.NewResult(c).ErrorCode(500, err.Error())
		return
	}
	dto.NewResult(c).Success()
}

func (b *BaseApi) UpdateUser(c *gin.Context) {
	var req dto.UserUpdate
	_ = c.ShouldBindJSON(&req)
	res := dto.NewResult(c)

	if err := global.Validator.Struct(req); err != nil {
		res.Error(errres.InvalidParam)
		return
	}
	upMap := make(map[string]interface{})
	upMap["email"] = req.Email
	if err := userService.Update(upMap); err != nil {
		global.Logger.Error("更新失败!", err)
		dto.NewResult(c).ErrorCode(500, err.Error())
		return
	}
	dto.NewResult(c).Success()
}

func (b *BaseApi) GetUserInfo(c *gin.Context) {
	name, ok := c.Params.Get("name")
	res := dto.NewResult(c)

	if !ok {
		res.Error(errres.InvalidParam)
		return
	}

	user, err := userService.Get(name)
	if err != nil {
		global.Logger.Error("更新失败!", err)
		dto.NewResult(c).ErrorCode(500, err.Error())
		return
	}
	dto.NewResult(c).SuccessWithData(user)
}
