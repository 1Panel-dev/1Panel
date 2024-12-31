package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags File
// @Summary List favorites
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /files/favorite/search [post]
func (b *BaseApi) SearchFavorite(c *gin.Context) {
	var req dto.PageInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := favoriteService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: list,
	})
}

// @Tags File
// @Summary Create favorite
// @Accept json
// @Param request body request.FavoriteCreate true "request"
// @Success 200 {object} model.Favorite
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /files/favorite [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"收藏文件/文件夹 [path]","formatEN":"收藏文件/文件夹 [path]"}
func (b *BaseApi) CreateFavorite(c *gin.Context) {
	var req request.FavoriteCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	favorite, err := favoriteService.Create(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, favorite)
}

// @Tags File
// @Summary Delete favorite
// @Accept json
// @Param request body request.FavoriteDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /files/favorite/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"favorites","output_column":"path","output_value":"path"}],"formatZH":"删除收藏 [path]","formatEN":"delete avorite [path]"}
func (b *BaseApi) DeleteFavorite(c *gin.Context) {
	var req request.FavoriteDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := favoriteService.Delete(req.ID); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
