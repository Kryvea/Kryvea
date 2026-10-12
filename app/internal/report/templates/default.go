package templates

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/model"
	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/Kryvea/Kryvea/internal/zip"
	"github.com/bytedance/sonic"
)

type ReportDataJSON struct {
	Customer         *model.Customer        `json:"customer"`
	Assessment       *model.Assessment      `json:"assessment"`
	Vulnerabilities  []model.Vulnerability  `json:"vulnerabilities"`
	DeliveryDateTime time.Time              `json:"delivery_date_time"`
	MaxCVSS          map[string]cvss.Vector `json:"max_cvss"`
}

func (t *ZipDefaultTemplate) renderReport(reportData *reportdata.ReportData, options *reportdata.Options) ([]byte, error) {
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	mediaDir := "media"
	if err := zipWriter.AddDirectory(mediaDir); err != nil {
		return nil, err
	}

	err := forEachUniquePocImage(reportData.Vulnerabilities, func(reference string, data []byte) error {
		imagePath := fmt.Sprintf("%s/%s", mediaDir, reference)
		return zipWriter.AddFile(bytes.NewReader(data), imagePath)
	})
	if err != nil {
		return nil, err
	}

	data := &ReportDataJSON{
		Customer:         reportData.Customer,
		Assessment:       reportData.Assessment,
		Vulnerabilities:  reportData.Vulnerabilities,
		DeliveryDateTime: reportData.DeliveryDateTime,
		MaxCVSS:          reportData.MaxCVSS,
	}

	var b []byte
	if options.FormatJson {
		b, err = sonic.MarshalIndent(data, "", "\t")
	} else {
		b, err = sonic.Marshal(data)
	}
	if err != nil {
		return nil, err
	}

	baseFileName := util.SanitizeFileName(t.filename) + ".json"

	if err := zipWriter.AddFile(bytes.NewReader(b), baseFileName); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return zipBuf.Bytes(), nil
}

type ZipDefaultTemplate struct {
	filename  string
	extension string
}

func NewZipDefaultTemplate() *ZipDefaultTemplate {
	return &ZipDefaultTemplate{
		extension: "zip",
	}
}

func (t *ZipDefaultTemplate) Render(reportData *reportdata.ReportData, options *reportdata.Options) ([]byte, error) {
	t.filename = fmt.Sprintf("%s - %s - %s", reportData.Assessment.Type.Short, reportData.Customer.Name, reportData.Assessment.Name)

	reportData.PrepareJSON(options.SortByCvss)

	return t.renderReport(reportData, options)
}

func (t *ZipDefaultTemplate) Filename() string {
	return fmt.Sprintf("%s.%s", t.filename, t.extension)
}
