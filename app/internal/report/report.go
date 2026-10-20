package report

import (
	"errors"

	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
	"github.com/Kryvea/Kryvea/internal/report/templates"
)

const (
	ReportTemplateXlsx string = "xlsx"
	ReportTemplateDocx string = "docx"

	ReportZipDefault string = "zip-default"
)

var (
	ErrTemplateTypeNA error = errors.New("template type not available")

	ReportTemplateMap map[string]struct{} = map[string]struct{}{
		ReportTemplateXlsx: {},
		ReportTemplateDocx: {},
	}
)

type Report interface {
	Render(reportData *reportdata.ReportData, options *reportdata.Options) ([]byte, error)
	Filename() string
}

func New(reportType string, templateBytes []byte) (Report, error) {
	switch reportType {
	case ReportTemplateXlsx:
		return templates.NewXlsxTemplate(templateBytes)
	case ReportTemplateDocx:
		return templates.NewDocxTemplate(templateBytes)
	case ReportZipDefault:
		return templates.NewZipDefaultTemplate(), nil
	}

	return nil, ErrTemplateTypeNA
}
