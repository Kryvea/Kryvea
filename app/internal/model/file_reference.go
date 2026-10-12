package model

import "github.com/google/uuid"

type FileReference struct {
	Model
	Checksum [16]byte    `json:"checksum"`
	MimeType string      `json:"mime_type"`
	UsedBy   []uuid.UUID `json:"used_by"`
}
