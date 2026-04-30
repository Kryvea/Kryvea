package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type FileReferenceStore interface {
	Insert(ctx context.Context, data []byte) (uuid.UUID, string, error)

	GetByID(ctx context.Context, id uuid.UUID) (*model.FileReference, error)
	GetByChecksum(ctx context.Context, checksum [16]byte) (*model.FileReference, error)

	ReadByID(ctx context.Context, id uuid.UUID) ([]byte, *model.FileReference, error)

	GCFiles(ctx context.Context) (int, error)
}
