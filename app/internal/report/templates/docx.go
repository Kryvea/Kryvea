package templates

import (
	"errors"
	"fmt"
	"text/template"

	gotemplatedocx "github.com/JJJJJJack/go-template-docx"
	"github.com/Kryvea/Kryvea/internal/model"
	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
	"github.com/google/uuid"
)

var ErrTemplateByteRequired error = errors.New("template required")

// forEachUniquePocImage calls fn once for each distinct image referenced by
// the vulnerabilities' PoC items.
func forEachUniquePocImage(vulnerabilities []model.Vulnerability, fn func(reference string, data []byte) error) error {
	addedImages := make(map[string]bool)
	for _, vulnerability := range vulnerabilities {
		for _, pocItem := range vulnerability.Poc.Pocs {
			if pocItem.Type != model.PocTypeImage || addedImages[pocItem.ImageReference] {
				continue
			}

			if err := fn(pocItem.ImageReference, pocItem.ImageData); err != nil {
				return err
			}
			addedImages[pocItem.ImageReference] = true
		}
	}

	return nil
}

type DocxTemplate struct {
	TemplateBytes []byte
	filename      string
	extension     string
}

func NewDocxTemplate(templateBytes []byte) (*DocxTemplate, error) {
	if templateBytes == nil {
		return nil, ErrTemplateByteRequired
	}

	return &DocxTemplate{
		TemplateBytes: templateBytes,
		extension:     "docx",
	}, nil
}

func (t *DocxTemplate) Render(reportData *reportdata.ReportData, options *reportdata.Options) ([]byte, error) {
	t.filename = fmt.Sprintf("%s - %s - %s", reportData.Assessment.Type.Short, reportData.Customer.Name, reportData.Assessment.Name)

	reportData.Prepare(options.SortByCvss)

	DocxTemplate, err := gotemplatedocx.NewDocxTemplateFromBytes(t.TemplateBytes)
	if err != nil {
		return nil, err
	}

	forEachUniquePocImage(reportData.Vulnerabilities, func(reference string, data []byte) error {
		DocxTemplate.Media(reference, data)
		return nil
	})

	if reportData.Customer.LogoID != uuid.Nil && reportData.Customer.LogoReference != "" {
		DocxTemplate.Media(reportData.Customer.LogoReference, reportData.Customer.LogoData)
	}

	DocxTemplate.AddTemplateFuncs(template.FuncMap{
		"formatDate":           reportdata.FormatDate,
		"getOWASPColor":        reportdata.GetOWASPColor,
		"tableSeverityColor":   reportdata.TableSeverityColor,
		"tableComplexityColor": reportdata.TableComplexityColor,
		"shadeTextBg":          reportdata.ShadeTextBg,
		"debug":                reportdata.Debug,
		"vulnIndex":            reportdata.MakeVulnIndexFunc(reportData.Vulnerabilities),
	})

	err = DocxTemplate.Apply(reportData)
	if err != nil {
		return nil, err
	}

	return DocxTemplate.Bytes(), nil
}

func (t *DocxTemplate) Filename() string {
	return fmt.Sprintf("%s.%s", t.filename, t.extension)
}
