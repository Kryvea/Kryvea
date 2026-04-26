package reportdata

import (
	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/google/uuid"
)

type AggregatedVulnerability struct {
	mongo.Vulnerability
	Targets []mongo.Target `json:"targets"`
}

// aggregateVulnerabilities aggregates vulnerabilities by Category ID,
// preserving the order of first occurrence (i.e. the sort order of the input slice).
//
// It must be called after parseHighlights to keep Highlighted fields.
func aggregateVulnerabilities(vulnerabilities []mongo.Vulnerability) []AggregatedVulnerability {
	aggregatedMap := make(map[uuid.UUID]*AggregatedVulnerability)
	order := []uuid.UUID{}

	for i := range vulnerabilities {
		v := vulnerabilities[i]
		categoryID := v.Category.ID

		if existing, ok := aggregatedMap[categoryID]; ok {
			existing.Targets = appendUniqueTarget(existing.Targets, v.Target)
			existing.Poc.Pocs = append(existing.Poc.Pocs, v.Poc.Pocs...)
			continue
		}

		aggregatedMap[categoryID] = vulnerabilityToAggregated(v)
		order = append(order, categoryID)
	}

	aggregatedVulnerabilities := make([]AggregatedVulnerability, len(order))
	for _, id := range order {
		aggregatedVulnerabilities = append(aggregatedVulnerabilities, *aggregatedMap[id])
	}

	return aggregatedVulnerabilities
}

func vulnerabilityToAggregated(vulnerability mongo.Vulnerability) *AggregatedVulnerability {
	aggregated := &AggregatedVulnerability{
		Vulnerability: vulnerability,
		Targets: []mongo.Target{
			vulnerability.Target,
		},
	}

	if len(vulnerability.Poc.Pocs) > 0 {
		cp := make([]mongo.PocItem, len(vulnerability.Poc.Pocs))
		copy(cp, vulnerability.Poc.Pocs)
		aggregated.Poc.Pocs = cp
	}
	return aggregated
}

func appendUniqueTarget(targets []mongo.Target, newTarget mongo.Target) []mongo.Target {
	for _, t := range targets {
		if t.ID == newTarget.ID {
			return targets
		}
	}
	return append(targets, newTarget)
}
