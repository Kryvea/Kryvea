package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
)

type SettingStore interface {
	Get(ctx context.Context) (*model.Setting, error)
	Update(ctx context.Context, setting *model.Setting) error
	ValidateImageSize(ctx context.Context, size int64) error
}
