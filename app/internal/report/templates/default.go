package templates

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/model"
	pocpkg "github.com/Kryvea/Kryvea/internal/poc"
	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/Kryvea/Kryvea/internal/zip"
)

type ReportDataJSON struct {
	Customer         *model.Customer       `json:"customer"`
	Assessment       *model.Assessment     `json:"assessment"`
	Vulnerabilities  []model.Vulnerability `json:"vulnerabilities"`
	DeliveryDateTime time.Time
	MaxCVSS          map[string]cvss.Vector
}

func (t *ZipDefaultTemplate) renderReport(reportData *reportdata.ReportData, options *reportdata.Options) ([]byte, error) {
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	defer zipWriter.Close()

	mediaDir := "media"
	zipWriter.AddDirectory(mediaDir)

	addedImages := make(map[string]bool)
	for _, vuln := range reportData.Vulnerabilities {
		for _, pocItem := range vuln.Poc.Pocs {
			if pocItem.Type == pocpkg.PocTypeImage {
				if _, ok := addedImages[pocItem.ImageReference]; ok {
					continue
				}

				imagePath := fmt.Sprintf("%s/%s", mediaDir, pocItem.ImageReference)

				// add image to the zip
				zipWriter.AddFile(bytes.NewBuffer(pocItem.ImageData), imagePath)

				addedImages[pocItem.ImageReference] = true

			}
		}
	}

	data := &ReportDataJSON{
		Customer:         reportData.Customer,
		Assessment:       reportData.Assessment,
		Vulnerabilities:  reportData.Vulnerabilities,
		DeliveryDateTime: reportData.DeliveryDateTime,
		MaxCVSS:          reportData.MaxCVSS,
	}

	b, err := util.MarshalJson(data, options.FormatJson)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)

	baseFileName := util.SanitizeFileName(t.filename) + ".json"

	zipWriter.AddFile(r, baseFileName)

	zipWriter.Close()

	return zipBuf.Bytes(), nil
}

type ZipDefaultTemplate struct {
	filename  string
	extension string
}

func NewZipDefaultTemplate() (*ZipDefaultTemplate, error) {
	return &ZipDefaultTemplate{
		extension: "zip",
	}, nil
}

func (t *ZipDefaultTemplate) Render(reportData *reportdata.ReportData, options *reportdata.Options) ([]byte, error) {
	t.filename = fmt.Sprintf("%s - %s - %s", reportData.Assessment.Type.Short, reportData.Customer.Name, reportData.Assessment.Name)

	reportData.MaxCVSS = reportdata.GetMaxCvss(reportData.Vulnerabilities, reportData.Assessment.CVSSVersions)

	// Sort vulnerabilities by score. if score is equal, sort by name in ascending order
	switch options.SortByCvss {
	case cvss.Cvss2:
		sort.Slice(reportData.Vulnerabilities, func(i, j int) bool {
			if reportData.Vulnerabilities[i].CVSSv2.Score == reportData.Vulnerabilities[j].CVSSv2.Score {
				return reportData.Vulnerabilities[i].DetailedTitle < reportData.Vulnerabilities[j].DetailedTitle
			}
			return reportData.Vulnerabilities[i].CVSSv2.Score > reportData.Vulnerabilities[j].CVSSv2.Score
		})
	case cvss.Cvss3:
		sort.Slice(reportData.Vulnerabilities, func(i, j int) bool {
			if reportData.Vulnerabilities[i].CVSSv3.Score == reportData.Vulnerabilities[j].CVSSv3.Score {
				return reportData.Vulnerabilities[i].DetailedTitle < reportData.Vulnerabilities[j].DetailedTitle
			}
			return reportData.Vulnerabilities[i].CVSSv3.Score > reportData.Vulnerabilities[j].CVSSv3.Score
		})
	case cvss.Cvss31:
		sort.Slice(reportData.Vulnerabilities, func(i, j int) bool {
			if reportData.Vulnerabilities[i].CVSSv31.Score == reportData.Vulnerabilities[j].CVSSv31.Score {
				return reportData.Vulnerabilities[i].DetailedTitle < reportData.Vulnerabilities[j].DetailedTitle
			}
			return reportData.Vulnerabilities[i].CVSSv31.Score > reportData.Vulnerabilities[j].CVSSv31.Score
		})
	case cvss.Cvss4:
		sort.Slice(reportData.Vulnerabilities, func(i, j int) bool {
			if reportData.Vulnerabilities[i].CVSSv4.Score == reportData.Vulnerabilities[j].CVSSv4.Score {
				return reportData.Vulnerabilities[i].DetailedTitle < reportData.Vulnerabilities[j].DetailedTitle
			}
			return reportData.Vulnerabilities[i].CVSSv4.Score > reportData.Vulnerabilities[j].CVSSv4.Score
		})
	}

	// Sort poc.Pocs for each poc in pocs
	for i := range reportData.Vulnerabilities {

		sort.Slice(reportData.Vulnerabilities[i].Poc.Pocs, func(j, k int) bool {
			return reportData.Vulnerabilities[i].Poc.Pocs[j].Index < reportData.Vulnerabilities[i].Poc.Pocs[k].Index
		})

	}

	return t.renderReport(reportData, options)
}

func (t *ZipDefaultTemplate) Filename() string {
	return fmt.Sprintf("%s.%s", t.filename, t.extension)
}

func (t *ZipDefaultTemplate) Extension() string {
	return t.extension
}
