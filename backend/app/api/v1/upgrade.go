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
		Tag:         string(*stats.TagName),
		ReleaseNote: string(*stats.Body),
		CreatedAt:   github.Timestamp(*stats.CreatedAt).Format("2006-01-02 15:04:05"),
		PublishedAt: github.Timestamp(*stats.PublishedAt).Format("2006-01-02 15:04:05"),
	}
	helper.SuccessWithData(c, info)
}
