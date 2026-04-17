package templates

import (
	"fmt"

	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
)

type XlsxTemplate struct {
	TemplateBytes []byte
	filename      string
	extension     string
}

func NewXlsxTemplate(templateBytes []byte) (*XlsxTemplate, error) {
	if templateBytes == nil {
		return nil, ErrTemplateByteRequired
	}

	return &XlsxTemplate{
		TemplateBytes: templateBytes,
		extension:     "xlsx",
	}, nil
}

func (t *XlsxTemplate) Render(reportData *reportdata.ReportData, options *reportdata.Options) ([]byte, error) {
	t.filename = fmt.Sprintf("%s - %s - %s", reportData.Assessment.Type.Short, reportData.Customer.Name, reportData.Assessment.Name)

	return []byte{}, nil
}

func (t *XlsxTemplate) Filename() string {
	return fmt.Sprintf("%s.%s", t.filename, t.extension)
}

func (t *XlsxTemplate) Extension() string {
	return t.extension
}
