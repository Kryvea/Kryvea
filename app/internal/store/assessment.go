package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type AssessmentStore interface {
	Insert(ctx context.Context, assessment *model.Assessment, customerID uuid.UUID) (uuid.UUID, error)

	GetByID(ctx context.Context, assessmentID uuid.UUID) (*model.Assessment, error)
	GetByIDWithRelations(ctx context.Context, assessmentID uuid.UUID) (*model.Assessment, error)
	GetMultipleByID(ctx context.Context, assessmentIDs []uuid.UUID) ([]model.Assessment, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]model.Assessment, error)

	Search(ctx context.Context, customers []uuid.UUID, customerID uuid.UUID, name string) ([]model.Assessment, error)

	Update(ctx context.Context, assessmentID uuid.UUID, assessment *model.Assessment) error
	UpdateStatus(ctx context.Context, assessmentID uuid.UUID, assessment *model.Assessment) error
	UpdateTargets(ctx context.Context, assessmentID uuid.UUID, target uuid.UUID) error

	Delete(ctx context.Context, assessmentID uuid.UUID) error
	Clone(ctx context.Context, assessmentID uuid.UUID, assessmentName string, includePocs bool) (uuid.UUID, error)

	BulkUpdateTargets(ctx context.Context, assessmentID uuid.UUID, targetIDs []uuid.UUID) error
}
