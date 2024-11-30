package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Container Image
// @Summary Page images
// @Description 获取镜像列表分页
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Produce json
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /containers/image/search [post]
func (b *BaseApi) SearchImage(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := imageService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Container Image
// @Summary List all images
// @Description 获取所有镜像列表
// @Produce json
// @Success 200 {array} dto.ImageInfo
// @Security ApiKeyAuth
// @Router /containers/image/all [get]
func (b *BaseApi) ListAllImage(c *gin.Context) {
	list, err := imageService.ListAll()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Container Image
// @Summary load images options
// @Description 获取镜像名称列表
// @Produce json
// @Success 200 {array} dto.Options
// @Security ApiKeyAuth
// @Router /containers/image [get]
func (b *BaseApi) ListImage(c *gin.Context) {
	list, err := imageService.List()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Container Image
// @Summary Build image
// @Description 构建镜像
// @Accept json
// @Param request body dto.ImageBuild true "request"
// @Success 200 {string} log
// @Security ApiKeyAuth
// @Router /containers/image/build [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"构建镜像 [name]","formatEN":"build image [name]"}
func (b *BaseApi) ImageBuild(c *gin.Context) {
	var req dto.ImageBuild
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	log, err := imageService.ImageBuild(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, log)
}

// @Tags Container Image
// @Summary Pull image
// @Description 拉取镜像
// @Accept json
// @Param request body dto.ImagePull true "request"
// @Success 200 {string} log
// @Security ApiKeyAuth
// @Router /containers/image/pull [post]
// @x-panel-log {"bodyKeys":["repoID","imageName"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"repoID","isList":false,"db":"image_repos","output_column":"name","output_value":"reponame"}],"formatZH":"镜像拉取 [reponame][imageName]","formatEN":"image pull [reponame][imageName]"}
func (b *BaseApi) ImagePull(c *gin.Context) {
	var req dto.ImagePull
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	logPath, err := imageService.ImagePull(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, logPath)
}

// @Tags Container Image
// @Summary Push image
// @Description 推送镜像
// @Accept json
// @Param request body dto.ImagePush true "request"
// @Success 200 {string} log
// @Security ApiKeyAuth
// @Router /containers/image/push [post]
// @x-panel-log {"bodyKeys":["repoID","tagName","name"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"repoID","isList":false,"db":"image_repos","output_column":"name","output_value":"reponame"}],"formatZH":"[tagName] 推送到 [reponame][name]","formatEN":"push [tagName] to [reponame][name]"}
func (b *BaseApi) ImagePush(c *gin.Context) {
	var req dto.ImagePush
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	logPath, err := imageService.ImagePush(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, logPath)
}

// @Tags Container Image
// @Summary Delete image
// @Description 删除镜像
// @Accept json
// @Param request body dto.BatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/image/remove [post]
// @x-panel-log {"bodyKeys":["names"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"移除镜像 [names]","formatEN":"remove image [names]"}
func (b *BaseApi) ImageRemove(c *gin.Context) {
	var req dto.BatchDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := imageService.ImageRemove(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Image
// @Summary Save image
// @Description 导出镜像
// @Accept json
// @Param request body dto.ImageSave true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/image/save [post]
// @x-panel-log {"bodyKeys":["tagName","path","name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"保留 [tagName] 为 [path]/[name]","formatEN":"save [tagName] as [path]/[name]"}
func (b *BaseApi) ImageSave(c *gin.Context) {
	var req dto.ImageSave
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := imageService.ImageSave(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Image
// @Summary Tag image
// @Description Tag 镜像
// @Accept json
// @Param request body dto.ImageTag true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/image/tag [post]
// @x-panel-log {"bodyKeys":["repoID","targetName"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"repoID","isList":false,"db":"image_repos","output_column":"name","output_value":"reponame"}],"formatZH":"tag 镜像 [reponame][targetName]","formatEN":"tag image [reponame][targetName]"}
func (b *BaseApi) ImageTag(c *gin.Context) {
	var req dto.ImageTag
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := imageService.ImageTag(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Container Image
// @Summary Load image
// @Description 导入镜像
// @Accept json
// @Param request body dto.ImageLoad true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /containers/image/load [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"从 [path] 加载镜像","formatEN":"load image from [path]"}
func (b *BaseApi) ImageLoad(c *gin.Context) {
	var req dto.ImageLoad
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := imageService.ImageLoad(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
