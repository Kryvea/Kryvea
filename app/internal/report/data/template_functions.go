package reportdata

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kryvea/Kryvea/internal/cvss"
)

const (
	STYLE_WRAPPER_F = `<w:rPr>%s</w:rPr><w:t>%s</w:t>`
	SHADING_W_TAG_F = `<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`
)

var (
	SHADING_WRAPPER_F = fmt.Sprintf(STYLE_WRAPPER_F, SHADING_W_TAG_F, "%s")
)

func Debug(v any) string {
	return escapeXMLString(fmt.Sprintf("%#v", v))
}

// formatDate formats the given time using a locale-aware layout inferred from timezone.
// Signature: formatDate(t, tz?, style?)
//   - tz: optional IANA timezone (e.g., "Europe/Rome"). If missing/invalid, UTC is used.
//   - style: optional override (case-insensitive): "US" (MM/DD/YYYY), "EU" (DD/MM/YYYY), "ISO" (2006-01-02),
//     "YMD" (2006/01/02), "DMY" (02/01/2006), "MDY" (01/02/2006).
//
// Usage in templates:
//
//	{{ formatDate .DateTime "UTC" "EU" }}
func FormatDate(t time.Time, args ...string) string {
	// Location resolution (default UTC)
	loc := time.UTC
	if len(args) > 0 && args[0] != "" {
		if l, err := time.LoadLocation(args[0]); err == nil {
			loc = l
		}
	}

	// Layout selection
	layout := "02/01/2006" // default (DD/MM/YYYY) keeps backward-compat behavior

	// Optional explicit style override
	if len(args) > 1 && args[1] != "" {
		switch strings.ToUpper(strings.TrimSpace(args[1])) {
		case "US", "MDY":
			layout = "01/02/2006"
		case "EU", "DMY":
			layout = "02/01/2006"
		case "ISO", "YYYY-MM-DD", "Y-M-D":
			layout = "2006-01-02"
		case "YMD":
			layout = "2006/01/02"
		}
	} else if len(args) > 0 && args[0] != "" {
		// Infer from timezone -> country
		if cc, ok := countryForZone(args[0]); ok {
			layout = layoutForCountry(cc)
		}
	}

	return t.In(loc).Format(layout)
}

// GetOWASPColor returns the color associated with a specific OWASP category for a given counter.
//
// Parameters:
//   - counter: An OWASPCounter that holds a mapping of categories to colors.
//   - category: The OWASP category string (e.g., "A02:2021").
//
// Returns:
//   - The color string associated with the category. If the category is not present
//     in the counter, it defaults to the color corresponding to CvssSeverityNone.
//
// Usage in templates:
//
//	{{ getOWASPColor (index .OWASPCounter "owasp_web") "A02:2021" }}
func GetOWASPColor(counter OWASPCounter, category string) string {
	if color, ok := counter.Categories[category]; ok {
		return color
	}
	return severityColors[cvss.CvssSeverityNone]
}

// TableSeverityColor returns a formatted string suitable for use in a table cell,
// applying a background color based on the severity level.
//
// Parameters:
//   - severity: A string representing the severity level (e.g., "Low", "High").
//
// Returns:
//   - A string in the format "[[TABLE_CELL_BG_COLOR:<COLOR>]]<SEVERITY>",
//     where <COLOR> is the uppercase color corresponding to the severity.
//
// Usage in templates:
//
//	{{ tableSeverityColor .CVSSv4.Severity }}
func TableSeverityColor(severity string) string {
	color := GetSeverityColor(severity)
	return fmt.Sprintf("[[TABLE_CELL_BG_COLOR:%s]]", strings.ToUpper(color))
}

// TableComplexityColor returns a formatted string suitable for use in a table cell,
// applying a background color based on the complexity level. Only the background color
// is included; the cell text is omitted.
//
// Parameters:
//   - complexity: A string representing the complexity level (e.g., "Low", "High").
//
// Returns:
//   - A string in the format "[[TABLE_CELL_BG_COLOR:<COLOR>]]",
//     where <COLOR> is the uppercase color corresponding to the complexity.
//
// Usage in templates:
//
//	{{ tableComplexityColor .CVSSv4.Complexity }}
func TableComplexityColor(complexity string) string {
	color := GetComplexityColor(complexity)
	return fmt.Sprintf("[[TABLE_CELL_BG_COLOR:%s]]", strings.ToUpper(color))
}

// ShadeTextBg applies a background color shading to the given text and returns
// a formatted string suitable for embedding in WordprocessingML (DOCX) content.
//
// Parameters:
//   - s: The text string to be wrapped with a background color.
//   - hex: The hex code of the color (e.g., "#FF0000" or "FF0000"). The "#" prefix is optional.
//
// Behavior:
//   - If the hex string is invalid (not exactly 6 characters after removing a leading '#'),
//     the original text `s` is returned unmodified.
//   - If valid, the function wraps the text in a WordprocessingML <w:rPr> and <w:t> tag
//     with a <w:shd> element specifying the background color.
//
// Returns:
//   - A string containing the original text `s` wrapped in WordprocessingML tags
//     with the specified background color applied.
//
// Usage in templates:
//
//	{{shadeTextBg "Important", "#FFCC00"}}
//	// Output: <w:rPr><w:shd w:val="clear" w:color="auto" w:fill="FFCC00"/></w:rPr><w:t>Important</w:t>
func ShadeTextBg(s, hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		hex = ""
	}

	return fmt.Sprintf(SHADING_WRAPPER_F, strings.ToUpper(hex), s)
}
