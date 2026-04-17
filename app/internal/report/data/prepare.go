package reportdata

import (
	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/mongo"
)

func GetMaxCvss(vulnerabilities []mongo.Vulnerability, cvssVersions map[string]bool) map[string]cvss.Vector {
	maxCvss := make(map[string]cvss.Vector)

	for _, vulnerability := range vulnerabilities {
		for version, enabled := range cvssVersions {
			if !enabled {
				continue
			}

			switch version {
			case cvss.Cvss2:
				if vulnerability.CVSSv2.Score > maxCvss[version].Score {
					maxCvss[version] = vulnerability.CVSSv2
				}
			case cvss.Cvss3:
				if vulnerability.CVSSv3.Score > maxCvss[version].Score {
					maxCvss[version] = vulnerability.CVSSv3
				}
			case cvss.Cvss31:
				if vulnerability.CVSSv31.Score > maxCvss[version].Score {
					maxCvss[version] = vulnerability.CVSSv31
				}
			case cvss.Cvss4:
				if vulnerability.CVSSv4.Score > maxCvss[version].Score {
					maxCvss[version] = vulnerability.CVSSv4
				}
			}
		}
	}

	return maxCvss
}

func getVulnerabilitiesOverview(vulnerabilities []mongo.Vulnerability, cvssVersions map[string]bool) map[string]map[string]uint {
	vulnerabilityOverview := make(map[string]map[string]uint)

	for _, version := range cvss.CvssVersions {
		vulnerabilityOverview[version] = make(map[string]uint)
		for _, severity := range cvss.CvssSeverities {
			vulnerabilityOverview[version][severity] = 0
		}
	}

	for _, vulnerability := range vulnerabilities {
		for _, version := range cvss.CvssVersions {
			if vulnerabilityOverview[version] == nil {
				vulnerabilityOverview[version] = make(map[string]uint)
				for _, severity := range cvss.CvssSeverities {
					vulnerabilityOverview[version][severity] = 0
				}
			}

			if !cvssVersions[version] {
				continue
			}

			switch version {
			case cvss.Cvss2:
				vulnerabilityOverview[version][vulnerability.CVSSv2.Severity] += 1
			case cvss.Cvss3:
				vulnerabilityOverview[version][vulnerability.CVSSv3.Severity] += 1
			case cvss.Cvss31:
				vulnerabilityOverview[version][vulnerability.CVSSv31.Severity] += 1
			case cvss.Cvss4:
				vulnerabilityOverview[version][vulnerability.CVSSv4.Severity] += 1
			}
		}
	}

	return vulnerabilityOverview
}

func getTargetsCategoryCounter(vulnerabilities []mongo.Vulnerability, maxVersion string) map[string]uint {
	targetsCategoryCounter := make(map[string]uint)

	for _, vulnerability := range vulnerabilities {
		if (maxVersion == cvss.Cvss2 && vulnerability.CVSSv2.Severity == cvss.CvssSeverityNone) ||
			(maxVersion == cvss.Cvss3 && vulnerability.CVSSv3.Severity == cvss.CvssSeverityNone) ||
			(maxVersion == cvss.Cvss31 && vulnerability.CVSSv31.Severity == cvss.CvssSeverityNone) ||
			(maxVersion == cvss.Cvss4 && vulnerability.CVSSv4.Severity == cvss.CvssSeverityNone) {
			continue
		}

		targetsCategoryCounter[vulnerability.Target.Tag] += 1
	}

	return targetsCategoryCounter
}

func getOWASPCounter(vulnerabilities []mongo.Vulnerability, maxVersion string) map[string]OWASPCounter {
	owaspCounter := make(map[string]OWASPCounter)

	highestSeverityByCategoryType := make(map[string]float64)

	for _, vulnerability := range vulnerabilities {
		if _, ok := owaspCounter[vulnerability.Category.Source]; !ok {
			owaspCounter[vulnerability.Category.Source] = OWASPCounter{
				Categories: make(map[string]string),
			}
		}
		if _, ok := owaspCounter[vulnerability.Category.Source].Categories[vulnerability.Category.Identifier]; !ok {
			counter := owaspCounter[vulnerability.Category.Source]
			counter.Total += 1

			switch maxVersion {
			case cvss.Cvss2:
				if vulnerability.CVSSv2.Score > highestSeverityByCategoryType[vulnerability.Category.Identifier] {
					highestSeverityByCategoryType[vulnerability.Category.Identifier] = vulnerability.CVSSv2.Score
					counter.Categories[vulnerability.Category.Identifier] = severityColors[vulnerability.CVSSv2.Severity]
				}
			case cvss.Cvss3:
				if vulnerability.CVSSv3.Score > highestSeverityByCategoryType[vulnerability.Category.Identifier] {
					highestSeverityByCategoryType[vulnerability.Category.Identifier] = vulnerability.CVSSv3.Score
					counter.Categories[vulnerability.Category.Identifier] = severityColors[vulnerability.CVSSv3.Severity]
				}
			case cvss.Cvss31:
				if vulnerability.CVSSv31.Score > highestSeverityByCategoryType[vulnerability.Category.Identifier] {
					highestSeverityByCategoryType[vulnerability.Category.Identifier] = vulnerability.CVSSv31.Score
					counter.Categories[vulnerability.Category.Identifier] = severityColors[vulnerability.CVSSv31.Severity]
				}
			case cvss.Cvss4:
				if vulnerability.CVSSv4.Score > highestSeverityByCategoryType[vulnerability.Category.Identifier] {
					highestSeverityByCategoryType[vulnerability.Category.Identifier] = vulnerability.CVSSv4.Score
					counter.Categories[vulnerability.Category.Identifier] = severityColors[vulnerability.CVSSv4.Severity]
				}
			}

			owaspCounter[vulnerability.Category.Source] = counter
		}
	}

	return owaspCounter
}

func parseHighlights(vulnerabilities []mongo.Vulnerability) {
	for i := range vulnerabilities {
		for j := range vulnerabilities[i].Poc.Pocs {
			parseHighlightedText(&vulnerabilities[i].Poc.Pocs[j])
		}
	}
}

func parseHighlightedText(pocitem *mongo.PocItem) {
	pocitem.RequestHighlighted = splitText(pocitem.Request, pocitem.RequestHighlights)
	pocitem.ResponseHighlighted = splitText(pocitem.Response, pocitem.ResponseHighlights)
	pocitem.TextHighlighted = splitText(pocitem.TextData, pocitem.TextHighlights)

	sanitizeReqResText(pocitem)
}
