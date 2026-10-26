package i18n

import (
	"fmt"
	"strings"
	"sync"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	mcpDocumentBundleOnce sync.Once
	mcpDocumentBundle     *goi18n.Bundle
	mcpDocumentBundleErr  error
)

func NewMCPDocumentLocalizer(lang string) (*goi18n.Localizer, error) {
	mcpDocumentBundleOnce.Do(func() {
		documentBundle := goi18n.NewBundle(language.English)
		documentBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
		for _, file := range langFiles {
			if _, err := documentBundle.LoadMessageFileFS(fs, file); err != nil {
				mcpDocumentBundleErr = fmt.Errorf("load MCP document language %s: %w", file, err)
				return
			}
		}
		mcpDocumentBundle = documentBundle
	})
	if mcpDocumentBundleErr != nil {
		return nil, mcpDocumentBundleErr
	}
	if strings.TrimSpace(lang) == "" {
		lang = GetLanguage()
	}
	return goi18n.NewLocalizer(mcpDocumentBundle, lang, language.English.String()), nil
}
