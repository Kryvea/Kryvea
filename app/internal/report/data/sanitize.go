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
	sanitizeTargets := make([]model.Target, len(assessment.Targets))
	for i, target := range assessment.Targets {
		sanitizeTarget(&target)
		sanitizeTargets[i] = target
	}

	assessment.Name = escapeXMLString(assessment.Name)
	assessment.Language = escapeXMLString(assessment.Language)
	assessment.Targets = sanitizeTargets
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

	// Sort by maxVersion score
	switch maxVersion {
	case cvss.Cvss2:
		sort.Slice(vulnerabilities, func(j, k int) bool {
			return vulnerabilities[j].CVSSv2.Score > vulnerabilities[k].CVSSv2.Score
		})
	case cvss.Cvss3:
		sort.Slice(vulnerabilities, func(j, k int) bool {
			return vulnerabilities[j].CVSSv3.Score > vulnerabilities[k].CVSSv3.Score
		})
	case cvss.Cvss31:
		sort.Slice(vulnerabilities, func(j, k int) bool {
			return vulnerabilities[j].CVSSv31.Score > vulnerabilities[k].CVSSv31.Score
		})
	case cvss.Cvss4:
		sort.Slice(vulnerabilities, func(j, k int) bool {
			return vulnerabilities[j].CVSSv4.Score > vulnerabilities[k].CVSSv4.Score
		})
	}
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

	sort.Slice(poc.Pocs, func(i, j int) bool {
		return poc.Pocs[i].Index < poc.Pocs[j].Index
	})
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
