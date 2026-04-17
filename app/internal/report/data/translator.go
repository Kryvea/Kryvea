package reportdata

import (
	"strings"

	internalI18n "github.com/Kryvea/Kryvea/internal/i18n"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	LanguageFallback = "en"
)

// Translate translates an English string into the target language.
// If translation is missing, it falls back to English or returns the original text.
func Translate(text, lang string) string {
	if lang == "" {
		lang = LanguageFallback
	}

	// Normalize language code
	lang = strings.ToLower(lang)

	// Create a localizer for the requested language
	localizer := internalI18n.NewLocalizer(lang)

	translated, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: text,
		DefaultMessage: &i18n.Message{
			ID:    text,
			Other: text,
		},
	})

	// Fallback logic
	if err != nil || translated == "" {
		if lang != LanguageFallback {
			fallbackLocalizer := internalI18n.NewLocalizer(LanguageFallback)
			fallbackTranslated, _ := fallbackLocalizer.Localize(&i18n.LocalizeConfig{
				MessageID: text,
				DefaultMessage: &i18n.Message{
					ID:    text,
					Other: text,
				},
			})
			if fallbackTranslated != "" {
				return fallbackTranslated
			}
		}
		return text
	}

	return translated
}
