package i18n

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bytedance/sonic"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

// copy LocalizeConfig
type LocalizeConfig struct {
	MessageID    string
	TemplateData map[string]string
}

func InitI18n(localesPath string) error {
	bundle = i18n.NewBundle(language.English) // default fallback
	bundle.RegisterUnmarshalFunc("json", sonic.Unmarshal)

	err := filepath.Walk(localesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			if _, err := bundle.LoadMessageFile(path); err != nil {
				return fmt.Errorf("failed to load locale file %s: %w", path, err)
			}
		}
		return nil
	})
	return err
}

func NewLocalizer(lang string) *i18n.Localizer {
	return i18n.NewLocalizer(bundle, lang)
}
