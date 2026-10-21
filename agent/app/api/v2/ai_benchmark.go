package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) AIBenchmarkSources(c *gin.Context) {
	sources, err := service.AIBenchmarkSources(c.Query("kind"), c.Query("selected"))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, sources)
}
