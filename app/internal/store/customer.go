package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type CustomerStore interface {
	Insert(ctx context.Context, customer *model.Customer) (uuid.UUID, error)

	Update(ctx context.Context, customerID uuid.UUID, customer *model.Customer) error
	UpdateLogo(ctx context.Context, customerID, logoID uuid.UUID, mime string) error

	Delete(ctx context.Context, customerID uuid.UUID) error

	GetByID(ctx context.Context, customerID uuid.UUID) (*model.Customer, error)
	GetByIDWithRelations(ctx context.Context, customerID uuid.UUID) (*model.Customer, error)
	GetAll(ctx context.Context, customerIDs []uuid.UUID) ([]model.Customer, error)
}
