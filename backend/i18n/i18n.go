package i18n

import (
	"embed"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

func GetMsg(msg string) string {
	content := ginI18n.MustGetMessage(msg)
	if content == "" {
		return msg
	} else {
		return content
	}
}

func GetMsgWithMap(msg string, maps map[string]interface{}) string {
	content := ginI18n.MustGetMessage(&i18n.LocalizeConfig{
		MessageID:    msg,
		TemplateData: maps,
	})
	if content == "" {
		return msg
	} else {
		return content
	}
}

//go:embed lang/*
var fs embed.FS

func GinI18nLocalize() gin.HandlerFunc {
	return ginI18n.Localize(ginI18n.WithBundle(&ginI18n.BundleCfg{
		RootPath:         "./lang",
		AcceptLanguage:   []language.Tag{language.Chinese, language.English},
		DefaultLanguage:  language.Chinese,
		FormatBundleFile: "toml",
		UnmarshalFunc:    toml.Unmarshal,
		Loader:           &ginI18n.EmbedLoader{FS: fs},
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
