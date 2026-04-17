package cvss

import (
	"strings"

	i18nUtils "github.com/Kryvea/Kryvea/internal/i18n"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	LanguageNotSupported = "Language not supported"
)

func (v *Vector) GenerateVectorDescription(lang string) string {
	localizer := i18nUtils.NewLocalizer(lang)

	fields := strings.Split(v.Vector, "/")[1:]
	vectorMap := make(map[string]string)

	for _, field := range fields {
		translated, err := localizer.Localize(&i18n.LocalizeConfig{
			MessageID: field,
		})
		if err == nil {
			metric := strings.Split(field, ":")[0]
			vectorMap[metric] = translated
		}
	}

	// Pick correct template based on version
	var msgID string
	switch v.Version {
	case Cvss31:
		msgID = "description.cvss31"
	case Cvss4:
		msgID = "description.cvss4"
	default:
		return LanguageNotSupported
	}

	description, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: vectorMap,
	})
	if err != nil {
		return LanguageNotSupported
	}
	return description
}
