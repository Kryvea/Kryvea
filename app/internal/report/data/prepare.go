package reportdata

import (
	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/model"
)

// vectorFor returns the vulnerability's CVSS vector for the given version.
// Unknown versions yield the zero Vector.
func vectorFor(vulnerability *model.Vulnerability, version string) cvss.Vector {
	switch version {
	case cvss.Cvss2:
		return vulnerability.CVSSv2
	case cvss.Cvss3:
		return vulnerability.CVSSv3
	case cvss.Cvss31:
		return vulnerability.CVSSv31
	case cvss.Cvss4:
		return vulnerability.CVSSv4
	}

	return cvss.Vector{}
}

func GetMaxCvss(vulnerabilities []model.Vulnerability, cvssVersions map[string]bool) map[string]cvss.Vector {
	maxCvss := make(map[string]cvss.Vector)

	for i := range vulnerabilities {
		for version, enabled := range cvssVersions {
			if !enabled {
				continue
			}

			if vector := vectorFor(&vulnerabilities[i], version); vector.Score > maxCvss[version].Score {
				maxCvss[version] = vector
			}
		}
	}

	return maxCvss
}

func getVulnerabilitiesOverview(vulnerabilities []model.Vulnerability, cvssVersions map[string]bool) map[string]map[string]uint {
	vulnerabilityOverview := make(map[string]map[string]uint)

	for _, version := range cvss.CvssVersions {
		vulnerabilityOverview[version] = make(map[string]uint)
		for _, severity := range cvss.CvssSeverities {
			vulnerabilityOverview[version][severity] = 0
		}
	}

	for i := range vulnerabilities {
		for _, version := range cvss.CvssVersions {
			if !cvssVersions[version] {
				continue
			}

			vulnerabilityOverview[version][vectorFor(&vulnerabilities[i], version).Severity] += 1
		}
	}

	return vulnerabilityOverview
}

// getTargetTagCounter counts, per Target.Tag, the vulnerabilities whose
// severity for maxVersion is not "None".
func getTargetTagCounter(vulnerabilities []model.Vulnerability, maxVersion string) map[string]uint {
	targetTagCounter := make(map[string]uint)

	for i := range vulnerabilities {
		if vectorFor(&vulnerabilities[i], maxVersion).Severity == cvss.CvssSeverityNone {
			continue
		}

		targetTagCounter[vulnerabilities[i].Target.Tag] += 1
	}

	return targetTagCounter
}

func getOWASPCounter(vulnerabilities []model.Vulnerability, maxVersion string) map[string]OWASPCounter {
	owaspCounter := make(map[string]OWASPCounter)

	highestSeverityByCategoryType := make(map[string]float64)

	for i := range vulnerabilities {
		vulnerability := &vulnerabilities[i]
		if _, ok := owaspCounter[vulnerability.Category.Source]; !ok {
			owaspCounter[vulnerability.Category.Source] = OWASPCounter{
				Categories: make(map[string]string),
			}
		}
		if _, ok := owaspCounter[vulnerability.Category.Source].Categories[vulnerability.Category.Identifier]; !ok {
			counter := owaspCounter[vulnerability.Category.Source]
			counter.Total += 1

			vector := vectorFor(vulnerability, maxVersion)
			if vector.Score > highestSeverityByCategoryType[vulnerability.Category.Identifier] {
				highestSeverityByCategoryType[vulnerability.Category.Identifier] = vector.Score
				counter.Categories[vulnerability.Category.Identifier] = severityColors[vector.Severity]
			}

			owaspCounter[vulnerability.Category.Source] = counter
		}
	}

	return owaspCounter
}

func parseHighlights(vulnerabilities []model.Vulnerability) {
	for i := range vulnerabilities {
		for j := range vulnerabilities[i].Poc.Pocs {
			parseHighlightedText(&vulnerabilities[i].Poc.Pocs[j])
		}
	}
}

func parseHighlightedText(pocitem *model.PocItem) {
	pocitem.RequestHighlighted = splitText(pocitem.Request, pocitem.RequestHighlights)
	pocitem.ResponseHighlighted = splitText(pocitem.Response, pocitem.ResponseHighlights)
	pocitem.TextHighlighted = splitText(pocitem.TextData, pocitem.TextHighlights)

	sanitizeReqResText(pocitem)
}
