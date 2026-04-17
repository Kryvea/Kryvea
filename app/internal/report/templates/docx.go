package templates

import (
	"fmt"
	"text/template"

	"github.com/Kryvea/Kryvea/internal/poc"
	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
	gotemplatedocx "github.com/JJJJJJack/go-template-docx"
	"github.com/google/uuid"
)

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

	reportData.Prepare()

	DocxTemplate, err := gotemplatedocx.NewDocxTemplateFromBytes(t.TemplateBytes)
	if err != nil {
		return nil, err
	}

	addedImages := make(map[string]bool)
	for _, vulnerability := range reportData.Vulnerabilities {
		for _, pocItem := range vulnerability.Poc.Pocs {
			if pocItem.Type == poc.PocTypeImage {
				if _, ok := addedImages[pocItem.ImageReference]; ok {
					continue
				}

				DocxTemplate.Media(pocItem.ImageReference, pocItem.ImageData)
				addedImages[pocItem.ImageReference] = true
			}
		}
	}

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

func (t *DocxTemplate) Extension() string {
	return t.extension
}
