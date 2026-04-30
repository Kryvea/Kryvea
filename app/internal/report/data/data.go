package reportdata

import (
	"time"

	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/util"
)

type ReportData struct {
	Customer                  *model.Customer
	Assessment                *model.Assessment
	Vulnerabilities           []model.Vulnerability
	AggregatedVulnerabilities []AggregatedVulnerability
	DeliveryDateTime          time.Time
	MaxCVSS                   map[string]cvss.Vector     // maps each cvss version to the vector with the highest score
	VulnerabilitiesOverview   map[string]map[string]uint // maps each cvss version to the aggregated vulnerability counts grouped by severity levels
	TargetsCategoryCounter    map[string]uint            // groups all targets by name and maps them to the number of their occurrences
	OWASPCounter              map[string]OWASPCounter    // maps each category type to the OWASPCounter struct
}

// OWASPCounter represents a summary of findings for a given assessment.
//   - Categories maps each OWASP category to the hex color corresponding
//     to the highest found severity for that category
//   - Total: the total number of unique categories for which at least
//     one vulnerability has been found
type OWASPCounter struct {
	Categories map[string]string
	Total      uint
}

// Prepare sanitizes and sorts all report data. sortByCvss specifies which CVSS
// version to use for ordering (e.g. "3.1", "4.0"); if empty, the highest
// enabled version in the assessment is used.
func (rd *ReportData) Prepare(sortByCvss string) {
	maxVersion := sortByCvss
	if maxVersion == "" {
		maxVersion = util.GetMaxCvssVersion(rd.Assessment.CVSSVersions)
	}

	// sanitize customer
	SanitizeCustomer(rd.Customer)

	// sanitize assessment
	SanitizeAssessment(rd.Assessment)

	// sanitize and sort vulnerabilities
	SanitizeAndSortVulnerabilities(rd.Vulnerabilities, maxVersion, rd.Assessment.Language)

	// get max cvss
	rd.MaxCVSS = GetMaxCvss(rd.Vulnerabilities, rd.Assessment.CVSSVersions)

	rd.VulnerabilitiesOverview = getVulnerabilitiesOverview(rd.Vulnerabilities, rd.Assessment.CVSSVersions)

	rd.TargetsCategoryCounter = getTargetsCategoryCounter(rd.Vulnerabilities, maxVersion)

	rd.OWASPCounter = getOWASPCounter(rd.Vulnerabilities, maxVersion)

	// parse pocitem Highlights
	parseHighlights(rd.Vulnerabilities)

	// aggregate vulnerabilities
	rd.AggregatedVulnerabilities = aggregateVulnerabilities(rd.Vulnerabilities)
}
