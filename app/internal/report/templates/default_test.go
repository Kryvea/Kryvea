package templates

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/i18n"
	"github.com/Kryvea/Kryvea/internal/model"
	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
	"github.com/google/uuid"
)

const (
	imageData = "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"
)

func randName(n int) string {
	var names = []string{"Ace", "Blaze", "Nova", "Zane", "Kai", "Orion", "Jett", "Echo", "Maverick", "Axel", "Ryder", "Phoenix", "Storm", "Dash", "Sable", "Ember", "Zephyr", "Titan", "Knox", "Luna", "Indigo", "Raven", "Aspen", "Atlas", "Juno", "Onyx", "Sage", "Vega", "Zara", "Xander", "Aria", "Dante", "Hunter", "Skye", "Rogue", "Kairos", "Hawk", "Shadow", "Nyx", "Lyric"}

	var name string
	for i := 0; i < n; i++ {
		name += names[rand.Intn(len(names))]
		if i < n-1 {
			name += " "
		}
	}
	return name
}

func randLanguage() string {
	var languages = []string{"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko"}
	return languages[rand.Intn(len(languages))]
}

func randIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func randHostname() string {
	return fmt.Sprintf("%s.example.com", randName(1))
}

func randStatus() string {
	var statuses = []string{"On Hold", "In Progress", "Completed"}
	return statuses[rand.Intn(len(statuses))]
}

func randAssessmentType() model.AssessmentType {
	typesKey := []string{"WAPT", "VAPT", "MAPT", "IoT", "Red Team Assessment"}
	typesValue := []string{
		"Web Application Penetration Test",
		"Vulnerability Assessment and Penetration Testing",
		"Mobile Application Penetration Test",
		"Internet of Things Penetration Test",
		"Red Team Assessment",
	}

	selected := rand.Intn(len(typesKey))

	return model.AssessmentType{
		Short: typesKey[selected],
		Full:  typesValue[selected],
	}
}

func randEnvironment() string {
	var environments = []string{"Pre-Production", "Production"}
	return environments[rand.Intn(len(environments))]
}

func randTestingType() string {
	var types = []string{"Black Box", "White Box", "Gray Box"}
	return types[rand.Intn(len(types))]
}

func randOSSTMMVector() string {
	var vectors = []string{"Inside to Inside", "Inside to Outside", "Outside to Inside", "Outside to Outside"}
	return vectors[rand.Intn(len(vectors))]
}

func randCVSSVector(version string) string {
	switch version {
	case cvss.Cvss2:
		vectors := []string{
			"AV:N/AC:L/Au:N/C:N/I:N/A:C",
			"AV:N/AC:L/Au:N/C:P/I:N/A:C",
			"AV:N/AC:L/Au:N/C:C/I:N/A:C",
			"AV:N/AC:L/Au:N/C:C/I:C/A:C",
			"AV:N/AC:L/Au:N/C:C/I:C/A:N",
			"AV:N/AC:L/Au:N/C:C/I:N/A:N",
			"AV:N/AC:L/Au:N/C:P/I:P/A:C",
			"AV:N/AC:L/Au:N/C:P/I:P/A:N",
			"AV:N/AC:L/Au:N/C:P/I:N/A:C",
			"AV:N/AC:L/Au:N/C:P/I:N/A:N",
			"AV:N/AC:L/Au:N/C:N/I:P/A:C",
			"AV:N/AC:L/Au:N/C:N/I:P/A:N",
		}
		return vectors[rand.Intn(len(vectors))]
	case cvss.Cvss3:
		vectors := []string{
			"CVSS:3.0/AV:A/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:L",
			"CVSS:3.0/AV:N/AC:H/PR:N/UI:R/S:U/C:N/I:L/A:H",
			"CVSS:3.0/AV:N/AC:H/PR:L/UI:R/S:C/C:L/I:L/A:H",
			"CVSS:3.0/AV:N/AC:H/PR:L/UI:R/S:C/C:L/I:H/A:N",
			"CVSS:3.0/AV:N/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:N",
			"CVSS:3.0/AV:L/AC:H/PR:L/UI:N/S:U/C:N/I:L/A:N",
			"CVSS:3.0/AV:A/AC:H/PR:H/UI:N/S:U/C:N/I:H/A:N",
			"CVSS:3.0/AV:P/AC:H/PR:H/UI:N/S:C/C:N/I:L/A:N",
			"CVSS:3.0/AV:N/AC:L/PR:L/UI:R/S:C/C:L/I:L/A:N",
			"CVSS:3.0/AV:A/AC:H/PR:L/UI:N/S:C/C:H/I:H/A:H",
		}
		return vectors[rand.Intn(len(vectors))]
	case cvss.Cvss31:
		vectors := []string{
			"CVSS:3.1/AV:A/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:L",
			"CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:U/C:N/I:L/A:H",
			"CVSS:3.1/AV:N/AC:H/PR:L/UI:R/S:C/C:L/I:L/A:H",
			"CVSS:3.1/AV:N/AC:H/PR:L/UI:R/S:C/C:L/I:H/A:N",
			"CVSS:3.1/AV:N/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:N",
			"CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:U/C:N/I:L/A:N",
			"CVSS:3.1/AV:A/AC:H/PR:H/UI:N/S:U/C:N/I:H/A:N",
			"CVSS:3.1/AV:P/AC:H/PR:H/UI:N/S:C/C:N/I:L/A:N",
			"CVSS:3.1/AV:N/AC:L/PR:L/UI:R/S:C/C:L/I:L/A:N",
			"CVSS:3.1/AV:A/AC:H/PR:L/UI:N/S:C/C:H/I:H/A:H",
		}
		return vectors[rand.Intn(len(vectors))]
	case cvss.Cvss4:
		vectors := []string{
			"CVSS:4.0/AV:A/AC:H/AT:P/PR:L/UI:P/VC:L/VI:L/VA:L/SC:L/SI:L/SA:L",
			"CVSS:4.0/AV:A/AC:H/AT:N/PR:L/UI:P/VC:H/VI:H/VA:L/SC:L/SI:L/SA:L",
			"CVSS:4.0/AV:A/AC:H/AT:N/PR:L/UI:A/VC:H/VI:N/VA:L/SC:L/SI:H/SA:L",
			"CVSS:4.0/AV:N/AC:L/AT:N/PR:L/UI:A/VC:H/VI:N/VA:L/SC:L/SI:H/SA:L",
			"CVSS:4.0/AV:N/AC:L/AT:P/PR:L/UI:P/VC:H/VI:N/VA:L/SC:L/SI:H/SA:L",
			"CVSS:4.0/AV:N/AC:L/AT:P/PR:L/UI:P/VC:H/VI:L/VA:L/SC:N/SI:H/SA:H",
		}
		return vectors[rand.Intn(len(vectors))]
	}

	return "CVSS:3.1/AV:A/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:L"
}

func randUrl() string {
	var urls = []string{"https://example.com", "https://example.org", "https://example.net"}
	return urls[rand.Intn(len(urls))]
}

func TestDefault(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() = %v, want %v", err, nil)
	}

	err = i18n.InitI18n(filepath.Join(path, "..", "..", "..", "locales"))
	if err != nil {
		t.Fatalf("i18n.InitI18n() = %v, want %v", err, nil)
	}

	customer := &model.Customer{
		Name:     randName(3),
		Language: randLanguage(),
	}

	var targets []model.Target
	for i := 0; i < 2; i++ {
		targets = append(targets, model.Target{
			IPv4: randIP(), FQDN: randHostname(),
		})
	}

	assessment := &model.Assessment{
		Name:            randName(3),
		Language:        customer.Language,
		StartDateTime:   time.Now().Add(-time.Hour * 24 * 7),
		EndDateTime:     time.Now(),
		KickoffDateTime: time.Now(),
		Targets:         targets,
		Status:          randStatus(),
		Type:            randAssessmentType(),
		Environment:     randEnvironment(),
		TestingType:     randTestingType(),
		OSSTMMVector:    randOSSTMMVector(),
	}

	cvssVersions := make(map[string]bool)
	for _, version := range cvss.CvssVersions {
		cvssVersions[version] = rand.Intn(2) == 1
	}
	assessment.CVSSVersions = cvssVersions

	imageDataDecoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		t.Errorf("base64.StdEncoding.DecodeString() = %v, want %v", err, nil)
	}

	var vulnerabilities []model.Vulnerability
	for i := 0; i < 5; i++ {
		vulnerability := model.Vulnerability{
			Model:         model.Model{ID: uuid.New()},
			Category:      model.Category{Name: randName(3)},
			DetailedTitle: randName(3),
			Status:        "Open",
			References:    []string{randUrl(), randUrl()},
			GenericDescription: model.VulnerabilityGeneric{
				Text: randName(20),
			},
			GenericRemediation: model.VulnerabilityGeneric{
				Text: randName(20),
			},
			Description: randName(20),
			Remediation: randName(10),
			Target:      assessment.Targets[rand.Intn(len(assessment.Targets))],
		}
		poc := model.Poc{VulnerabilityID: vulnerability.ID}

		for _, version := range cvss.CvssVersions {
			cvssVector := randCVSSVector(version)
			vector, err := cvss.ParseVector(cvssVector, version, assessment.Language)
			if err != nil {
				t.Errorf("ParseVector() = %v, want %v, cvss version %s", err, nil, version)
			}

			switch version {
			case cvss.Cvss2:
				vulnerability.CVSSv2 = *vector
			case cvss.Cvss3:
				vulnerability.CVSSv3 = *vector
			case cvss.Cvss31:
				vulnerability.CVSSv31 = *vector
			case cvss.Cvss4:
				vulnerability.CVSSv4 = *vector
			}
		}

		for j := 0; j < 3; j++ {
			poc.Pocs = append(poc.Pocs, model.PocItem{
				Index:       j + 1,
				Type:        "request",
				Description: randName(10),
				URI:         fmt.Sprintf("https://%s", vulnerability.Target.FQDN),
				Request:     randName(20),
				Response:    randName(20),
			})
		}
		poc.Pocs = append(poc.Pocs, model.PocItem{
			Index:        4,
			Type:         "image",
			Description:  randName(10),
			URI:          fmt.Sprintf("https://%s", vulnerability.Target.FQDN),
			ImageData:    imageDataDecoded,
			ImageCaption: "Caption" + randName(2),
		})

		poc.Pocs = append(poc.Pocs, model.PocItem{
			Index:        5,
			Type:         "text",
			Description:  randName(10),
			URI:          fmt.Sprintf("https://%s", vulnerability.Target.FQDN),
			TextLanguage: "JavaScript",
			TextData:     randName(20),
		})

		vulnerability.Poc = poc

		vulnerabilities = append(vulnerabilities, vulnerability)
	}

	assessment.VulnerabilityCount = len(vulnerabilities)

	reportData := &reportdata.ReportData{
		Customer:        customer,
		Assessment:      assessment,
		Vulnerabilities: vulnerabilities,
	}

	options := &reportdata.Options{
		FormatJson: true,
	}

	report, _ := NewZipDefaultTemplate()

	t.Run("test", func(t *testing.T) {
		data, err := report.Render(reportData, options)
		if err != nil {
			t.Errorf("Render() = %v, want %v, cvss versions %v", err, true, assessment.CVSSVersions)
		}

		err = os.WriteFile("report.zip", data, 0644)
		if err != nil {
			t.Errorf("os.WriteFile() = %v, want %v", err, nil)
		}
	})
}
