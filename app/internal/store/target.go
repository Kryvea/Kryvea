package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type TargetStore interface {
	Insert(ctx context.Context, target *model.Target, customerID uuid.UUID) (uuid.UUID, error)
	FirstOrInsert(ctx context.Context, target *model.Target, customerID uuid.UUID) (uuid.UUID, bool, error)

	Update(ctx context.Context, targetID uuid.UUID, target *model.Target) error
	Delete(ctx context.Context, targetID uuid.UUID) error

	GetByIDWithRelations(ctx context.Context, targetID uuid.UUID) (*model.Target, error)

	Search(ctx context.Context, customerID uuid.UUID, query string) ([]model.Target, error)
}
