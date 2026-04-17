package cvss

import (
	"errors"
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
	if _, ok := VersionToValue[version]; !ok {
		return nil, errors.New("no severity levels found for given CVSS version")
	}

	if version == Cvss2 && strings.HasPrefix(vectorString, Cvss2Prefix) {
		vectorParts := strings.Split(vectorString, Cvss2Prefix)
		if len(vectorParts) < 2 {
			return nil, errors.New("failed to parse CVSS2 vector with prefix")
		}
		vectorString = vectorParts[1]
	}

	score, complexity, err := calculateScoreAndComplexity(vectorString, version)
	if err != nil {
		return nil, err
	}

	vector := &Vector{
		Version:     version,
		Vector:      vectorString,
		Score:       score,
		Complexity:  complexity,
		Severity:    "",
		Description: "",
	}

	severityThresholds := severityLevels[version]
	for _, threshold := range severityThresholds {
		if vector.Score >= threshold.Score {
			vector.Severity = threshold.Severity
			break
		}
	}

	if vector.Score > 0 {
		vector.Description = vector.GenerateVectorDescription(language)
	}

	if vector.Severity == "" {
		vector.Severity = CvssSeverityNone
	}

	return vector, nil
}

// calculateScore calculates the CVSS score and complexity based on the vector string and version.
func calculateScoreAndComplexity(vector string, version string) (float64, string, error) {
	switch version {
	case Cvss2:
		cvss, err := gocvss20.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}

		ac, err := cvss.Get("AC")
		if err != nil {
			return 0, "", err
		}

		return cvss.EnvironmentalScore(), getComplexity(ac), nil
	case Cvss3:
		cvss, err := gocvss30.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}

		ac, err := cvss.Get("AC")
		if err != nil {
			return 0, "", err
		}

		return cvss.EnvironmentalScore(), getComplexity(ac), nil
	case Cvss31:
		cvss, err := gocvss31.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}

		ac, err := cvss.Get("AC")
		if err != nil {
			return 0, "", err
		}

		return cvss.EnvironmentalScore(), getComplexity(ac), nil
	case Cvss4:
		cvss, err := gocvss40.ParseVector(vector)
		if err != nil {
			return 0, "", err
		}

		ac, err := cvss.Get("AC")
		if err != nil {
			return 0, "", err
		}

		return cvss.Score(), getComplexity(ac), nil
	default:
		return 0, "", errors.New("invalid CVSS version")
	}
}

var complexities map[string]string = map[string]string{
	"L": CvssSeverityLow,
	"M": CvssSeverityMedium,
	"H": CvssSeverityHigh,
}

func getComplexity(c string) string {
	if complexity, ok := complexities[c]; ok {
		return complexity
	}

	return CvssSeverityLow
}
