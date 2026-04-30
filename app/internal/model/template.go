package model

import "github.com/google/uuid"

var (
	TemplateTypeXlsx           = "xlsx"
	TemplateTypeDocx           = "docx"
	TemplateTypeZip            = "generic-zip"
	XlsxMimeType               = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	DocxMimeType               = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ZipMimeType                = "application/zip"
	SupportedTemplateMimeTypes = map[string]string{
		XlsxMimeType: TemplateTypeXlsx,
		DocxMimeType: TemplateTypeDocx,
		ZipMimeType:  TemplateTypeZip,
	}
)

type Template struct {
	Model
	Name         string    `json:"name"`
	Filename     string    `json:"filename,omitempty"`
	Language     string    `json:"language,omitempty"`
	TemplateType string    `json:"template_type"`
	MimeType     string    `json:"-"`
	Identifier   string    `json:"identifier,omitempty"`
	FileID       uuid.UUID `json:"file_id,omitempty"`
	Customer     *Customer `json:"customer,omitempty"`
}
