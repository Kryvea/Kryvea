package model

import "github.com/google/uuid"

type Customer struct {
	Model
	Name          string     `json:"name"`
	Language      string     `json:"language"`
	LogoID        uuid.UUID  `json:"logo_id"`
	LogoMimeType  string     `json:"-"`
	LogoReference string     `json:"logo_reference"`
	Templates     []Template `json:"templates"`

	LogoData []byte `json:"-"`
}
