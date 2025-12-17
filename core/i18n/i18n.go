package i18n

import (
	"embed"
	"fmt"
	"os/exec"
	"strings"
	"sync/atomic"

	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const defaultLang = "en"

var langFiles = map[string]string{
	"zh":      "lang/zh.yaml",
	"en":      "lang/en.yaml",
	"zh-Hant": "lang/zh-Hant.yaml",
	"pt-BR":   "lang/pt-BR.yaml",
	"ja":      "lang/ja.yaml",
	"ru":      "lang/ru.yaml",
	"ms":      "lang/ms.yaml",
	"ko":      "lang/ko.yaml",
	"tr":      "lang/tr.yaml",
	"es-ES":   "lang/es-ES.yaml",
}

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

func Get(key string) string {
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	if content != "" {
		return content
	}
	return key
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
			lang = GetLanguage()
		}
		global.I18n = i18n.NewLocalizer(bundle, lang)
	}
}

func Init() {
	if bundle == nil {
		initBundle()
	}
	dbLang := getLanguageFromDBInternal()
	if dbLang == "" {
		dbLang = defaultLang
	}
	SetCachedDBLanguage(dbLang)

	global.I18n = i18n.NewLocalizer(bundle, dbLang)
}

func UseI18nForCmd(lang string) {
	if bundle == nil {
		initBundle()
	}
	if lang == "" {
		langFrom1pctl := getLanguageFrom1pctl()
		if langFrom1pctl == "" {
			lang = defaultLang
		} else {
			lang = langFrom1pctl
		}
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

func getLanguageFromDBInternal() string {
	if global.DB == nil {
		return defaultLang
	}
	lang, _ := repo.NewISettingRepo().GetValueByKey("Language")
	if lang == "" {
		return defaultLang
	}
	return lang
}
func getLanguageFrom1pctl() string {
	cmd := exec.Command("bash", "-c", "grep '^LANGUAGE=' /usr/local/bin/1pctl | cut -d'=' -f2")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	info := strings.ReplaceAll(string(stdout), "\n", "")
	if len(info) == 0 || info == `""` {
		panic("error `LANGUAGE` find in /usr/local/bin/1pctl")
	}
	return info
}

var cachedDBLang atomic.Value

func GetLanguage() string {
	if v := cachedDBLang.Load(); v != nil {
		return v.(string)
	}
	return defaultLang
}

func SetCachedDBLanguage(lang string) {
	if lang == "" {
		lang = defaultLang
	}
	cachedDBLang.Store(lang)
}

func initBundle() {
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	isSuccess := true
	for _, file := range langFiles {
		if _, err := bundle.LoadMessageFileFS(fs, file); err != nil {
			fmt.Printf("[i18n] load language file %s failed: %v\n", file, err)
			isSuccess = false
		}
	}

	if !isSuccess {
		panic("[i18n] failed to init language files, See log above for details")
	}
}
