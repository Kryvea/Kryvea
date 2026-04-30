package model

import "github.com/google/uuid"

type Poc struct {
	Model
	Pocs            []PocItem `json:"pocs"`
	VulnerabilityID uuid.UUID `json:"vulnerability_id"`
}

type PocItem struct {
	ID                  uuid.UUID         `json:"id"`
	Index               int               `json:"index"`
	Type                string            `json:"type"`
	Description         string            `json:"description"`
	URI                 string            `json:"uri,omitempty"`
	Request             string            `json:"request,omitempty"`
	RequestHighlights   []HighlightedText `json:"request_highlights,omitempty"`
	RequestHighlighted  []Highlighted     `json:"request_highlighted,omitempty"`
	Response            string            `json:"response,omitempty"`
	ResponseHighlights  []HighlightedText `json:"response_highlights,omitempty"`
	ResponseHighlighted []Highlighted     `json:"response_highlighted,omitempty"`
	ImageID             uuid.UUID         `json:"image_id,omitempty"`
	ImageReference      string            `json:"image_reference,omitempty"`
	ImageFilename       string            `json:"image_filename,omitempty"`
	ImageMimeType       string            `json:"-"`
	ImageCaption        string            `json:"image_caption,omitempty"`
	TextLanguage        string            `json:"text_language,omitempty"`
	TextData            string            `json:"text_data,omitempty"`
	TextHighlights      []HighlightedText `json:"text_highlights,omitempty"`
	TextHighlighted     []Highlighted     `json:"text_highlighted,omitempty"`
	StartingLineNumber  int               `json:"starting_line_number,omitempty"`
	// Only populated on report generation
	ImageData []byte `json:"-"`
}

type HighlightedText struct {
	Start           LineCol `json:"start"`
	End             LineCol `json:"end"`
	SelectedPreview string  `json:"selectionPreview"`
	Color           string  `json:"color"`
}

type LineCol struct {
	Line int `json:"line"`
	Col  int `json:"col"`
}

type Highlighted struct {
	Text  string `json:"text,omitempty"`
	Color string `json:"color,omitempty"`
}
