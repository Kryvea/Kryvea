// Package model contains the data types shared by every layer of the
// application. It is driver-agnostic: types here have no DB-specific code.
package model

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
