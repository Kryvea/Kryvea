package util

import "github.com/Kryvea/Kryvea/internal/cvss"

func GetMaxCvssVersion(versions map[string]bool) string {
	maxVersionString := ""
	maxVersionValue := 0

	for cvssVersion, enabled := range versions {
		if !enabled {
			continue
		}

		if value := cvss.VersionToValue[cvssVersion]; value > maxVersionValue {
			maxVersionValue = value
			maxVersionString = cvssVersion
		}
	}

	return maxVersionString
}
