package middlerware

import (
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

func GinI18nLocalize() gin.HandlerFunc {
	return ginI18n.Localize(ginI18n.WithBundle(&ginI18n.BundleCfg{
		RootPath:         "./backend/lang",
		AcceptLanguage:   []language.Tag{language.Chinese, language.English},
		DefaultLanguage:  language.Chinese,
		FormatBundleFile: "toml",
		UnmarshalFunc:    toml.Unmarshal,
	}),
		ginI18n.WithGetLngHandle(
			func(context *gin.Context, defaultLng string) string {
				lng := context.GetHeader("Accept-Language")
				if lng == "" {
					return defaultLng
				}
				return lng
			},
		))
}
