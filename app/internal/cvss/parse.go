package cvss

import (
	"errors"
	"fmt"
	"strings"

	gocvss20 "github.com/pandatix/go-cvss/20"
	gocvss30 "github.com/pandatix/go-cvss/30"
	gocvss31 "github.com/pandatix/go-cvss/31"
	gocvss40 "github.com/pandatix/go-cvss/40"
)

type Vector struct {
	Version     string  `json:"version" bson:"version"`
	Vector      string  `json:"vector" bson:"vector"`
	Score       float64 `json:"score" bson:"score"`
	Severity    string  `json:"severity" bson:"severity"`
	Complexity  string  `json:"complexity" bson:"complexity"`
	Description string  `json:"description" bson:"description"`
}

// ParseVector parses a CVSS vector string and returns a pointer to
// a Vector calculated based on the specified CVSS version.
// Description field should be generated separately.
func ParseVector(vectorString, version, language string) (*Vector, error) {
	if version == Cvss2 {
		vectorString = strings.TrimPrefix(vectorString, Cvss2Prefix)
	}

	score, complexity, err := calculateScoreAndComplexity(vectorString, version)
	if err != nil {
		return nil, err
	}

	vector := &Vector{
		Version:    version,
		Vector:     vectorString,
		Score:      score,
		Complexity: complexity,
	}

	for _, threshold := range severityLevels[version] {
		if vector.Score >= threshold.Score {
			vector.Severity = threshold.Severity
			break
		}
	}

	if vector.Score > 0 {
		vector.Description = vector.GenerateVectorDescription(language)
	}

	return vector, nil
}

// calculateScoreAndComplexity calculates the CVSS score and complexity based on the vector string and version.
func calculateScoreAndComplexity(vector string, version string) (float64, string, error) {
	switch version {
	case Cvss2:
		cvss, err := gocvss20.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}
		return scoreAndComplexity(cvss.EnvironmentalScore, cvss.Get)
	case Cvss3:
		cvss, err := gocvss30.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}
		return scoreAndComplexity(cvss.EnvironmentalScore, cvss.Get)
	case Cvss31:
		cvss, err := gocvss31.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}
		return scoreAndComplexity(cvss.EnvironmentalScore, cvss.Get)
	case Cvss4:
		cvss, err := gocvss40.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}
		return scoreAndComplexity(cvss.Score, cvss.Get)
	default:
		return 0, "", errors.New("invalid CVSS version")
	}
}

// scoreAndComplexity extracts the score and the complexity ("AC" metric)
// from an already parsed CVSS vector.
func scoreAndComplexity(score func() float64, get func(string) (string, error)) (float64, string, error) {
	ac, err := get("AC")
	if err != nil {
		return 0, "", err
	}

	complexity, err := getComplexity(ac)
	if err != nil {
		return 0, "", err
	}

	return score(), complexity, nil
}

// Complexity display values for the CVSS "AC" (attack complexity) metric.
const (
	complexityLow    = "Low"
	complexityMedium = "Medium"
	complexityHigh   = "High"
)

var complexities = map[string]string{
	"L": complexityLow,
	"M": complexityMedium,
	"H": complexityHigh,
}

// getComplexity maps a raw CVSS "AC" metric value to its display name.
// Every CVSS version defines AC as one of L, M or H, so any other value
// is rejected instead of being silently defaulted.
func getComplexity(c string) (string, error) {
	if complexity, ok := complexities[c]; ok {
		return complexity, nil
	}

	return "", fmt.Errorf("unknown CVSS attack complexity value %q", c)
}
