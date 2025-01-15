package i18n

import (
	"embed"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/global"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

func GetMsgWithMap(key string, maps map[string]interface{}) string {
	var content string
	if maps == nil {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
	} else {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: maps,
		})
	}
	content = strings.ReplaceAll(content, ": <no value>", "")
	if content == "" {
		return key
	} else {
		return content
	}
}

func GetMsgWithName(key string, name string, err error) string {
	var (
		content string
		dataMap = make(map[string]interface{})
	)
	dataMap["name"] = name
	if err != nil {
		dataMap["err"] = err.Error()
	}
	content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: dataMap,
	})
	content = strings.ReplaceAll(content, "<no value>", "")
	if content == "" {
		return key
	} else {
		return content
	}
}

func GetErrMsg(key string, maps map[string]interface{}) string {
	var content string
	if maps == nil {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
	} else {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: maps,
		})
	}
	return content
}

func GetMsgByKey(key string) string {
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	return content
}

//go:embed lang/*
var fs embed.FS
var bundle *i18n.Bundle

func UseI18n() gin.HandlerFunc {
	return func(context *gin.Context) {
		lang := context.GetHeader("Accept-Language")
		if lang == "" {
			lang = "zh"
		}
		global.I18n = i18n.NewLocalizer(bundle, lang)
	}
}

func Init() {
	if bundle != nil {
		return
	}
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	_, _ = bundle.LoadMessageFileFS(fs, "lang/zh.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/en.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/zh-Hant.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/fa.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/pt.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/pt-BR.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/ja.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/ru.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/ms.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/ko.yaml")
}

func UseI18nForCmd(lang string) {
	if lang == "" {
		lang = "en"
	}

	if bundle == nil {
		Init()
	}
	global.I18nForCmd = i18n.NewLocalizer(bundle, lang)
}
func GetMsgByKeyForCmd(key string) string {
	if global.I18nForCmd == nil {
		UseI18nForCmd("")
	}
	content, _ := global.I18nForCmd.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	return content
}
func GetMsgWithMapForCmd(key string, maps map[string]interface{}) string {
	if global.I18nForCmd == nil {
		UseI18nForCmd("")
	}
	var content string
	if maps == nil {
		content, _ = global.I18nForCmd.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
	} else {
		content, _ = global.I18nForCmd.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: maps,
		})
	}
	content = strings.ReplaceAll(content, ": <no value>", "")
	if content == "" {
		return key
	} else {
		return content
	}
}
