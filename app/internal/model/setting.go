package model

import "github.com/google/uuid"

// SettingID is the fixed UUID of the singleton settings row.
var SettingID uuid.UUID = [16]byte{
	'K', 'R', 'Y', 'V',
	'E', 'A', '-', 'S',
	'E', 'T', 'T', 'I',
	'N', 'G', 'I', 'D',
}

type Setting struct {
	Model
	MaxImageSize            int64  `json:"-"`
	DefaultCategoryLanguage string `json:"default_category_language"`
}
