package util

import "strings"

func SanitizeFileName(name string) string {
	// Replace invalid characters with underscores
	replacer := strings.NewReplacer(
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
	return replacer.Replace(name)
}
