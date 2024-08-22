package i18n

import (
	"embed"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"

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

func GetMsgWithDetail(key string, detail string) string {
	var (
		content string
		dataMap = make(map[string]interface{})
	)
	dataMap["detail"] = detail
	content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: dataMap,
	})
	if content != "" {
		return content
	}
	return key
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

func GetWithName(key string, name string) string {
	var (
		dataMap = make(map[string]interface{})
	)
	dataMap["name"] = name
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: dataMap,
	})
	return content
}

func GetWithNameAndErr(key string, name string, err error) string {
	var (
		dataMap = make(map[string]interface{})
	)
	dataMap["name"] = name
	dataMap["err"] = err.Error()
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: dataMap,
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
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	_, _ = bundle.LoadMessageFileFS(fs, "lang/zh.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/en.yaml")
	_, _ = bundle.LoadMessageFileFS(fs, "lang/zh-Hant.yaml")
}
