package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type CategoryStore interface {
	Insert(ctx context.Context, category *model.Category) (uuid.UUID, error)
	Upsert(ctx context.Context, category *model.Category, override bool) (uuid.UUID, error)
	FirstOrInsert(ctx context.Context, category *model.Category) (uuid.UUID, bool, error)

	Update(ctx context.Context, ID uuid.UUID, category *model.Category) error
	Delete(ctx context.Context, ID uuid.UUID) error

	GetAll(ctx context.Context) ([]model.Category, error)
	GetByID(ctx context.Context, categoryID uuid.UUID) (*model.Category, error)

	Search(ctx context.Context, query string) ([]model.Category, error)
}
