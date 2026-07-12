package i18n

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/bytedance/sonic"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

func InitI18n(localesPath string) error {
	bundle = i18n.NewBundle(language.English) // default fallback
	bundle.RegisterUnmarshalFunc("json", sonic.Unmarshal)

	return filepath.WalkDir(localesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".json" {
			if _, err := bundle.LoadMessageFile(path); err != nil {
				return fmt.Errorf("failed to load locale file %s: %w", path, err)
			}
		}
		return nil
	})
}

func NewLocalizer(lang string) *i18n.Localizer {
	b := bundle
	if b == nil {
		// InitI18n has not been called (e.g. in tests): fall back to an
		// empty bundle so lookups fail gracefully instead of panicking.
		b = i18n.NewBundle(language.English)
	}
	return i18n.NewLocalizer(b, lang)
}
