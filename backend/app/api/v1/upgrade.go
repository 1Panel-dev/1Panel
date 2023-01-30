package v1

import (
	"context"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
)

// @Tags System Setting
// @Summary Load upgrade info
// @Description 加载系统更新信息
// @Success 200 {object} dto.UpgradeInfo
// @Security ApiKeyAuth
// @Router /settings/upgrade [get]
func (b *BaseApi) GetUpgradeInfo(c *gin.Context) {
	client := github.NewClient(nil)
	stats, _, err := client.Repositories.GetLatestRelease(context.Background(), "KubeOperator", "KubeOperator")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	info := dto.UpgradeInfo{
		NewVersion:  string(*stats.Name),
		ReleaseNote: string(*stats.Body),
		CreatedAt:   github.Timestamp(*stats.CreatedAt).Format("2006-01-02 15:04:05"),
	}
	helper.SuccessWithData(c, info)
}

// @Tags System Setting
// @Summary Load upgrade info
// @Description 从 OSS 加载系统更新信息
// @Success 200 {object} dto.UpgradeInfo
// @Security ApiKeyAuth
// @Router /settings/upgrade [get]
func (b *BaseApi) GetUpgradeInfoByOSS(c *gin.Context) {
	info, err := upgradeService.SearchUpgrade()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags System Setting
// @Summary Upgrade
// @Description 系统更新
// @Accept json
// @Param request body dto.Upgrade true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/upgrade [post]
// @x-panel-log {"bodyKeys":["version"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"更新系统 => [version]","formatEN":"upgrade service => [version]"}
func (b *BaseApi) Upgrade(c *gin.Context) {
	var req dto.Upgrade
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := upgradeService.Upgrade(req.Version); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
