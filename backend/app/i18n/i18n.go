package i18n

import ginI18n "github.com/gin-contrib/i18n"

func GetMsg(msg string) string {
	content := ginI18n.MustGetMessage(msg)
	if content == "" {
		return msg
	} else {
		return content
	}
}
