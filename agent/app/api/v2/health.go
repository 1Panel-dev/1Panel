package v2

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) CheckHealth(c *gin.Context) {
	_, err := xpack.RequestToMaster("/api/v2/agent/xpack/health", http.MethodGet, nil)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}
