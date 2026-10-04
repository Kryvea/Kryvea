package reportdata

import (
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type AggregatedVulnerability struct {
	model.Vulnerability
	Targets []model.Target `json:"targets"`
}

// aggregateVulnerabilities aggregates vulnerabilities by Category ID,
// preserving the order of first occurrence (i.e. the sort order of the input slice).
//
// It must be called after parseHighlights to keep Highlighted fields.
func aggregateVulnerabilities(vulnerabilities []model.Vulnerability) []AggregatedVulnerability {
	aggregatedMap := make(map[uuid.UUID]*AggregatedVulnerability)
	order := []uuid.UUID{}

	for i := range vulnerabilities {
		v := vulnerabilities[i]
		categoryID := v.Category.ID

		if existing, ok := aggregatedMap[categoryID]; ok {
			existing.Targets = appendUniqueTarget(existing.Targets, v.Target)
			existing.Poc.Pocs = appendUniquePocItems(existing.Poc.Pocs, v.Poc.Pocs)
			continue
		}

		aggregatedMap[categoryID] = vulnerabilityToAggregated(v)
		order = append(order, categoryID)
	}

	aggregatedVulnerabilities := make([]AggregatedVulnerability, 0, len(order))
	for _, id := range order {
		aggregatedVulnerabilities = append(aggregatedVulnerabilities, *aggregatedMap[id])
	}

	return aggregatedVulnerabilities
}

func vulnerabilityToAggregated(vulnerability model.Vulnerability) *AggregatedVulnerability {
	aggregated := &AggregatedVulnerability{
		Vulnerability: vulnerability,
		Targets: []model.Target{
			vulnerability.Target,
		},
	}

	if len(vulnerability.Poc.Pocs) > 0 {
		cp := make([]model.PocItem, len(vulnerability.Poc.Pocs))
		copy(cp, vulnerability.Poc.Pocs)
		aggregated.Poc.Pocs = cp
	}
	return aggregated
}

func appendUniqueTarget(targets []model.Target, newTarget model.Target) []model.Target {
	for _, t := range targets {
		if t.ID == newTarget.ID {
			return targets
		}
	}
	return append(targets, newTarget)
}

// dedupes by content; ID and Index are intentionally ignored
func appendUniquePocItems(existing []model.PocItem, incoming []model.PocItem) []model.PocItem {
	for _, item := range incoming {
		if containsPocItem(existing, item) {
			continue
		}
		existing = append(existing, item)
	}
	return existing
}

func containsPocItem(items []model.PocItem, target model.PocItem) bool {
	for _, item := range items {
		if pocItemContentEqual(item, target) {
			return true
		}
	}
	return false
}

func pocItemContentEqual(a, b model.PocItem) bool {
	return a.Type == b.Type &&
		a.Description == b.Description &&
		a.URI == b.URI &&
		a.Request == b.Request &&
		a.Response == b.Response &&
		a.ImageID == b.ImageID &&
		a.ImageReference == b.ImageReference &&
		a.ImageFilename == b.ImageFilename &&
		a.ImageCaption == b.ImageCaption &&
		a.TextLanguage == b.TextLanguage &&
		a.TextData == b.TextData &&
		a.StartingLineNumber == b.StartingLineNumber
}
