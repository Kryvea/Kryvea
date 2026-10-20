package reportdata

import (
	"bytes"
	"encoding/xml"
	"sort"
	"strings"

	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/model"
)

func SanitizeCustomer(customer *model.Customer) {
	customer.Name = escapeXMLString(customer.Name)
	customer.Language = escapeXMLString(customer.Language)
}

func SanitizeAssessment(assessment *model.Assessment) {
	for i := range assessment.Targets {
		sanitizeTarget(&assessment.Targets[i])
	}

	assessment.Name = escapeXMLString(assessment.Name)
	assessment.Language = escapeXMLString(assessment.Language)
	assessment.Status = escapeXMLString(assessment.Status)
	assessment.Type.Short = escapeXMLString(assessment.Type.Short)
	assessment.Type.Full = escapeXMLString(assessment.Type.Full)
	assessment.Environment = escapeXMLString(assessment.Environment)
	assessment.TestingType = escapeXMLString(assessment.TestingType)
	assessment.OSSTMMVector = escapeXMLString(assessment.OSSTMMVector)
}

func sanitizeTarget(target *model.Target) {
	target.IPv4 = escapeXMLString(target.IPv4)
	target.IPv6 = escapeXMLString(target.IPv6)
	target.Protocol = escapeXMLString(target.Protocol)
	target.FQDN = escapeXMLString(target.FQDN)
	target.Tag = escapeXMLString(target.Tag)
}

func SanitizeAndSortVulnerabilities(vulnerabilities []model.Vulnerability, maxVersion string, language string) {
	if len(vulnerabilities) == 0 {
		return
	}

	for i := range vulnerabilities {
		sanitizeVulnerability(&vulnerabilities[i])
	}

	sortVulnerabilitiesByScore(vulnerabilities, maxVersion)
}

// SortVulnerabilities is the escaping-free counterpart of
// SanitizeAndSortVulnerabilities: same ordering, no text mutation.
func SortVulnerabilities(vulnerabilities []model.Vulnerability, maxVersion string) {
	for i := range vulnerabilities {
		sortPocsByIndex(vulnerabilities[i].Poc.Pocs)
	}

	sortVulnerabilitiesByScore(vulnerabilities, maxVersion)
}

func sortVulnerabilitiesByScore(vulnerabilities []model.Vulnerability, maxVersion string) {
	if !cvss.IsValidVersion(maxVersion) {
		return
	}

	sort.Slice(vulnerabilities, func(j, k int) bool {
		return vectorFor(&vulnerabilities[j], maxVersion).Score > vectorFor(&vulnerabilities[k], maxVersion).Score
	})
}

func sortPocsByIndex(pocs []model.PocItem) {
	sort.Slice(pocs, func(i, j int) bool {
		return pocs[i].Index < pocs[j].Index
	})
}

func sanitizeVulnerability(item *model.Vulnerability) {
	SanitizeAndSortPoc(&item.Poc)

	item.Category.Identifier = escapeXMLString(item.Category.Identifier)
	item.Category.Name = escapeXMLString(item.Category.Name)
	item.Category.Subcategory = escapeXMLString(item.Category.Subcategory)
	item.DetailedTitle = escapeXMLString(item.DetailedTitle)
	item.Status = escapeXMLString(item.Status)

	sanitizeVector(&item.CVSSv2)
	sanitizeVector(&item.CVSSv3)
	sanitizeVector(&item.CVSSv31)
	sanitizeVector(&item.CVSSv4)

	for i, reference := range item.References {
		item.References[i] = escapeXMLString(reference)
	}

	item.GenericDescription.Text = escapeXMLString(item.GenericDescription.Text)
	item.GenericRemediation.Text = escapeXMLString(item.GenericRemediation.Text)
	item.Description = escapeXMLString(item.Description)
	item.Remediation = escapeXMLString(item.Remediation)
	sanitizeTarget(&item.Target)
}

func sanitizeVector(item *cvss.Vector) {
	item.Version = escapeXMLString(item.Version)
	item.Vector = escapeXMLString(item.Vector)
	item.Severity = escapeXMLString(item.Severity)
	item.Complexity = escapeXMLString(item.Complexity)
	item.Description = escapeXMLString(item.Description)
}

func SanitizeAndSortPoc(poc *model.Poc) {
	if len(poc.Pocs) == 0 {
		return
	}

	for i := range poc.Pocs {
		sanitizePocItem(&poc.Pocs[i])
	}

	sortPocsByIndex(poc.Pocs)
}

func sanitizePocItem(item *model.PocItem) {
	item.Type = escapeXMLString(item.Type)
	item.Description = escapeXMLString(item.Description)
	item.URI = escapeXMLString(item.URI)
	item.ImageFilename = escapeXMLString(item.ImageFilename)
	item.ImageCaption = escapeXMLString(item.ImageCaption)
	item.TextLanguage = escapeXMLString(item.TextLanguage)
}

func sanitizeReqResText(item *model.PocItem) {
	item.Request = escapeXMLString(item.Request)
	item.Response = escapeXMLString(item.Response)
	item.TextData = escapeXMLString(item.TextData)
}

func escapeXMLString(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	escaped := strings.ReplaceAll(buf.String(), "&#xA;", "\n")
	return escaped
}
