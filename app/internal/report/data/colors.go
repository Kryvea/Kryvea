package reportdata

import (
	"github.com/Kryvea/Kryvea/internal/cvss"
)

var complexityColors = map[string]string{
	cvss.CvssSeverityLow:    "EE0000", // #EE0000
	cvss.CvssSeverityMedium: "FFC000", // #FFC000
	cvss.CvssSeverityHigh:   "92d050", // #92d050
}

func GetComplexityColor(complexity string) string {
	color, ok := complexityColors[complexity]
	if !ok {
		color = complexityColors[cvss.CvssSeverityHigh]
	}
	return color
}

var severityColors = map[string]string{
	cvss.CvssSeverityCritical: "7030A0", // #7030A0
	cvss.CvssSeverityHigh:     "EE0000", // #EE0000
	cvss.CvssSeverityMedium:   "FFC000", // #FFC000
	cvss.CvssSeverityLow:      "FFFF00", // #FFFF00
	cvss.CvssSeverityNone:     "92d050", // #92d050
}

func GetSeverityColor(severity string) string {
	color, ok := severityColors[severity]
	if !ok {
		color = severityColors[cvss.CvssSeverityNone]
	}
	return color
}
