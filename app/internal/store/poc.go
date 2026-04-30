package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type PocStore interface {
	Upsert(ctx context.Context, poc *model.Poc) error

	GetByID(ctx context.Context, ID uuid.UUID) (*model.Poc, error)
	GetByVulnerabilityID(ctx context.Context, vulnerabilityID uuid.UUID) (*model.Poc, error)
	GetByImageID(ctx context.Context, imageID uuid.UUID) ([]model.Poc, error)

	Clone(ctx context.Context, pocID, vulnerabilityID uuid.UUID) (uuid.UUID, error)

	// Only safe for freshly-inserted vulnerabilities; skips the existence check Upsert performs.
	BulkInsertNew(ctx context.Context, pocs []model.Poc) error
}
