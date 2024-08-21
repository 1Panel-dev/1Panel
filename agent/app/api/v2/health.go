package v2

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) CheckHealth(c *gin.Context) {
	_, err := xpack.RequestToMaster("/api/v2/agent/health", http.MethodGet, nil)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
