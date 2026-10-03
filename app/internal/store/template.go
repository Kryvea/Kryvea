package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type TemplateStore interface {
	Insert(ctx context.Context, template *model.Template) (uuid.UUID, error)

	GetByID(ctx context.Context, id uuid.UUID) (*model.Template, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID) (*model.Template, error)
	GetAll(ctx context.Context) ([]model.Template, error)

	Delete(ctx context.Context, id uuid.UUID) error
}
