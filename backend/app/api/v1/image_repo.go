package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags Container Image-repo
// @Summary Page image repos
// @Description 获取镜像仓库列表分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Produce json
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /containers/repo/search [post]
func (b *BaseApi) SearchRepo(c *gin.Context) {
	var req dto.SearchWithPage
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	total, list, err := imageRepoService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Container Image-repo
// @Summary List image repos
// @Description 获取镜像仓库列表
// @Produce json
// @Success 200 {array} dto.ImageRepoOption
// @Security ApiKeyAuth
// @Router /containers/repo [get]
func (b *BaseApi) ListRepo(c *gin.Context) {
	list, err := imageRepoService.List()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Container Image-repo
// @Summary Load repo status
// @Description 获取 docker 仓库状态
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/repo/status [get]
func (b *BaseApi) CheckRepoStatus(c *gin.Context) {
	var req dto.OperateByID
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := imageRepoService.Login(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Image-repo
// @Summary Create image repo
// @Description 创建镜像仓库
// @Accept json
// @Param request body dto.ImageRepoDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/repo [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"创建镜像仓库 [name]","formatEN":"create image repo [name]"}
func (b *BaseApi) CreateRepo(c *gin.Context) {
	var req dto.ImageRepoCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := imageRepoService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Image-repo
// @Summary Delete image repo
// @Description 删除镜像仓库
// @Accept json
// @Param request body dto.ImageRepoDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/repo/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"ids","isList":true,"db":"image_repos","output_column":"name","output_value":"names"}],"formatZH":"删除镜像仓库 [names]","formatEN":"delete image repo [names]"}
func (b *BaseApi) DeleteRepo(c *gin.Context) {
	var req dto.ImageRepoDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := imageRepoService.BatchDelete(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Image-repo
// @Summary Update image repo
// @Description 更新镜像仓库
// @Accept json
// @Param request body dto.ImageRepoUpdate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/repo/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFuntions":[{"input_column":"id","input_value":"id","isList":false,"db":"image_repos","output_column":"name","output_value":"name"}],"formatZH":"更新镜像仓库 [name]","formatEN":"update image repo information [name]"}
func (b *BaseApi) UpdateRepo(c *gin.Context) {
	var req dto.ImageRepoUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := imageRepoService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
