package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Container Image-repo
// @Summary Page image repos
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Produce json
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/repo/search [post]
func (b *BaseApi) SearchRepo(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Produce json
// @Success 200 {array} dto.ImageRepoOption
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/repo/status [get]
func (b *BaseApi) CheckRepoStatus(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Accept json
// @Param request body dto.ImageRepoDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/repo [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建镜像仓库 [name]","formatEN":"create image repo [name]"}
func (b *BaseApi) CreateRepo(c *gin.Context) {
	var req dto.ImageRepoCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Accept json
// @Param request body dto.ImageRepoDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/repo/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"image_repos","output_column":"name","output_value":"names"}],"formatZH":"删除镜像仓库 [names]","formatEN":"delete image repo [names]"}
func (b *BaseApi) DeleteRepo(c *gin.Context) {
	var req dto.ImageRepoDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Accept json
// @Param request body dto.ImageRepoUpdate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/repo/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"image_repos","output_column":"name","output_value":"name"}],"formatZH":"更新镜像仓库 [name]","formatEN":"update image repo information [name]"}
func (b *BaseApi) UpdateRepo(c *gin.Context) {
	var req dto.ImageRepoUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := imageRepoService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
