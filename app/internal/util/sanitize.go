package util

import (
	"fmt"
	"net/url"
	"strings"
)

// fileNameReplacer replaces invalid characters with underscores and strips double quotes
var fileNameReplacer = strings.NewReplacer(
	"/", "_",
	"\\", "_",
	":", "_",
	"*", "_",
	"?", "_",
	"<", "_",
	">", "_",
	"|", "_",
	"\"", "",
)

func SanitizeFileName(name string) string {
	return fileNameReplacer.Replace(name)
}

func ContentDispositionAttachment(name string) string {
	sanitized := SanitizeFileName(name)
	encoded := strings.ReplaceAll(url.QueryEscape(sanitized), "+", "%20")
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
		asciiFilename(sanitized), encoded)
}

func asciiFilename(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
