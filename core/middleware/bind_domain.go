package middleware

import (
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/utils/security"
	"github.com/gin-gonic/gin"
)

// DomainPolicy is configured by the router and must only trust server-verified state.
type DomainPolicy func(*gin.Context) bool

func BindDomain(policies ...DomainPolicy) gin.HandlerFunc {
	policies = append([]DomainPolicy(nil), policies...)
	return func(c *gin.Context) {
		for _, policy := range policies {
			if policy != nil && policy(c) {
				c.Next()
				return
			}
		}
		localRequest := c.GetBool("LOCAL_REQUEST")
		if localRequest {
			c.Next()
			return
		}
		settingRepo := repo.NewISettingRepo()
		bindDomain, err := settingRepo.GetValueByKey("BindDomain")
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		if len(bindDomain) == 0 {
			c.Next()
			return
		}
		domains := c.Request.Host
		parts := strings.Split(c.Request.Host, ":")
		if len(parts) > 0 {
			domains = parts[0]
		}

		if domains != bindDomain {
			code := security.LoadErrCode()
			helper.ErrWithHtml(c, code, "err_domain")
			return
		}
		c.Next()
	}
}
